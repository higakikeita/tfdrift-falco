// Package main provides the TFDrift-Falco CLI for real-time Terraform drift detection.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/detector"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	version     = "0.3.0"
	cfgFile     string
	autoDetect  bool
	dryRun      bool
	daemon      bool
	interactive bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tfdrift",
		Short: "Real-time Terraform drift detection powered by Falco",
		Long: `TFDrift-Falco detects manual (non-IaC) changes in your cloud environment
in real-time by combining Falco runtime security monitoring, CloudTrail events,
and Terraform state comparison.`,
		Version: version,
		Run:     run,
	}

	rootCmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is config.yaml)")
	rootCmd.Flags().BoolVar(&autoDetect, "auto", false, "auto-detect Terraform state from current directory")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "run in dry-run mode (no notifications)")
	rootCmd.Flags().BoolVar(&daemon, "daemon", false, "run in daemon mode")
	rootCmd.Flags().BoolVar(&interactive, "interactive", false, "run in interactive mode (for approval prompts)")

	// Add approval subcommands
	rootCmd.AddCommand(newApprovalCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

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

	// Start detection
	log.Info("Drift detection started")
	if err := det.Start(ctx); err != nil {
		log.Fatalf("Detector error: %v", err)
	}

	log.Info("TFDrift-Falco stopped")
}

func loadAutoConfig() (*config.Config, error) {
	// Get current working directory
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	log.Infof("Searching for Terraform state in: %s", workDir)

	// Auto-detect Terraform state
	result, err := config.AutoDetectTerraformState(workDir)
	if err != nil {
		return nil, fmt.Errorf("auto-detection error: %w", err)
	}

	// Print detection results
	config.PrintAutoDetectHelp(result)

	// Check if state was found
	if !result.Found {
		return nil, fmt.Errorf("no Terraform state detected")
	}

	// Create configuration from auto-detection results
	cfg, err := config.CreateAutoConfig(result)
	if err != nil {
		return nil, fmt.Errorf("failed to create config from auto-detection: %w", err)
	}

	log.Info("✓ Auto-detection successful, starting drift detection...")
	return cfg, nil
}

func initLogger() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

// newApprovalCmd creates the approval subcommand
func newApprovalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approval",
		Short: "Manage import approval requests",
		Long:  "Manage terraform import approval requests for unmanaged resources",
	}

	cmd.AddCommand(newApprovalListCmd())
	cmd.AddCommand(newApprovalApproveCmd())
	cmd.AddCommand(newApprovalRejectCmd())
	cmd.AddCommand(newApprovalCleanupCmd())

	return cmd
}

// newApprovalListCmd lists pending approval requests
func newApprovalListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List pending approval requests",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("⚠️  This feature requires a running TFDrift-Falco instance with persistent state.")
			fmt.Println("Currently, approval requests are only available during interactive sessions.")
			fmt.Println("\nTo use approval workflow:")
			fmt.Println("  1. Enable auto_import in config.yaml")
			fmt.Println("  2. Set require_approval: true")
			fmt.Println("  3. Run: tfdrift --config config.yaml --interactive")
		},
	}
}

// newApprovalApproveCmd approves a specific request
func newApprovalApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve [request-id]",
		Short: "Approve a specific import request",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			requestID := args[0]
			fmt.Printf("⚠️  Approving request %s\n", requestID)
			fmt.Println("This feature requires a running TFDrift-Falco instance with persistent state.")
			fmt.Println("\nFor now, use interactive mode:")
			fmt.Println("  tfdrift --config config.yaml --interactive")
		},
	}
}

// newApprovalRejectCmd rejects a specific request
func newApprovalRejectCmd() *cobra.Command {
	var reason string

	cmd := &cobra.Command{
		Use:   "reject [request-id]",
		Short: "Reject a specific import request",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			requestID := args[0]
			fmt.Printf("⚠️  Rejecting request %s\n", requestID)
			if reason != "" {
				fmt.Printf("Reason: %s\n", reason)
			}
			fmt.Println("This feature requires a running TFDrift-Falco instance with persistent state.")
			fmt.Println("\nFor now, use interactive mode:")
			fmt.Println("  tfdrift --config config.yaml --interactive")
		},
	}

	cmd.Flags().StringVar(&reason, "reason", "", "reason for rejection")
	return cmd
}

// newApprovalCleanupCmd cleans up expired requests
func newApprovalCleanupCmd() *cobra.Command {
	var olderThan string

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up expired approval requests",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("⚠️  Cleaning up requests older than %s\n", olderThan)
			fmt.Println("This feature requires a running TFDrift-Falco instance with persistent state.")
			fmt.Println("\nFor now, approval requests are automatically cleaned up during interactive sessions.")
		},
	}

	cmd.Flags().StringVar(&olderThan, "older-than", "24h", "clean up requests older than this duration")
	return cmd
}
