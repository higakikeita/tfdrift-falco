// Package main provides configuration auto-detection and override functionality.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	log "github.com/sirupsen/logrus"
)

// loadAutoConfig performs auto-detection of Terraform configuration and state.
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

	// Apply L1 Semi-auto mode overrides
	if err := applyConfigOverrides(cfg); err != nil {
		return nil, fmt.Errorf("failed to apply config overrides: %w", err)
	}

	log.Info("✓ Auto-detection successful, starting drift detection...")
	return cfg, nil
}

// applyConfigOverrides applies L1 semi-auto mode flag overrides to the config.
func applyConfigOverrides(cfg *config.Config) error {
	// Override AWS regions if specified
	if len(regionOverride) > 0 {
		cfg.Providers.AWS.Regions = regionOverride
		log.Infof("✓ Using custom region(s): %v", regionOverride)
	}

	// Override Falco endpoint if specified
	if falcoEndpoint != "" {
		parts := strings.Split(falcoEndpoint, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid Falco endpoint format (expected host:port): %s", falcoEndpoint)
		}

		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid port in Falco endpoint: %s", parts[1])
		}

		cfg.Falco.Hostname = parts[0]
		cfg.Falco.Port = uint16(port)
		log.Infof("✓ Using custom Falco endpoint: %s", falcoEndpoint)
	}

	// Override state path if specified
	if statePathOverride != "" {
		cfg.Providers.AWS.State.LocalPath = statePathOverride
		cfg.Providers.AWS.State.Backend = "local"
		log.Infof("✓ Using custom state path: %s", statePathOverride)
	}

	// Override backend type if specified
	if backendTypeOverride != "" {
		if backendTypeOverride != "local" && backendTypeOverride != "s3" {
			return fmt.Errorf("invalid backend type (must be 'local' or 's3'): %s", backendTypeOverride)
		}
		cfg.Providers.AWS.State.Backend = backendTypeOverride
		log.Infof("✓ Using backend type: %s", backendTypeOverride)
	}

	return nil
}
