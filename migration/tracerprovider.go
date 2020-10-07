package migration

import (
	"context"
	"time"

	oteltrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/label"

	octrace "go.opencensus.io/trace"
)

// NewTracerProvider is an implementation of the TraceProvider interface which
// uses opencensus under the hood.
func NewTracerProvider() oteltrace.TracerProvider {
	return &opencensusTraceProvider{}
}

type opencensusTraceProvider struct{}

var _ oteltrace.TracerProvider = &opencensusTraceProvider{}

func (w *opencensusTraceProvider) Tracer(instrumentationName string, opts ...oteltrace.TracerOption) oteltrace.Tracer {
	return opencensusTracer{opts: opts, name: instrumentationName}
}

var _ oteltrace.Tracer = &opencensusTracer{}

type opencensusTracer struct {
	// TODO what do we need to do with the name?
	name string
	// TODO handle tracer options
	opts []oteltrace.TracerOption
}

func (o opencensusTracer) Start(ctx context.Context, spanName string, opts ...oteltrace.SpanOption) (context.Context, oteltrace.Span) {
	ctx, span := octrace.StartSpan(ctx, spanName)
	ocSpan := &opencensusSpan{span: span, tracer: o}
	spanConfig := oteltrace.NewSpanConfig(opts...)
	ocSpan.SetAttributes(spanConfig.Attributes...)
	// TODO Handle the following SpanConfig pieces
	// Timestamp time.Time
	// Links []Link
	// Record bool
	// NewRoot bool
	// SpanKind SpanKind

	// for _, link := range spanConfig.Links {
	// 	span.AddLink(trace.Link{
	// 		TraceID: link.TraceID
	// 		SpanID: link.SpanID
	// 		Type    oteltrace.Link
	// 		// Attributes is a set of attributes on the link.
	// 		Attributes map[string]interface{}
	// 	})
	// }
	return ctx, ocSpan
}

var _ oteltrace.Span = &opencensusSpan{}

type opencensusSpan struct {
	span   *octrace.Span
	tracer oteltrace.Tracer
}

func (o *opencensusSpan) Tracer() oteltrace.Tracer {
	return o.tracer
}

func (o *opencensusSpan) End(options ...oteltrace.SpanOption) {
	// OpenCensus End does not take span options
	// TODO: Figure out what we should do.
	o.span.End()
}

func (o *opencensusSpan) AddEvent(ctx context.Context, name string, attrs ...label.KeyValue) {
	o.AddEventWithTimestamp(ctx, time.Now(), name, attrs...)
}

func (o *opencensusSpan) AddEventWithTimestamp(ctx context.Context, timestamp time.Time, name string, attrs ...label.KeyValue) {
	// OpenCensus has "MessageEvents", which don't map well to opentelemetry events
	// There is also an "annotation", maybe we can use that?
	// TODO
}

func (o *opencensusSpan) IsRecording() bool {
	return o.span.IsRecordingEvents()
}

func (o *opencensusSpan) RecordError(ctx context.Context, err error, opts ...oteltrace.ErrorOption) {
	// OpenCensus doesn't seem have a concept of an error...
	// TODO: see if there is a concept we can map this to
}

func (o *opencensusSpan) SpanContext() oteltrace.SpanContext {
	var traceFlags byte
	if o.span.SpanContext().IsSampled() {
		traceFlags = oteltrace.FlagsSampled
	}
	return oteltrace.SpanContext{
		TraceID:    oteltrace.ID(o.span.SpanContext().TraceID),
		SpanID:     oteltrace.SpanID(o.span.SpanContext().SpanID),
		TraceFlags: traceFlags,
	}
}

func (o *opencensusSpan) SetStatus(code codes.Code, msg string) {
	o.span.SetStatus(octrace.Status{Code: int32(code), Message: msg})
}

func (o *opencensusSpan) SetName(name string) {
	o.span.SetName(name)
}

func (o *opencensusSpan) SetAttributes(kv ...label.KeyValue) {
	for _, pair := range kv {
		o.SetAttribute(string(pair.Key), pair.Value.AsInterface())
	}
}

func (o *opencensusSpan) SetAttribute(k string, v interface{}) {
	o.span.AddAttributes(toOCAttributes(k, v))
}

func toOCAttributes(k string, v interface{}) octrace.Attribute {
	var attr octrace.Attribute
	switch v.(type) {
	case int64:
		attr = octrace.Int64Attribute(k, v.(int64))
	case float64:
		attr = octrace.Float64Attribute(k, v.(float64))
	case bool:
		attr = octrace.BoolAttribute(k, v.(bool))
	case string:
		attr = octrace.StringAttribute(k, v.(string))
	default:
		// Not all opentelemetry types are supported by opencensus
		// int32, float32, uint32, uint64, array are not supported.
		attr = octrace.StringAttribute(k, "unsupported")
	}
	return attr
}
