package detector

import (
	"context"
	"fmt"

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
