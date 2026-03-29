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
	"github.com/keitahigaki/tfdrift-falco/pkg/policy"
	"github.com/keitahigaki/tfdrift-falco/pkg/provider"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// Detector is the main drift detection engine
type Detector struct {
	cfg               *config.Config
	stateManager      *terraform.StateManager       // Legacy: default state manager for backward compatibility
	stateManagers     map[string]*terraform.StateManager // Multi-provider state managers
	providerRegistry  *provider.Registry            // Provider registry for handling multiple clouds
	falcoSubscriber   *falco.Subscriber
	notifier          *notifier.Manager
	formatter         *diff.DiffFormatter
	importer          *terraform.Importer
	approvalManager   *terraform.ApprovalManager
	broadcaster       *broadcaster.Broadcaster
	graphStore        *graph.Store
	policyEngine      *policy.Engine
	eventCh           chan types.Event
	wg                sync.WaitGroup
}

// New creates a new Detector instance
func New(cfg *config.Config) (*Detector, error) {
	// Initialize provider registry
	registry := provider.NewRegistry()

	// Initialize state managers for each enabled provider
	stateManagers := make(map[string]*terraform.StateManager)

	// AWS provider
	var defaultStateManager *terraform.StateManager
	if cfg.Providers.AWS.Enabled {
		sm, err := terraform.NewStateManager(cfg.Providers.AWS.State)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS state manager: %w", err)
		}
		stateManagers["aws"] = sm
		defaultStateManager = sm

		awsOpts := []provider.AWSProviderOption{}
		if len(cfg.Providers.AWS.Regions) > 0 {
			awsOpts = append(awsOpts, provider.WithAWSRegions(cfg.Providers.AWS.Regions))
		}
		if err := registry.Register(provider.NewAWSProvider(awsOpts...)); err != nil {
			return nil, fmt.Errorf("failed to register AWS provider: %w", err)
		}
	}

	// GCP provider
	if cfg.Providers.GCP.Enabled {
		sm, err := terraform.NewStateManager(cfg.Providers.GCP.State)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCP state manager: %w", err)
		}
		stateManagers["gcp"] = sm

		// Use first enabled provider as default if AWS is not enabled
		if defaultStateManager == nil {
			defaultStateManager = sm
		}

		if err := registry.Register(provider.NewGCPProvider()); err != nil {
			return nil, fmt.Errorf("failed to register GCP provider: %w", err)
		}
	}

	// Azure provider
	if cfg.Providers.Azure.Enabled {
		sm, err := terraform.NewStateManager(cfg.Providers.Azure.State)
		if err != nil {
			return nil, fmt.Errorf("failed to create Azure state manager: %w", err)
		}
		stateManagers["azure"] = sm

		// Use first enabled provider as default if neither AWS nor GCP is enabled
		if defaultStateManager == nil {
			defaultStateManager = sm
		}

		if err := registry.Register(provider.NewAzureProvider()); err != nil {
			return nil, fmt.Errorf("failed to register Azure provider: %w", err)
		}
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

	// Initialize policy engine
	var policyEngine *policy.Engine
	if cfg.Policy.Enabled {
		policyEngine = policy.NewEngine()
		if cfg.Policy.PolicyDir != "" {
			if err := policyEngine.LoadDir(cfg.Policy.PolicyDir); err != nil {
				return nil, fmt.Errorf("failed to load policy files: %w", err)
			}
			log.Infof("Policy engine loaded %d module(s) from %s", policyEngine.ModuleCount(), cfg.Policy.PolicyDir)
		}
	}

	log.Infof("Initialized %d cloud provider(s): %v", registry.Count(), registry.Names())

	return &Detector{
		cfg:              cfg,
		stateManager:     defaultStateManager,
		stateManagers:    stateManagers,
		providerRegistry: registry,
		falcoSubscriber:  falcoSub,
		notifier:         notifierManager,
		formatter:        formatter,
		importer:         importer,
		approvalManager:  approvalManager,
		policyEngine:     policyEngine,
		eventCh:          make(chan types.Event, 100),
	}, nil
}

// GetStateManager returns the state manager for API access
func (d *Detector) GetStateManager() *terraform.StateManager {
	return d.stateManager
}

// GetProviderRegistry returns the provider registry
func (d *Detector) GetProviderRegistry() *provider.Registry {
	return d.providerRegistry
}

// GetStateManagers returns all state managers indexed by provider name
func (d *Detector) GetStateManagers() map[string]*terraform.StateManager {
	return d.stateManagers
}

// SetBroadcaster sets the broadcaster for real-time event distribution
func (d *Detector) SetBroadcaster(bc *broadcaster.Broadcaster) {
	d.broadcaster = bc
}

// GetBroadcaster returns the broadcaster
func (d *Detector) GetBroadcaster() *broadcaster.Broadcaster {
	return d.broadcaster
}

// GetPolicyEngine returns the policy engine
func (d *Detector) GetPolicyEngine() *policy.Engine {
	return d.policyEngine
}

// SetPolicyEngine sets the policy engine for drift evaluation
func (d *Detector) SetPolicyEngine(pe *policy.Engine) {
	d.policyEngine = pe
}

// SetGraphStore sets the graph store for drift visualization
func (d *Detector) SetGraphStore(gs *graph.Store) {
	d.graphStore = gs
}

// GetGraphStore returns the graph store
func (d *Detector) GetGraphStore() *graph.Store {
	return d.graphStore
}
