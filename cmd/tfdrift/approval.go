// Package main provides approval workflow subcommand functionality.
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newApprovalCmd creates the approval subcommand for managing import approvals.
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

// newApprovalListCmd lists pending approval requests.
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

// newApprovalApproveCmd approves a specific import request.
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

// newApprovalRejectCmd rejects a specific import request.
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

// newApprovalCleanupCmd cleans up expired approval requests.
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
