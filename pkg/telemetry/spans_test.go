package telemetry

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestStartSpan(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test.span")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestStartSpanWithTracer(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	ctx := context.Background()

	ctx, span := StartSpanWithTracer(ctx, tracer, "test.span")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestStartSpanWithTracer_NilTracer(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpanWithTracer(ctx, nil, "test.span")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestRecordError(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	_, span := tracer.Start(context.Background(), "test")

	// Should not panic with nil error
	RecordError(span, nil)

	// Should not panic with real error
	RecordError(span, errors.New("test error"))

	// Should not panic with nil span
	RecordError(nil, errors.New("test error"))

	span.End()
}

func TestSetOK(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	_, span := tracer.Start(context.Background(), "test")

	// Should not panic
	SetOK(span)
	SetOK(nil) // nil span should be safe

	span.End()
}

func TestEventAttrs(t *testing.T) {
	attrs := EventAttrs("aws", "aws_cloudtrail", "RunInstances", "aws_instance", "i-12345")
	assert.Len(t, attrs, 5)
	assert.Equal(t, "aws", attrs[0].Value.AsString())
	assert.Equal(t, "aws_cloudtrail", attrs[1].Value.AsString())
	assert.Equal(t, "RunInstances", attrs[2].Value.AsString())
	assert.Equal(t, "aws_instance", attrs[3].Value.AsString())
	assert.Equal(t, "i-12345", attrs[4].Value.AsString())
}

func TestDriftAttrs(t *testing.T) {
	attrs := DriftAttrs("gcp", "google_compute_instance", "inst-1", "high", "modified")
	assert.Len(t, attrs, 5)
	assert.Equal(t, "gcp", attrs[0].Value.AsString())
	assert.Equal(t, "google_compute_instance", attrs[1].Value.AsString())
}

func TestAttrKeys(t *testing.T) {
	// Verify attribute keys are properly defined
	assert.Equal(t, "tfdrift.provider", string(AttrProvider))
	assert.Equal(t, "tfdrift.resource_type", string(AttrResourceType))
	assert.Equal(t, "tfdrift.resource_id", string(AttrResourceID))
	assert.Equal(t, "tfdrift.event_name", string(AttrEventName))
	assert.Equal(t, "tfdrift.event_source", string(AttrEventSource))
	assert.Equal(t, "tfdrift.severity", string(AttrSeverity))
	assert.Equal(t, "tfdrift.change_type", string(AttrChangeType))
	assert.Equal(t, "tfdrift.user_id", string(AttrUserID))
	assert.Equal(t, "tfdrift.region", string(AttrRegion))
	assert.Equal(t, "tfdrift.drift_count", string(AttrDriftCount))
	assert.Equal(t, "tfdrift.rule_count", string(AttrRuleCount))
	assert.Equal(t, "tfdrift.notification_channel", string(AttrChannel))
}

// Verify the codes package is usable for span status
func TestCodesConstants(t *testing.T) {
	assert.Equal(t, codes.Ok, codes.Ok)
	assert.Equal(t, codes.Error, codes.Error)
}

// Verify trace.SpanKind is available
func TestSpanKinds(t *testing.T) {
	assert.Equal(t, trace.SpanKindServer, trace.SpanKindServer)
	assert.Equal(t, trace.SpanKindClient, trace.SpanKindClient)
}
