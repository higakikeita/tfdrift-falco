package benchmark

import (
	"runtime"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// TestMemoryUsage_EventProcessing measures memory usage for event processing
func TestMemoryUsage_EventProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	det := setupBenchmarkDetector(t.(*testing.T))

	// Force GC before measurement
	runtime.GC()

	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Process 10,000 events
	numEvents := 10000
	for i := 0; i < numEvents; i++ {
		event := createBenchmarkEvent()
		det.HandleEvent(event)
	}

	// Force GC after processing
	runtime.GC()

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Calculate memory usage
	allocDiff := memAfter.Alloc - memBefore.Alloc
	avgPerEvent := allocDiff / uint64(numEvents)

	t.Logf("Memory Usage for %d events:", numEvents)
	t.Logf("  Total Allocated: %d bytes (%.2f MB)", allocDiff, float64(allocDiff)/1024/1024)
	t.Logf("  Avg Per Event: %d bytes", avgPerEvent)
	t.Logf("  Total Allocs: %d", memAfter.TotalAlloc-memBefore.TotalAlloc)
	t.Logf("  Heap Objects: %d", memAfter.HeapObjects-memBefore.HeapObjects)

	// Assert reasonable memory usage (<10KB per event)
	maxBytesPerEvent := uint64(10 * 1024)
	if avgPerEvent > maxBytesPerEvent {
		t.Errorf("Memory usage too high: %d bytes/event (max: %d)", avgPerEvent, maxBytesPerEvent)
	}
}

// TestMemoryLeak_LongRunning checks for memory leaks over time
func TestMemoryLeak_LongRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running memory leak test")
	}

	det := setupBenchmarkDetector(t.(*testing.T))

	measurements := make([]uint64, 0, 10)

	// Process events in batches and measure memory
	for batch := 0; batch < 10; batch++ {
		runtime.GC()

		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		measurements = append(measurements, mem.Alloc)

		// Process 1000 events
		for i := 0; i < 1000; i++ {
			event := createBenchmarkEvent()
			det.HandleEvent(event)
		}
	}

	// Check if memory grows linearly (potential leak)
	first := measurements[0]
	last := measurements[len(measurements)-1]
	growth := float64(last-first) / float64(first) * 100

	t.Logf("Memory Leak Check:")
	t.Logf("  First measurement: %d bytes (%.2f MB)", first, float64(first)/1024/1024)
	t.Logf("  Last measurement:  %d bytes (%.2f MB)", last, float64(last)/1024/1024)
	t.Logf("  Growth: %.2f%%", growth)

	// Alert if memory grows >50% over 10 batches
	if growth > 50 {
		t.Errorf("Potential memory leak detected: %.2f%% growth", growth)
	}
}

// TestMemoryUsage_StateComparison measures memory for state comparisons
func TestMemoryUsage_StateComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	det := setupBenchmarkDetector(t.(*testing.T))

	runtime.GC()

	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Perform 1000 state comparisons
	numComparisons := 1000
	for i := 0; i < numComparisons; i++ {
		event := types.Event{
			Provider:     "aws",
			ResourceType: "aws_instance",
			ResourceID:   "i-test-instance",
		}
		_ = det.GetStateManager().GetResource(event.ResourceType, event.ResourceID)
	}

	runtime.GC()

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	allocDiff := memAfter.Alloc - memBefore.Alloc
	avgPerComparison := allocDiff / uint64(numComparisons)

	t.Logf("State Comparison Memory for %d comparisons:", numComparisons)
	t.Logf("  Total: %d bytes (%.2f MB)", allocDiff, float64(allocDiff)/1024/1024)
	t.Logf("  Avg: %d bytes/comparison", avgPerComparison)

	// Should use <5KB per comparison
	maxBytes := uint64(5 * 1024)
	if avgPerComparison > maxBytes {
		t.Errorf("State comparison memory too high: %d bytes (max: %d)", avgPerComparison, maxBytes)
	}
}

// TestMemoryUsage_ConcurrentEvents measures memory under concurrent load
func TestMemoryUsage_ConcurrentEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent memory test")
	}

	det := setupBenchmarkDetector(t.(*testing.T))

	runtime.GC()

	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Process events concurrently
	numGoroutines := 10
	eventsPerGoroutine := 100

	done := make(chan bool, numGoroutines)
	for g := 0; g < numGoroutines; g++ {
		go func() {
			for i := 0; i < eventsPerGoroutine; i++ {
				event := createBenchmarkEvent()
				det.HandleEvent(event)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for g := 0; g < numGoroutines; g++ {
		<-done
	}

	runtime.GC()

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	allocDiff := memAfter.Alloc - memBefore.Alloc
	totalEvents := uint64(numGoroutines * eventsPerGoroutine)
	avgPerEvent := allocDiff / totalEvents

	t.Logf("Concurrent Memory Usage:")
	t.Logf("  Goroutines: %d", numGoroutines)
	t.Logf("  Events: %d", totalEvents)
	t.Logf("  Total: %d bytes (%.2f MB)", allocDiff, float64(allocDiff)/1024/1024)
	t.Logf("  Avg: %d bytes/event", avgPerEvent)

	// Concurrent overhead should be minimal
	maxBytes := uint64(15 * 1024) // Allow 50% overhead for concurrency
	if avgPerEvent > maxBytes {
		t.Errorf("Concurrent memory usage too high: %d bytes/event (max: %d)", avgPerEvent, maxBytes)
	}
}
