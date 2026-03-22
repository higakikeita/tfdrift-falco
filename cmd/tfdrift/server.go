// Package main provides server mode functionality for TFDrift-Falco.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/keitahigaki/tfdrift-falco/pkg/api"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/detector"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// run is the main entry point for CLI execution, dispatching to server or detector mode.
func run(_ *cobra.Command, _ []string) {
	// Initialize logger
	initLogger()

	log.Infof("Starting TFDrift-Falco v%s", version)

	var cfg *config.Config
	var err error

	// Load configuration
	if autoDetect {
		log.Info("Auto-detection mode enabled")
		cfg, err = loadAutoConfig()
		if err != nil {
			log.Fatalf("Auto-detection failed: %v", err)
		}
	} else {
		cfg, err = config.Load(cfgFile)
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	}

	if dryRun {
		log.Info("Running in DRY-RUN mode - notifications disabled")
		cfg.DryRun = true
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Infof("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Initialize detector
	det, err := detector.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize detector: %v", err)
	}

	if serverMode {
		runServer(ctx, cfg, det)
	} else {
		runDetector(ctx, det)
	}

	log.Info("TFDrift-Falco stopped")
}

// runServer starts the API server with drift detection in background.
func runServer(ctx context.Context, cfg *config.Config, det *detector.Detector) {
	// API Server mode
	log.Info("Starting TFDrift-Falco in API server mode")
	srv := api.NewServer(cfg, det, apiPort, version)

	// Connect detector to broadcaster for real-time events
	det.SetBroadcaster(srv.GetBroadcaster())

	// Start detector in background
	go func() {
		log.Info("Starting drift detection engine...")
		if err := det.Start(ctx); err != nil {
			log.Errorf("Detector error: %v", err)
		}
	}()

	// Start API server (blocks until shutdown)
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("API server error: %v", err)
	}
}

// runDetector starts drift detection in standard mode.
func runDetector(ctx context.Context, det *detector.Detector) {
	// Standard detection mode
	log.Info("Drift detection started")
	if err := det.Start(ctx); err != nil {
		log.Fatalf("Detector error: %v", err)
	}
}
