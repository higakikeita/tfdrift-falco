package detector

import (
	"context"
	"fmt"
	"sync"

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
	cfg             *config.Config
	stateManager    *terraform.StateManager
	falcoSubscriber *falco.Subscriber
	notifier        *notifier.Manager
	formatter       *diff.DiffFormatter
	importer        *terraform.Importer
	approvalManager *terraform.ApprovalManager
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

// startCollectors starts the Falco event subscriber
func (d *Detector) startCollectors(ctx context.Context) error {
	log.Info("Starting Falco subscriber...")
	if err := d.falcoSubscriber.Start(ctx, d.eventCh); err != nil {
		return fmt.Errorf("falco subscriber error: %w", err)
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
		// Send alert for unmanaged resource
		d.sendUnmanagedResourceAlert(&event)

		// Handle auto-import if enabled
		if d.cfg.AutoImport.Enabled && d.importer != nil && d.approvalManager != nil {
			d.handleAutoImport(context.Background(), &event)
		}
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

		// Extract timestamp safely
		timestamp := ""
		if rawEvent, ok := event.RawEvent.(map[string]interface{}); ok {
			if eventTime, ok := rawEvent["eventTime"].(string); ok {
				timestamp = eventTime
			}
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
			Timestamp:    timestamp,
			AlertType:    "drift", // Mark as drift alert
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

// sendUnmanagedResourceAlert sends an alert for unmanaged resources
func (d *Detector) sendUnmanagedResourceAlert(event *types.Event) {
	// Extract timestamp from raw event
	timestamp := ""
	if rawEvent, ok := event.RawEvent.(map[string]interface{}); ok {
		if eventTime, ok := rawEvent["eventTime"].(string); ok {
			timestamp = eventTime
		}
	}

	alert := &types.UnmanagedResourceAlert{
		Severity:     "warning", // Default severity for unmanaged resources
		ResourceType: event.ResourceType,
		ResourceID:   event.ResourceID,
		EventName:    event.EventName,
		UserIdentity: event.UserIdentity,
		Changes:      event.Changes,
		Timestamp:    timestamp,
		Reason:       fmt.Sprintf("Resource %s (%s) is not found in Terraform state", event.ResourceID, event.ResourceType),
	}

	// Format and display
	consoleOutput := d.formatter.FormatUnmanagedResource(alert)
	fmt.Println(consoleOutput)

	// Also log
	log.Warnf("UNMANAGED RESOURCE: %s (%s) - Event: %s by %s",
		alert.ResourceID, alert.ResourceType, alert.EventName, alert.UserIdentity.UserName)

	if d.cfg.DryRun {
		log.Info("[DRY-RUN] Unmanaged resource alert notification skipped")
		fmt.Println("\n=== Markdown Format (for Slack) ===")
		fmt.Println(d.formatter.FormatUnmanagedResourceMarkdown(alert))
		return
	}

	// TODO: Send to notification channels
	// For now, console output is enough
}

// AttributeDrift represents a single attribute change
type AttributeDrift struct {
	Attribute string
	OldValue  interface{}
	NewValue  interface{}
}

// handleAutoImport handles automatic terraform import for unmanaged resources
func (d *Detector) handleAutoImport(ctx context.Context, event *types.Event) {
	log.Infof("Auto-import triggered for %s (%s)", event.ResourceID, event.ResourceType)

	// Create approval request
	userIdentity := fmt.Sprintf("%s (%s)", event.UserIdentity.UserName, event.UserIdentity.ARN)
	request := d.approvalManager.RequestApproval(
		event.ResourceType,
		event.ResourceID,
		event.Changes,
		userIdentity,
	)

	var result *terraform.ImportResult
	var err error

	// Handle based on approval mode
	if d.cfg.AutoImport.RequireApproval {
		// Manual approval mode - prompt user
		approved, promptErr := d.approvalManager.PromptForApproval(ctx, request)
		if promptErr != nil {
			log.Errorf("Failed to prompt for approval: %v", promptErr)
			return
		}

		if approved {
			fmt.Printf("🚀 Executing: %s\n", request.ImportCommand.String())
			result, err = d.approvalManager.ApproveAndExecute(ctx, request.ID, "console-user")
		} else {
			log.Info("Import rejected by user")
			return
		}
	} else {
		// Auto-approval mode - check whitelist
		result, err = d.approvalManager.AutoApproveIfAllowed(ctx, request, d.cfg.AutoImport.AllowedResources)
		if err != nil {
			log.Warnf("Auto-approval denied: %v", err)
			return
		}
	}

	// Handle result
	if err != nil {
		log.Errorf("Import failed: %v", err)
		fmt.Printf("❌ Import failed: %v\n", err)
		return
	}

	if result.Success {
		fmt.Println("✅ Import successful!")
		if result.GeneratedCode != "" {
			// Save the generated code to output directory
			outputFile := fmt.Sprintf("%s/%s_%s.tf",
				d.cfg.AutoImport.OutputDir,
				event.ResourceType,
				result.Command.ResourceName)
			fmt.Printf("📄 Generated Terraform code:\n%s\n", result.GeneratedCode)
			fmt.Printf("💡 Save this to: %s\n", outputFile)
		}
		log.Infof("Successfully imported %s", event.ResourceID)
	} else {
		fmt.Printf("❌ Import failed: %s\n", result.Error)
		log.Errorf("Import failed for %s: %s", event.ResourceID, result.Error)
	}
}
