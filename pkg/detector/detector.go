package detector

import (
	"context"
	"fmt"
	"sync"

	"github.com/keitahigaki/tfdrift-falco/pkg/cloudtrail"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/diff"
	"github.com/keitahigaki/tfdrift-falco/pkg/falco"
	"github.com/keitahigaki/tfdrift-falco/pkg/notifier"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// Detector is the main drift detection engine
type Detector struct {
	cfg              *config.Config
	stateManager     *terraform.StateManager
	cloudtrail       *cloudtrail.Collector
	falcoSubscriber  *falco.Subscriber
	notifier         *notifier.Manager
	formatter        *diff.DiffFormatter
	eventCh          chan types.Event
	wg               sync.WaitGroup
}

// New creates a new Detector instance
func New(cfg *config.Config) (*Detector, error) {
	// Initialize state manager
	stateManager, err := terraform.NewStateManager(cfg.Providers.AWS.State)
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	// Initialize CloudTrail collector (if AWS is enabled)
	var ctCollector *cloudtrail.Collector
	if cfg.Providers.AWS.Enabled {
		ctCollector, err = cloudtrail.NewCollector(cfg.Providers.AWS.CloudTrail)
		if err != nil {
			return nil, fmt.Errorf("failed to create cloudtrail collector: %w", err)
		}
	}

	// Initialize Falco subscriber (if Falco is enabled)
	var falcoSub *falco.Subscriber
	if cfg.Falco.Enabled {
		falcoSub, err = falco.NewSubscriber(cfg.Falco)
		if err != nil {
			return nil, fmt.Errorf("failed to create falco subscriber: %w", err)
		}
	}

	// Initialize notifier
	notifierManager, err := notifier.NewManager(cfg.Notifications)
	if err != nil {
		return nil, fmt.Errorf("failed to create notifier: %w", err)
	}

	// Initialize diff formatter
	formatter := diff.NewFormatter(true) // Enable colors for console output

	return &Detector{
		cfg:             cfg,
		stateManager:    stateManager,
		cloudtrail:      ctCollector,
		falcoSubscriber: falcoSub,
		notifier:        notifierManager,
		formatter:       formatter,
		eventCh:         make(chan types.Event, 100),
	}, nil
}

// Start starts the drift detection process
func (d *Detector) Start(ctx context.Context) error {
	log.Info("Loading Terraform state...")
	if err := d.stateManager.Load(ctx); err != nil {
		return fmt.Errorf("failed to load terraform state: %w", err)
	}

	resourceCount := d.stateManager.ResourceCount()
	log.Infof("Loaded Terraform state: %d resources", resourceCount)

	// Start event collectors
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		if err := d.startCollectors(ctx); err != nil {
			log.Errorf("Collector error: %v", err)
		}
	}()

	// Start event processor
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.processEvents(ctx)
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Wait for goroutines to finish
	d.wg.Wait()

	return nil
}

// startCollectors starts all configured event collectors
func (d *Detector) startCollectors(ctx context.Context) error {
	// Start CloudTrail collector (if enabled)
	if d.cfg.Providers.AWS.Enabled && d.cloudtrail != nil {
		log.Info("Starting CloudTrail event collection...")
		if err := d.cloudtrail.Start(ctx, d.eventCh); err != nil {
			return fmt.Errorf("cloudtrail collector error: %w", err)
		}
	}

	// Start Falco subscriber (if enabled)
	if d.cfg.Falco.Enabled && d.falcoSubscriber != nil {
		log.Info("Starting Falco subscriber...")
		if err := d.falcoSubscriber.Start(ctx, d.eventCh); err != nil {
			return fmt.Errorf("falco subscriber error: %w", err)
		}
	}

	return nil
}

// processEvents processes events from the event channel
func (d *Detector) processEvents(ctx context.Context) {
	log.Info("Event processor started")

	for {
		select {
		case <-ctx.Done():
			log.Info("Event processor stopping...")
			return

		case event := <-d.eventCh:
			d.handleEvent(event)
		}
	}
}

