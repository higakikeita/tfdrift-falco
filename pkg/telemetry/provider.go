// Package telemetry provides OpenTelemetry integration for TFDrift-Falco.
// It manages trace and metric providers with OTLP export and Prometheus bridge.
package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"

	log "github.com/sirupsen/logrus"
)

// Config holds OpenTelemetry configuration.
type Config struct {
	Enabled        bool              `yaml:"enabled" json:"enabled"`
	ServiceName    string            `yaml:"service_name" json:"service_name"`
	ServiceVersion string            `yaml:"service_version" json:"service_version"`
	Environment    string            `yaml:"environment" json:"environment"`
	Traces         TracesConfig      `yaml:"traces" json:"traces"`
	Metrics        MetricsConfig     `yaml:"metrics" json:"metrics"`
	Attributes     map[string]string `yaml:"attributes" json:"attributes"`
}

// TracesConfig holds trace-specific configuration.
type TracesConfig struct {
	Enabled       bool    `yaml:"enabled" json:"enabled"`
	Endpoint      string  `yaml:"endpoint" json:"endpoint"`
	SamplingRatio float64 `yaml:"sampling_ratio" json:"sampling_ratio"`
	Insecure      bool    `yaml:"insecure" json:"insecure"`
}

// MetricsConfig holds metrics-specific configuration.
type MetricsConfig struct {
	Enabled         bool   `yaml:"enabled" json:"enabled"`
	OTLPEnabled     bool   `yaml:"otlp_enabled" json:"otlp_enabled"`
	OTLPEndpoint    string `yaml:"otlp_endpoint" json:"otlp_endpoint"`
	IntervalSeconds int    `yaml:"interval_seconds" json:"interval_seconds"`
	Insecure        bool   `yaml:"insecure" json:"insecure"`
}

// Provider wraps OpenTelemetry trace and metric providers.
type Provider struct {
	config         Config
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:        false,
		ServiceName:    "tfdrift-falco",
		ServiceVersion: "0.9.0",
		Environment:    "development",
		Traces: TracesConfig{
			Enabled:       true,
			Endpoint:      "localhost:4317",
			SamplingRatio: 1.0,
			Insecure:      true,
		},
		Metrics: MetricsConfig{
			Enabled:         true,
			OTLPEnabled:     true,
			OTLPEndpoint:    "localhost:4317",
			IntervalSeconds: 60,
			Insecure:        true,
		},
	}
}

// NewProvider creates and initializes an OpenTelemetry provider.
// Returns a shutdown function that must be called on application exit.
func NewProvider(cfg Config) (*Provider, func(), error) {
	if !cfg.Enabled {
		log.Info("OpenTelemetry is disabled")
		return &Provider{config: cfg}, func() {}, nil
	}

	ctx := context.Background()

	// Build resource with service metadata
	res, err := buildResource(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTel resource: %w", err)
	}

	p := &Provider{config: cfg}

	// Initialize trace provider
	if cfg.Traces.Enabled {
		tp, err := initTracerProvider(ctx, cfg, res)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to init tracer provider: %w", err)
		}
		p.tracerProvider = tp
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
		log.WithField("endpoint", cfg.Traces.Endpoint).Info("OpenTelemetry tracing enabled")
	}

	// Initialize metric provider
	if cfg.Metrics.Enabled && cfg.Metrics.OTLPEnabled {
		mp, err := initMeterProvider(ctx, cfg, res)
		if err != nil {
			log.WithError(err).Warn("Failed to init OTLP metric provider, continuing without OTLP metrics")
		} else {
			p.meterProvider = mp
			otel.SetMeterProvider(mp)
			log.WithField("endpoint", cfg.Metrics.OTLPEndpoint).Info("OpenTelemetry OTLP metrics enabled")
		}
	}

	shutdown := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if p.tracerProvider != nil {
			if err := p.tracerProvider.Shutdown(shutdownCtx); err != nil {
				log.WithError(err).Error("Failed to shutdown tracer provider")
			}
		}
		if p.meterProvider != nil {
			if err := p.meterProvider.Shutdown(shutdownCtx); err != nil {
				log.WithError(err).Error("Failed to shutdown meter provider")
			}
		}
		log.Info("OpenTelemetry providers shut down")
	}

	return p, shutdown, nil
}

// Tracer returns a named tracer from the provider.
func (p *Provider) Tracer(name string) trace.Tracer {
	if p.tracerProvider != nil {
		return p.tracerProvider.Tracer(name)
	}
	return otel.Tracer(name)
}

// IsEnabled returns whether telemetry is enabled.
func (p *Provider) IsEnabled() bool {
	return p.config.Enabled
}

// TracerProvider returns the underlying SDK TracerProvider.
func (p *Provider) TracerProvider() *sdktrace.TracerProvider {
	return p.tracerProvider
}

// MeterProvider returns the underlying SDK MeterProvider.
func (p *Provider) MeterProvider() *sdkmetric.MeterProvider {
	return p.meterProvider
}

func buildResource(cfg Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(cfg.ServiceName),
		semconv.ServiceVersionKey.String(cfg.ServiceVersion),
		attribute.String("deployment.environment", cfg.Environment),
	}

	for k, v := range cfg.Attributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attrs...),
	)
}

func initTracerProvider(ctx context.Context, cfg Config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Traces.Endpoint),
	}
	if cfg.Traces.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	var sampler sdktrace.Sampler
	switch {
	case cfg.Traces.SamplingRatio <= 0:
		sampler = sdktrace.NeverSample()
	case cfg.Traces.SamplingRatio >= 1.0:
		sampler = sdktrace.AlwaysSample()
	default:
		sampler = sdktrace.TraceIDRatioBased(cfg.Traces.SamplingRatio)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	return tp, nil
}

func initMeterProvider(ctx context.Context, cfg Config, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.Metrics.OTLPEndpoint),
	}
	if cfg.Metrics.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	interval := time.Duration(cfg.Metrics.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 60 * time.Second
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(interval))),
	)

	return mp, nil
}
