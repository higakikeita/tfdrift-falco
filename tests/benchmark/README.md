# Benchmark Tests

Performance benchmarks for TFDrift-Falco to measure throughput, latency, and resource usage.

## Overview

Benchmark tests measure:
- **Event processing speed** - Events per second
- **State comparison latency** - Time to detect drift
- **Memory usage** - Bytes allocated per event
- **Concurrent handling** - Performance under load

## Running Benchmarks

```bash
# All benchmarks
go test -bench=. -benchmem -benchtime=10s

# Specific benchmark
go test -bench=BenchmarkEventProcessing -benchmem

# With CPU profiling
go test -bench=. -cpuprofile=cpu.prof

# With memory profiling
go test -bench=. -memprofile=mem.prof

# Compare with baseline
go test -bench=. -benchmem > new.txt
benchstat old.txt new.txt
```

## Expected Performance

### Baseline (Development Machine)

| Benchmark | Operations/sec | ns/op | bytes/op | allocs/op |
|-----------|---------------|-------|----------|-----------|
| EventProcessing | ~10,000 | ~100,000 | ~5,000 | ~50 |
| StateComparison | ~1,000 | ~1,000,000 | ~50,000 | ~500 |
| ConcurrentEvents | ~5,000 | ~200,000 | ~10,000 | ~100 |

*Note: Actual performance depends on hardware*

## Performance Goals

- **Event Processing**: >5,000 events/sec
- **State Comparison**: <1 second per resource
- **Memory**: <10MB per 1000 events
- **Concurrent**: Handle 100 concurrent events

## Benchmark Structure

```go
func BenchmarkEventProcessing(b *testing.B) {
    // Setup
    detector := setupTestDetector()
    event := createTestEvent()

    // Reset timer (exclude setup time)
    b.ResetTimer()

    // Run benchmark
    for i := 0; i < b.N; i++ {
        detector.HandleEvent(event)
    }
}
```

## Profiling

### CPU Profiling

```bash
# Generate CPU profile
go test -bench=BenchmarkEventProcessing -cpuprofile=cpu.prof

# Analyze with pprof
go tool pprof cpu.prof

# Web UI
go tool pprof -http=:8080 cpu.prof
```

### Memory Profiling

```bash
# Generate memory profile
go test -bench=BenchmarkEventProcessing -memprofile=mem.prof

# Analyze
go tool pprof mem.prof

# Check for leaks
go test -bench=BenchmarkEventProcessing -memprofilerate=1
```

### Flame Graphs

```bash
# Install flamegraph tool
go install github.com/uber/go-torch@latest

# Generate flame graph
go-torch -b cpu.prof
```

## Performance Monitoring

### Continuous Benchmarking

Benchmarks run nightly in CI/CD to detect regressions:

```yaml
# .github/workflows/benchmark.yml
- name: Run benchmarks
  run: |
    go test -bench=. -benchmem -run=^$ ./tests/benchmark/ \
      | tee benchmark-results.txt

- name: Compare with baseline
  run: benchstat baseline.txt benchmark-results.txt
```

### Performance Alerts

Alert if performance degrades >20%:
- Event processing drops below 4,000 ops/sec
- Memory usage increases >20%
- Latency increases >20%

## Optimization Tips

### Common Bottlenecks

1. **JSON Parsing** - Use efficient JSON libraries
2. **String Operations** - Minimize allocations
3. **Map Lookups** - Use caching for hot paths
4. **Goroutine Creation** - Pool goroutines

### Before Optimizing

1. Run benchmarks to establish baseline
2. Profile to find actual bottlenecks
3. Optimize the hot path
4. Re-run benchmarks to verify improvement

### Example Optimization

```go
// Before (slow)
func compareValues(a, b interface{}) bool {
    return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// After (fast)
func compareValues(a, b interface{}) bool {
    return reflect.DeepEqual(a, b)
}
```

## Adding New Benchmarks

1. Create benchmark function: `Benchmark<Name>(b *testing.B)`
2. Add fixtures in `fixtures/` if needed
3. Document expected performance
4. Add to CI/CD pipeline

---

**Tools**:
- [pprof](https://golang.org/pkg/runtime/pprof/)
- [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [go-torch](https://github.com/uber/go-torch)
