// Package main provides the TFDrift-Falco CLI for real-time Terraform drift detection.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/keitahigaki/tfdrift-falco/pkg/app"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	version     = "0.4.1"
	cfgFile     string
	autoDetect  bool
	outputMode  string
	dryRun      bool
	daemon      bool
	interactive bool
	// API server flags
	serverMode bool
	apiPort    int
	// L1 Semi-auto mode flags
	regionOverride      []string
	falcoEndpoint       string
	statePathOverride   string
	backendTypeOverride string
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
	rootCmd.Flags().StringVar(&outputMode, "output", "human", "output mode: human, json, or both (e.g., --output json)")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "run in dry-run mode (no notifications)")
	rootCmd.Flags().BoolVar(&daemon, "daemon", false, "run in daemon mode")
	rootCmd.Flags().BoolVar(&interactive, "interactive", false, "run in interactive mode (for approval prompts)")

	// API server flags
	rootCmd.Flags().BoolVar(&serverMode, "server", false, "run API server mode")
	rootCmd.Flags().IntVar(&apiPort, "api-port", 8080, "API server port (default: 8080)")

	// L1 Semi-auto mode flags (used with --auto)
	rootCmd.Flags().StringSliceVar(&regionOverride, "region", nil, "AWS region(s) to monitor (e.g., --region us-west-2,ap-northeast-1)")
	rootCmd.Flags().StringVar(&falcoEndpoint, "falco-endpoint", "", "Falco gRPC endpoint (e.g., --falco-endpoint localhost:5060)")
	rootCmd.Flags().StringVar(&statePathOverride, "state-path", "", "Terraform state file path (e.g., --state-path ./terraform.tfstate)")
	rootCmd.Flags().StringVar(&backendTypeOverride, "backend", "", "Backend type: local or s3 (e.g., --backend local)")

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

	// Create app configuration from command line flags
	appCfg := &app.Config{
		ConfigFile:          cfgFile,
		AutoDetect:          autoDetect,
		OutputMode:          outputMode,
		DryRun:              dryRun,
		Daemon:              daemon,
		Interactive:         interactive,
		ServerMode:          serverMode,
		APIPort:             apiPort,
		RegionOverride:      regionOverride,
		FalcoEndpoint:       falcoEndpoint,
		StatePathOverride:   statePathOverride,
		BackendTypeOverride: backendTypeOverride,
		Version:             version,
	}

	// Create and run the application
	application := app.New(appCfg)
	if err := application.Run(ctx); err != nil {
		log.Fatalf("Application error: %v", err)
	}

	log.Info("TFDrift-Falco stopped")
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
