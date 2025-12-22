package detector

import (
	"fmt"
	"sync"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/broadcaster"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/diff"
	"github.com/keitahigaki/tfdrift-falco/pkg/falco"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	"github.com/keitahigaki/tfdrift-falco/pkg/notifier"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// Detector is the main drift detection engine
type Detector struct {
	cfg             *config.Config
	stateManager    *terraform.StateManager
	falcoSubscriber *falco.Subscriber
	notifier        *notifier.Manager
	formatter       *diff.DiffFormatter
	importer        *terraform.Importer
	approvalManager *terraform.ApprovalManager
	broadcaster     *broadcaster.Broadcaster
	graphStore      *graph.Store
	eventCh         chan types.Event
	wg              sync.WaitGroup
}

// New creates a new Detector instance
func New(cfg *config.Config) (*Detector, error) {
	// Initialize state manager
	stateManager, err := terraform.NewStateManager(cfg.Providers.AWS.State)
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	// Initialize Falco subscriber
	falcoSub, err := falco.NewSubscriber(cfg.Falco)
	if err != nil {
		return nil, fmt.Errorf("failed to create falco subscriber: %w", err)
	}

	// Initialize notifier
	notifierManager, err := notifier.NewManager(cfg.Notifications)
	if err != nil {
		return nil, fmt.Errorf("failed to create notifier: %w", err)
	}

	// Initialize diff formatter
	formatter := diff.NewFormatter(true) // Enable colors for console output

	// Initialize importer and approval manager (only if auto_import is enabled)
	var importer *terraform.Importer
	var approvalManager *terraform.ApprovalManager

	if cfg.AutoImport.Enabled {
		importer = terraform.NewImporter(
			cfg.AutoImport.TerraformDir,
			cfg.DryRun,
		)

		// Check if interactive mode should be enabled
		interactiveMode := cfg.AutoImport.RequireApproval
		approvalManager = terraform.NewApprovalManager(importer, interactiveMode)

		log.Info("Auto-import feature enabled")
		if cfg.AutoImport.RequireApproval {
			log.Info("Approval workflow: MANUAL (interactive prompts)")
		} else if len(cfg.AutoImport.AllowedResources) > 0 {
			log.Infof("Approval workflow: AUTO (whitelist: %v)", cfg.AutoImport.AllowedResources)
		} else {
			log.Warn("Approval workflow: AUTO (all resources) - USE WITH CAUTION")
		}
	}

	return &Detector{
		cfg:             cfg,
		stateManager:    stateManager,
		falcoSubscriber: falcoSub,
		notifier:        notifierManager,
		formatter:       formatter,
		importer:        importer,
		approvalManager: approvalManager,
		eventCh:         make(chan types.Event, 100),
	}, nil
}

// GetStateManager returns the state manager for API access
func (d *Detector) GetStateManager() *terraform.StateManager {
	return d.stateManager
}

// SetBroadcaster sets the broadcaster for real-time event distribution
func (d *Detector) SetBroadcaster(bc *broadcaster.Broadcaster) {
	d.broadcaster = bc
}

// GetBroadcaster returns the broadcaster
func (d *Detector) GetBroadcaster() *broadcaster.Broadcaster {
	return d.broadcaster
}

// SetGraphStore sets the graph store for drift visualization
func (d *Detector) SetGraphStore(gs *graph.Store) {
	d.graphStore = gs
}

// GetGraphStore returns the graph store
func (d *Detector) GetGraphStore() *graph.Store {
	return d.graphStore
}
