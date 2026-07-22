// Command falcosubprobe is a MINIMAL Falco gRPC-outputs subscriber used to
// isolate issue #311: it subscribes to Falco's outputs over mTLS with the same
// client library tfdrift uses (falcosecurity/client-go), and prints every
// response it receives. It shares no code with tfdrift's detector, so it
// answers one question cleanly:
//
//	Does Falco's grpc_output deliver alerts to a client-go subscriber that is
//	attached BEFORE the alert is generated?
//
// If this probe receives alerts but tfdrift does not -> the bug is in tfdrift's
// subscribe/parse path. If neither receives -> the bug is Falco's grpc_output
// (config / version / plugin-source behavior). Attach this probe first, keep
// Falco resident (SQS mode), then inject one event.
//
// Not part of the product; safe to delete once #311 is resolved.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/falcosecurity/client-go/pkg/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	host := flag.String("host", "127.0.0.1", "Falco gRPC host (use 127.0.0.1, not localhost, to avoid IPv6 [::1])")
	port := flag.Uint("port", 5060, "Falco gRPC port")
	cert := flag.String("cert", "", "client cert (mTLS)")
	key := flag.String("key", "", "client key (mTLS)")
	ca := flag.String("ca", "", "CA root cert (mTLS)")
	insec := flag.Bool("insecure", false, "use plaintext gRPC (no mTLS) — to isolate whether mTLS is the blocker")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sig; fmt.Println("\n[probe] shutting down"); cancel() }()

	cfg := &client.Config{
		Hostname:   *host,
		Port:       uint16(*port),
		CertFile:   *cert,
		KeyFile:    *key,
		CARootFile: *ca,
	}

	// Reconnect loop so startup races (Falco gRPC not ready yet) don't kill us.
	backoff := time.Second
	for ctx.Err() == nil {
		if err := run(ctx, cfg, *insec); err != nil && ctx.Err() == nil {
			fmt.Printf("[probe] subscription ended: %v; reconnecting in %s\n", err, backoff)
			select {
			case <-ctx.Done():
			case <-time.After(backoff):
			}
			if backoff < 15*time.Second {
				backoff *= 2
			}
			continue
		}
		return
	}
}

// subFunc abstracts obtaining an outputs Sub stream over either mTLS
// (client-go) or plaintext (raw grpc), plus a cleanup.
func run(ctx context.Context, cfg *client.Config, insec bool) error {
	var stream outputs.Service_SubClient
	var closer func()

	if insec {
		target := fmt.Sprintf("%s:%d", cfg.Hostname, cfg.Port)
		conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return fmt.Errorf("dial (insecure): %w", err)
		}
		closer = func() { _ = conn.Close() }
		s, err := outputs.NewServiceClient(conn).Sub(ctx)
		if err != nil {
			closer()
			return fmt.Errorf("subscribe (insecure): %w", err)
		}
		stream = s
	} else {
		c, err := client.NewForConfig(ctx, cfg)
		if err != nil {
			return fmt.Errorf("new client: %w", err)
		}
		closer = func() { _ = c.Close() }
		oc, err := c.Outputs()
		if err != nil {
			closer()
			return fmt.Errorf("outputs client: %w", err)
		}
		s, err := oc.Sub(ctx)
		if err != nil {
			closer()
			return fmt.Errorf("subscribe: %w", err)
		}
		stream = s
	}
	defer closer()
	fmt.Printf("[probe] SUBSCRIBED (insecure=%v) to Falco outputs at %s:%d — waiting for alerts...\n", insec, cfg.Hostname, cfg.Port)

	var n int
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("recv (received %d so far): %w", n, err)
		}
		n++
		f := res.GetOutputFields()
		fmt.Printf("[probe] #%d rule=%q source=%q priority=%v\n         ct.name=%q ct.user=%q ct.user.arn=%q ct.request.groupid=%q\n         output=%s\n",
			n, res.GetRule(), res.GetSource(), res.GetPriority(),
			f["ct.name"], f["ct.user"], f["ct.user.arn"], f["ct.request.groupid"],
			truncate(res.GetOutput(), 200))
	}
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}
