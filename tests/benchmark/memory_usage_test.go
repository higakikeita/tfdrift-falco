package benchmark

import (
	"runtime"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
)

// TestMemoryUsage_EventProcessing measures memory usage for event processing
func TestMemoryUsage_EventProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	det := setupBenchmarkDetector(t)

	// Force GC before measurement
	runtime.GC()

	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Process 10,000 events
	numEvents := 10000
	for i := 0; i < numEvents; i++ {
		event := createBenchmarkEvent()
		det.HandleEventForTest(event)
	}

	// Force GC after processing
	runtime.GC()

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Calculate memory usage using TotalAlloc which only increases
	totalAllocDiff := memAfter.TotalAlloc - memBefore.TotalAlloc
	avgPerEvent := totalAllocDiff / uint64(numEvents)

	t.Logf("Memory Usage for %d events:", numEvents)
	t.Logf("  Total Allocated: %d bytes (%.2f MB)", totalAllocDiff, float64(totalAllocDiff)/1024/1024)
	t.Logf("  Avg Per Event: %d bytes", avgPerEvent)
	t.Logf("  Current Alloc Before: %d bytes", memBefore.Alloc)
	t.Logf("  Current Alloc After: %d bytes", memAfter.Alloc)
	t.Logf("  Heap Objects: %d", memAfter.HeapObjects)

	// Assert reasonable memory usage (<100KB per event accounting for GC)
	maxBytesPerEvent := uint64(100 * 1024)
	if avgPerEvent > maxBytesPerEvent {
		t.Errorf("Memory usage too high: %d bytes/event (max: %d)", avgPerEvent, maxBytesPerEvent)
	}
}

// TestMemoryLeak_LongRunning checks for memory leaks over time
func TestMemoryLeak_LongRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running memory leak test")
	}

	det := setupBenchmarkDetector(t)

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
			det.HandleEventForTest(event)
		}
	}

	// Check if memory grows linearly (potential leak)
	first := measurements[0]
	last := measurements[len(measurements)-1]

	var growth float64
	if last > first {
		growth = float64(last-first) / float64(first) * 100
	} else {
		// Memory decreased or stayed the same - no leak
		growth = 0
	}

	t.Logf("Memory Leak Check:")
	t.Logf("  First measurement: %d bytes (%.2f MB)", first, float64(first)/1024/1024)
	t.Logf("  Last measurement:  %d bytes (%.2f MB)", last, float64(last)/1024/1024)
	if last > first {
		t.Logf("  Growth: +%.2f%%", growth)
	} else {
		t.Logf("  Growth: %.2f%% (decreased)", float64(int64(last)-int64(first))/float64(first)*100)
	}

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

	det := setupBenchmarkDetector(t)

	runtime.GC()

	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Perform 1000 state comparisons
	numComparisons := 1000
	sm := det.GetStateManagerForTest().(*terraform.StateManager)
	for i := 0; i < numComparisons; i++ {
		resourceID := "i-test-instance"
		_, _ = sm.GetResource(resourceID)
	}

	runtime.GC()

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	totalAllocDiff := memAfter.TotalAlloc - memBefore.TotalAlloc
	avgPerComparison := totalAllocDiff / uint64(numComparisons)

	t.Logf("State Comparison Memory for %d comparisons:", numComparisons)
	t.Logf("  Total Allocated: %d bytes (%.2f MB)", totalAllocDiff, float64(totalAllocDiff)/1024/1024)
	t.Logf("  Avg: %d bytes/comparison", avgPerComparison)

	// Should use <50KB per comparison accounting for allocations
	maxBytes := uint64(50 * 1024)
	if avgPerComparison > maxBytes {
		t.Errorf("State comparison memory too high: %d bytes (max: %d)", avgPerComparison, maxBytes)
	}
}

// TestMemoryUsage_ConcurrentEvents measures memory under concurrent load
func TestMemoryUsage_ConcurrentEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent memory test")
	}

	det := setupBenchmarkDetector(t)

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
				det.HandleEventForTest(event)
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

	totalAllocDiff := memAfter.TotalAlloc - memBefore.TotalAlloc
	totalEvents := uint64(numGoroutines * eventsPerGoroutine)
	avgPerEvent := totalAllocDiff / totalEvents

	t.Logf("Concurrent Memory Usage:")
	t.Logf("  Goroutines: %d", numGoroutines)
	t.Logf("  Events: %d", totalEvents)
	t.Logf("  Total Allocated: %d bytes (%.2f MB)", totalAllocDiff, float64(totalAllocDiff)/1024/1024)
	t.Logf("  Avg: %d bytes/event", avgPerEvent)

	// Concurrent overhead - allow reasonable overhead
	maxBytes := uint64(150 * 1024) // Allow overhead for concurrency
	if avgPerEvent > maxBytes {
		t.Errorf("Concurrent memory usage too high: %d bytes/event (max: %d)", avgPerEvent, maxBytes)
	}
}
