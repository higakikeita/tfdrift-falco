// Package app provides the core application logic for TFDrift-Falco.
package app

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/keitahigaki/tfdrift-falco/pkg/api"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/detector"
	log "github.com/sirupsen/logrus"
)

// Config holds the configuration flags and settings for the application
type Config struct {
	ConfigFile          string
	AutoDetect          bool
	OutputMode          string
	DryRun              bool
	Daemon              bool
	Interactive         bool
	ServerMode          bool
	APIPort             int
	RegionOverride      []string
	FalcoEndpoint       string
	StatePathOverride   string
	BackendTypeOverride string
	Version             string
}

// App represents the main application instance
type App struct {
	cfg    *Config
	appCfg *config.Config
}

// New creates a new App instance with the given configuration
func New(cfg *Config) *App {
	return &App{
		cfg: cfg,
	}
}

// Run executes the main application logic and returns an error if something fails
func (a *App) Run(ctx context.Context) error {
	log.Infof("Starting TFDrift-Falco v%s", a.cfg.Version)

	// Load configuration
	if err := a.loadConfig(); err != nil {
		return err
	}

	// Apply dry-run mode if requested
	if a.cfg.DryRun {
		log.Info("Running in DRY-RUN mode - notifications disabled")
		a.appCfg.DryRun = true
	}

	// Initialize detector
	det, err := detector.New(a.appCfg)
	if err != nil {
		return fmt.Errorf("failed to initialize detector: %w", err)
	}

	// Run detector or API server
	if a.cfg.ServerMode {
		return a.runAPIServer(ctx, det)
	}

	return a.runDetector(ctx, det)
}

// loadConfig loads the application configuration
func (a *App) loadConfig() error {
	var cfg *config.Config
	var err error

	if a.cfg.AutoDetect {
		log.Info("Auto-detection mode enabled")
		cfg, err = a.loadAutoConfig()
		if err != nil {
			return fmt.Errorf("auto-detection failed: %w", err)
		}
	} else {
		if a.cfg.ConfigFile == "" {
			return fmt.Errorf("config file not specified and auto-detection disabled")
		}
		cfg, err = config.Load(a.cfg.ConfigFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	a.appCfg = cfg
	return nil
}

// loadAutoConfig loads configuration through auto-detection
func (a *App) loadAutoConfig() (*config.Config, error) {
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
	if err := a.applyConfigOverrides(cfg); err != nil {
		return nil, fmt.Errorf("failed to apply config overrides: %w", err)
	}

	log.Info("✓ Auto-detection successful, starting drift detection...")
	return cfg, nil
}

// applyConfigOverrides applies L1 semi-auto mode flag overrides to the config
func (a *App) applyConfigOverrides(cfg *config.Config) error {
	// Override AWS regions if specified
	if len(a.cfg.RegionOverride) > 0 {
		cfg.Providers.AWS.Regions = a.cfg.RegionOverride
		log.Infof("✓ Using custom region(s): %v", a.cfg.RegionOverride)
	}

	// Override Falco endpoint if specified
	if a.cfg.FalcoEndpoint != "" {
		parts := strings.Split(a.cfg.FalcoEndpoint, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid Falco endpoint format (expected host:port): %s", a.cfg.FalcoEndpoint)
		}

		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid port in Falco endpoint: %s", parts[1])
		}

		cfg.Falco.Hostname = parts[0]
		cfg.Falco.Port = uint16(port)
		log.Infof("✓ Using custom Falco endpoint: %s", a.cfg.FalcoEndpoint)
	}

	// Override state path if specified
	if a.cfg.StatePathOverride != "" {
		cfg.Providers.AWS.State.LocalPath = a.cfg.StatePathOverride
		cfg.Providers.AWS.State.Backend = "local"
		log.Infof("✓ Using custom state path: %s", a.cfg.StatePathOverride)
	}

	// Override backend type if specified
	if a.cfg.BackendTypeOverride != "" {
		if a.cfg.BackendTypeOverride != "local" && a.cfg.BackendTypeOverride != "s3" {
			return fmt.Errorf("invalid backend type (must be 'local' or 's3'): %s", a.cfg.BackendTypeOverride)
		}
		cfg.Providers.AWS.State.Backend = a.cfg.BackendTypeOverride
		log.Infof("✓ Using backend type: %s", a.cfg.BackendTypeOverride)
	}

	return nil
}

// runDetector runs the detector in standard detection mode
func (a *App) runDetector(ctx context.Context, det *detector.Detector) error {
	log.Info("Drift detection started")
	if err := det.Start(ctx); err != nil {
		return fmt.Errorf("detector error: %w", err)
	}
	log.Info("TFDrift-Falco stopped")
	return nil
}

// runAPIServer runs the detector with the API server
func (a *App) runAPIServer(ctx context.Context, det *detector.Detector) error {
	log.Info("Starting TFDrift-Falco in API server mode")
	srv := api.NewServer(a.appCfg, det, a.cfg.APIPort, a.cfg.Version)

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
		return fmt.Errorf("API server error: %w", err)
	}

	return nil
}