// handleEvent processes a single event
func (d *Detector) handleEvent(event types.Event) {
	log.Debugf("Processing event: %s - %s", event.EventName, event.ResourceID)

	// Look up resource in Terraform state
	resource, exists := d.stateManager.GetResource(event.ResourceID)
	if !exists {
		log.Warnf("Resource %s not found in Terraform state (unmanaged resource)", event.ResourceID)
		// TODO: Send alert for unmanaged resource
		return
	}

	// Compare with state
	drifts := d.detectDrifts(resource, event.Changes)
	if len(drifts) == 0 {
		log.Debugf("No drift detected for %s", event.ResourceID)
		return
	}

	// Evaluate rules
	for _, drift := range drifts {
		matchedRules := d.evaluateRules(resource.Type, drift.Attribute)
		if len(matchedRules) == 0 {
			continue
		}

		alert := &types.DriftAlert{
			Severity:     d.getSeverity(matchedRules),
			ResourceType: resource.Type,
			ResourceName: resource.Name,
			ResourceID:   event.ResourceID,
			Attribute:    drift.Attribute,
			OldValue:     drift.OldValue,
			NewValue:     drift.NewValue,
			UserIdentity: event.UserIdentity,
			MatchedRules: matchedRules,
			Timestamp:    event.RawEvent.(map[string]interface{})["eventTime"].(string),
		}

		d.sendAlert(alert)
	}
}

// detectDrifts detects attribute changes
func (d *Detector) detectDrifts(resource *terraform.Resource, changes map[string]interface{}) []AttributeDrift {
	var drifts []AttributeDrift

	for key, newValue := range changes {
		oldValue, exists := resource.Attributes[key]
		if !exists || oldValue != newValue {
			drifts = append(drifts, AttributeDrift{
				Attribute: key,
				OldValue:  oldValue,
				NewValue:  newValue,
			})
		}
	}

	return drifts
}

// evaluateRules evaluates drift rules
func (d *Detector) evaluateRules(resourceType, attribute string) []string {
	var matched []string

	for _, rule := range d.cfg.DriftRules {
		// Check if resource type matches
		typeMatch := false
		for _, rt := range rule.ResourceTypes {
			if rt == resourceType {
				typeMatch = true
				break
			}
		}
		if !typeMatch {
			continue
		}

		// Check if attribute matches
		attrMatch := false
		for _, wa := range rule.WatchedAttributes {
			if wa == attribute {
				attrMatch = true
				break
			}
		}
		if !attrMatch {
			continue
		}

		matched = append(matched, rule.Name)
	}

	return matched
}

// getSeverity determines the highest severity from matched rules
func (d *Detector) getSeverity(matchedRules []string) string {
	severity := "low"

	for _, ruleName := range matchedRules {
		for _, rule := range d.cfg.DriftRules {
			if rule.Name == ruleName {
				if rule.Severity == "critical" {
					return "critical"
				}
				if rule.Severity == "high" && severity != "critical" {
					severity = "high"
				}
				if rule.Severity == "medium" && severity == "low" {
					severity = "medium"
				}
			}
		}
	}

	return severity
}

// sendAlert sends a drift alert
func (d *Detector) sendAlert(alert *types.DriftAlert) {
	// Format and display the drift in console
	consoleDiff := d.formatter.FormatConsole(alert)
	fmt.Println(consoleDiff)

	// Also log in traditional format
	log.Warnf("DRIFT DETECTED: %s.%s - %s: %v → %v",
		alert.ResourceType, alert.ResourceName, alert.Attribute,
		alert.OldValue, alert.NewValue)

	if d.cfg.DryRun {
		log.Info("[DRY-RUN] Alert notification skipped")

		// In dry-run, also show other formats as examples
		fmt.Println("\n=== Unified Diff Format ===")
		fmt.Println(d.formatter.FormatUnifiedDiff(alert))

		fmt.Println("\n=== Side-by-Side Format ===")
		fmt.Println(d.formatter.FormatSideBySide(alert))

		fmt.Println("\n=== Markdown Format (for Slack/GitHub) ===")
		fmt.Println(d.formatter.FormatMarkdown(alert))

		return
	}

	if err := d.notifier.Send(alert); err != nil {
		log.Errorf("Failed to send alert: %v", err)
	}
}

// AttributeDrift represents a single attribute change
type AttributeDrift struct {
	Attribute string
	OldValue  interface{}
	NewValue  interface{}
}
