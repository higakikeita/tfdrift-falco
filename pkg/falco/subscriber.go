package falco

import (
	"context"
	"fmt"

	"github.com/falcosecurity/client-go/pkg/client"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Subscriber subscribes to Falco outputs via gRPC
type Subscriber struct {
	cfg    config.FalcoConfig
	client *client.Client
}

// NewSubscriber creates a new Falco subscriber
func NewSubscriber(cfg config.FalcoConfig) (*Subscriber, error) {
	return &Subscriber{
		cfg: cfg,
	}, nil
}

// Start starts subscribing to Falco outputs
func (s *Subscriber) Start(ctx context.Context, eventCh chan<- types.Event) error {
	log.Info("Starting Falco subscriber...")

	// Create Falco client configuration
	var c *client.Client
	var err error

	// Check if TLS certificates are configured
	if s.cfg.CertFile != "" && s.cfg.KeyFile != "" {
		// Use TLS connection with certificates
		log.Infof("Using TLS connection to Falco at %s:%d", s.cfg.Hostname, s.cfg.Port)
		clientConfig := &client.Config{
			Hostname:   s.cfg.Hostname,
			Port:       s.cfg.Port,
			CertFile:   s.cfg.CertFile,
			KeyFile:    s.cfg.KeyFile,
			CARootFile: s.cfg.CARootFile,
		}
		c, err = client.NewForConfig(ctx, clientConfig)
	} else {
		// Use insecure plaintext connection (no TLS)
		log.Infof("Using insecure plaintext connection to Falco at %s:%d", s.cfg.Hostname, s.cfg.Port)
		c, err = client.NewForConfig(ctx, &client.Config{
			Hostname:                  s.cfg.Hostname,
			Port:                      s.cfg.Port,
			InsecureSkipMutualTLSAuth: true,
			DialOptions: []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			},
		})
	}

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
