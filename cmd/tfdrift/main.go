// Package main provides the TFDrift-Falco CLI for real-time Terraform drift detection.
package main

import (
	"fmt"
	"os"

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

func initLogger() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}
