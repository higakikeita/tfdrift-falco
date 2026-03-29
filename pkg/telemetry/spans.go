package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/keitahigaki/tfdrift-falco"

// Span attribute keys used across the pipeline.
var (
	AttrProvider     = attribute.Key("tfdrift.provider")
	AttrResourceType = attribute.Key("tfdrift.resource_type")
	AttrResourceID   = attribute.Key("tfdrift.resource_id")
	AttrEventName    = attribute.Key("tfdrift.event_name")
	AttrEventSource  = attribute.Key("tfdrift.event_source")
	AttrSeverity     = attribute.Key("tfdrift.severity")
	AttrChangeType   = attribute.Key("tfdrift.change_type")
	AttrUserID       = attribute.Key("tfdrift.user_id")
	AttrRegion       = attribute.Key("tfdrift.region")
	AttrDriftCount   = attribute.Key("tfdrift.drift_count")
	AttrRuleCount    = attribute.Key("tfdrift.rule_count")
	AttrChannel      = attribute.Key("tfdrift.notification_channel")
)

// StartSpan starts a new span using the global tracer provider.
// This is a convenience function for callers that don't hold a Tracer reference.
func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(instrumentationName).Start(ctx, spanName, opts...)
}

// StartSpanWithTracer starts a new span using a specific tracer.
func StartSpanWithTracer(ctx context.Context, tracer trace.Tracer, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if tracer == nil {
		tracer = otel.Tracer(instrumentationName)
	}
	return tracer.Start(ctx, spanName, opts...)
}

// RecordError records an error on the span and sets status to Error.
func RecordError(span trace.Span, err error) {
	if err != nil && span != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetOK sets the span status to OK.
func SetOK(span trace.Span) {
	if span != nil {
		span.SetStatus(codes.Ok, "")
	}
}

// EventAttrs returns common span attributes for a Falco event.
func EventAttrs(provider, source, eventName, resourceType, resourceID string) []attribute.KeyValue {
	return []attribute.KeyValue{
		AttrProvider.String(provider),
		AttrEventSource.String(source),
		AttrEventName.String(eventName),
		AttrResourceType.String(resourceType),
		AttrResourceID.String(resourceID),
	}
}

// DriftAttrs returns common span attributes for a detected drift.
func DriftAttrs(provider, resourceType, resourceID, severity, changeType string) []attribute.KeyValue {
	return []attribute.KeyValue{
		AttrProvider.String(provider),
		AttrResourceType.String(resourceType),
		AttrResourceID.String(resourceID),
		AttrSeverity.String(severity),
		AttrChangeType.String(changeType),
	}
}
