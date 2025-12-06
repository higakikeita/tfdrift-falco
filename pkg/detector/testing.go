package detector

import "github.com/keitahigaki/tfdrift-falco/pkg/types"

// HandleEventForTest is a test helper that exposes handleEvent for benchmarking
// and testing purposes. It should not be used in production code.
func (d *Detector) HandleEventForTest(event types.Event) {
	d.handleEvent(event)
}

// GetStateManagerForTest is a test helper that exposes the state manager
// for testing purposes. It should not be used in production code.
func (d *Detector) GetStateManagerForTest() interface{} {
	return d.stateManager
}
