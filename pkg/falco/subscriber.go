package falco

import (
	"context"
	"fmt"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/falcosecurity/client-go/pkg/client"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/gcp"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Subscriber subscribes to Falco outputs via gRPC
type Subscriber struct {
	cfg        config.FalcoConfig
	client     *client.Client
	grpcConn   *grpc.ClientConn
	isInsecure bool
	gcpParser  *gcp.AuditParser // GCP Audit Log parser
}

// NewSubscriber creates a new Falco subscriber
func NewSubscriber(cfg config.FalcoConfig) (*Subscriber, error) {
	return &Subscriber{
		cfg:       cfg,
		gcpParser: gcp.NewAuditParser(),
	}, nil
}

// Start starts subscribing to Falco outputs
func (s *Subscriber) Start(ctx context.Context, eventCh chan<- types.Event) error {
	log.Info("Starting Falco subscriber...")

	// Check if TLS certificates are configured
	if s.cfg.CertFile != "" && s.cfg.KeyFile != "" {
		// Use TLS connection with certificates via client-go library
		log.Infof("Using TLS connection to Falco at %s:%d", s.cfg.Hostname, s.cfg.Port)
		return s.startWithTLS(ctx, eventCh)
	} else {
		// Use insecure plaintext connection (direct gRPC)
		log.Infof("Using insecure plaintext connection to Falco at %s:%d", s.cfg.Hostname, s.cfg.Port)
		return s.startWithInsecure(ctx, eventCh)
	}
}

// startWithTLS creates a TLS connection using the client-go library
func (s *Subscriber) startWithTLS(ctx context.Context, eventCh chan<- types.Event) error {
	clientConfig := &client.Config{
		Hostname:   s.cfg.Hostname,
		Port:       s.cfg.Port,
		CertFile:   s.cfg.CertFile,
		KeyFile:    s.cfg.KeyFile,
		CARootFile: s.cfg.CARootFile,
	}
	c, err := client.NewForConfig(ctx, clientConfig)
	if err != nil {
		return fmt.Errorf("failed to create Falco client: %w", err)
	}
	s.client = c
	defer func() {
		if closeErr := c.Close(); closeErr != nil {
			log.Warnf("Failed to close Falco client: %v", closeErr)
		}
	}()

	log.Infof("Connected to Falco at %s:%d", s.cfg.Hostname, s.cfg.Port)

	// Subscribe to outputs stream
	outputClient, err := c.Outputs()
	if err != nil {
		return fmt.Errorf("failed to get outputs client: %w", err)
	}

	// Start streaming using Sub method
	stream, err := outputClient.Sub(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe to Falco outputs: %w", err)
	}

	return s.receiveMessages(ctx, stream, eventCh)
}

// startWithInsecure creates a plaintext gRPC connection directly
func (s *Subscriber) startWithInsecure(ctx context.Context, eventCh chan<- types.Event) error {
	// Create direct gRPC connection with insecure credentials
	target := fmt.Sprintf("%s:%d", s.cfg.Hostname, s.cfg.Port)
	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to Falco: %w", err)
	}
	s.grpcConn = conn
	s.isInsecure = true
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Warnf("Failed to close gRPC connection: %v", closeErr)
		}
	}()

	log.Infof("Connected to Falco at %s", target)

	// Create outputs service client directly from gRPC connection
	outputClient := outputs.NewServiceClient(conn)

	// Start streaming using Sub method
	stream, err := outputClient.Sub(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe to Falco outputs: %w", err)
	}

	return s.receiveMessages(ctx, stream, eventCh)
}

// receiveMessages receives and processes messages from the Falco stream
func (s *Subscriber) receiveMessages(ctx context.Context, stream outputs.Service_SubClient, eventCh chan<- types.Event) error {
	// Receive messages from stream
	for {
		select {
		case <-ctx.Done():
			log.Info("Falco subscriber stopping...")
			return ctx.Err()
		default:
			res, err := stream.Recv()
			if err != nil {
				return fmt.Errorf("error receiving Falco output: %w", err)
			}

			// Parse Falco output
			event := s.parseFalcoOutput(res)
			if event != nil {
				select {
				case eventCh <- *event:
					log.Debugf("Sent Falco event to channel: %s", res.Rule)
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
}
