package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "tfdrift-falco", cfg.ServiceName)
	assert.Equal(t, "0.9.0", cfg.ServiceVersion)
	assert.Equal(t, "development", cfg.Environment)
	assert.False(t, cfg.Enabled)
	assert.True(t, cfg.Traces.Enabled)
	assert.Equal(t, "localhost:4317", cfg.Traces.Endpoint)
	assert.Equal(t, 1.0, cfg.Traces.SamplingRatio)
	assert.True(t, cfg.Metrics.Enabled)
	assert.True(t, cfg.Metrics.OTLPEnabled)
}

func TestNewProvider_Disabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = false

	p, shutdown, err := NewProvider(cfg)
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.NotNil(t, shutdown)
	assert.False(t, p.IsEnabled())
	assert.Nil(t, p.TracerProvider())
	assert.Nil(t, p.MeterProvider())

	// Shutdown should be safe to call
	shutdown()
}

func TestNewProvider_Enabled_NoEndpoint(t *testing.T) {
	// With tracing enabled but endpoint unreachable, provider still
	// initializes (OTLP exporter uses async batch sender).
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.Traces.Enabled = true
	cfg.Traces.Insecure = true
	cfg.Metrics.Enabled = false

	p, shutdown, err := NewProvider(cfg)
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.True(t, p.IsEnabled())
	assert.NotNil(t, p.TracerProvider())

	// Get a tracer
	tracer := p.Tracer("test")
	assert.NotNil(t, tracer)

	shutdown()
}

func TestNewProvider_MetricsEnabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.Traces.Enabled = false
	cfg.Metrics.Enabled = true
	cfg.Metrics.OTLPEnabled = true
	cfg.Metrics.Insecure = true
	cfg.Metrics.IntervalSeconds = 1

	p, shutdown, err := NewProvider(cfg)
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.Nil(t, p.TracerProvider())
	assert.NotNil(t, p.MeterProvider())

	shutdown()
}

func TestBuildResource(t *testing.T) {
	cfg := Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "staging",
		Attributes: map[string]string{
			"cluster": "primary",
			"region":  "us-east-1",
		},
	}

	res, err := buildResource(cfg)
	require.NoError(t, err)
	assert.NotNil(t, res)

	// Check that resource has attributes
	attrs := res.Attributes()
	assert.True(t, len(attrs) > 0)
}

func TestTracer_WithNilProvider(t *testing.T) {
	p := &Provider{config: Config{}}
	tracer := p.Tracer("test")
	assert.NotNil(t, tracer) // Should return global no-op tracer
}
