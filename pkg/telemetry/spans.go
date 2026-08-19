package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/higakikeita/driftwire"

// Span attribute keys used across the pipeline.
var (
	AttrProvider     = attribute.Key("driftwire.provider")
	AttrResourceType = attribute.Key("driftwire.resource_type")
	AttrResourceID   = attribute.Key("driftwire.resource_id")
	AttrEventName    = attribute.Key("driftwire.event_name")
	AttrEventSource  = attribute.Key("driftwire.event_source")
	AttrSeverity     = attribute.Key("driftwire.severity")
	AttrChangeType   = attribute.Key("driftwire.change_type")
	AttrUserID       = attribute.Key("driftwire.user_id")
	AttrRegion       = attribute.Key("driftwire.region")
	AttrDriftCount   = attribute.Key("driftwire.drift_count")
	AttrRuleCount    = attribute.Key("driftwire.rule_count")
	AttrChannel      = attribute.Key("driftwire.notification_channel")
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
