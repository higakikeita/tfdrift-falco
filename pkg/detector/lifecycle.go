package detector

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// Start starts the drift detection process
func (d *Detector) Start(ctx context.Context) error {
	log.Info("Loading Terraform state...")
	if err := d.stateManager.Load(ctx); err != nil {
		return fmt.Errorf("failed to load terraform state: %w", err)
	}

	resourceCount := d.stateManager.ResourceCount()
	log.Infof("Loaded Terraform state: %d resources", resourceCount)

	// Rebuild graph database with loaded resources
	if d.graphStore == nil {
		log.Warn("GraphStore is nil in detector, cannot rebuild graph database")
	} else {
		log.Info("GraphStore is available, rebuilding graph database...")
		d.graphStore.RebuildGraphDB()
		log.Info("Rebuilt graph database with loaded resources")
	}

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

	// Periodically re-read Terraform state so legit applies aren't flagged as
	// drift forever (#331). Disabled (load-once) when the interval is 0.
	if d.cfg.StateRefreshIntervalSec > 0 {
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			d.refreshStatePeriodically(ctx)
		}()
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Wait for goroutines to finish
	d.wg.Wait()

	return nil
}

// refreshStatePeriodically re-reads every provider's Terraform state on a timer
// and rebuilds the graph, so a running detector picks up legitimate applies
// instead of comparing against the startup snapshot forever (#331).
func (d *Detector) refreshStatePeriodically(ctx context.Context) {
	interval := time.Duration(d.cfg.StateRefreshIntervalSec) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	log.Infof("Terraform state refresh enabled: every %s", interval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.refreshAllState(ctx)
		}
	}
}

// refreshAllState re-reads every provider's state once and rebuilds the graph.
// Extracted from the ticker loop so the refresh is unit-testable without waiting
// on wall-clock time (#331).
func (d *Detector) refreshAllState(ctx context.Context) {
	for name, sm := range d.stateManagers {
		if err := sm.Refresh(ctx); err != nil {
			log.Warnf("State refresh failed for provider %s: %v", name, err)
		}
	}
	if d.graphStore != nil {
		d.graphStore.RebuildGraphDB()
	}
	log.Debug("Terraform state refreshed")
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
