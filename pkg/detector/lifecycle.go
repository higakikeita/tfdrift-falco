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

	// Wait for context cancellation
	<-ctx.Done()

	// Wait for goroutines to finish
	d.wg.Wait()

	return nil
}

// startCollectors starts the Falco event source. With the HTTP transport
// (ADR-006) there is no outbound stream to maintain — Falco POSTs alerts to the
// receiver route mounted by the API server — so this goroutine just parks until
// shutdown. With the legacy gRPC transport it runs the reconnecting Sub loop.
func (d *Detector) startCollectors(ctx context.Context) error {
	if d.cfg.Falco.UsesHTTPTransport() {
		log.Info("Falco transport=http: ingesting alerts via the HTTP receiver route")
		<-ctx.Done()
		return ctx.Err()
	}

	log.Info("Starting Falco subscriber (transport=grpc)...")
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
