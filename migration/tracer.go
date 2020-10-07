package migration

import (
	"context"

	octrace "go.opencensus.io/trace"
	"go.opentelemetry.io/otel/api/global"
	oteltrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/codes"
)

func NewTracer() octrace.Tracer {
	return &otelTracer{tracer: global.Tracer("ocmigration")}
}

type otelTracer struct {
	tracer oteltrace.Tracer
}

func (o *otelTracer) StartSpan(ctx context.Context, name string, s ...octrace.StartOption) (context.Context, octrace.Span) {
	// TODO: use start options
	ctx, sp := o.tracer.Start(ctx, name)
	return ctx, &span{otSpan: sp}
}
func (o *otelTracer) StartSpanWithRemoteParent(ctx context.Context, name string, parent octrace.SpanContext, s ...octrace.StartOption) (context.Context, octrace.Span) {
	// make sure span context is zero'd out so we use the remote parent
	ctx = oteltrace.ContextWithSpan(ctx, nil)
	ctx = oteltrace.ContextWithRemoteSpanContext(ctx, ocSpanContextToOtel(parent))
	return o.StartSpan(ctx, name, s...)
}

type span struct {
	otSpan oteltrace.Span
}

func (s *span) IsRecordingEvents() bool {
	return s.otSpan.IsRecording()
}

func (s *span) End() {
	s.otSpan.End()
}

func (s *span) SpanContext() octrace.SpanContext {
	return otelSpanContextToOc(s.otSpan.SpanContext())
}

func (s *span) SetName(name string) {
	s.otSpan.SetName(name)
}

func (s *span) SetStatus(status octrace.Status) {
	s.otSpan.SetStatus(codes.Code(status.Code), status.Message)
}

func (s *span) AddAttributes(attributes ...octrace.Attribute) {
	// TODO: implement with s.otSpan.SetAttributes
}

func (s *span) Annotate(attributes []octrace.Attribute, str string) {
	// TODO implement
}

func (s *span) Annotatef(attributes []octrace.Attribute, format string, a ...interface{}) {
	// TODO implement
}

func (s *span) AddMessageSendEvent(messageID, uncompressedByteSize, compressedByteSize int64) {
	// TODO implement with s.otSpan.AddEvent()
}

func (s *span) AddMessageReceiveEvent(messageID, uncompressedByteSize, compressedByteSize int64) {
	// TODO implement with s.otSpan.AddEvent()
}

func (s *span) AddLink(l octrace.Link) {
	// TODO Will this work?
}

func (s *span) String() string {
	return "TODO"
}

func (s *span) AddChild() {
	// Not needed.
}

func (s *span) SpanData() *octrace.SpanData {
	return &octrace.SpanData{}
}

func otelSpanContextToOc(sc oteltrace.SpanContext) octrace.SpanContext {
	return octrace.SpanContext{
		TraceID:      octrace.TraceID(sc.TraceID),
		SpanID:       octrace.SpanID(sc.SpanID),
		TraceOptions: octrace.TraceOptions(sc.TraceFlags), // TODO I dont this this actually works...
	}
}

func ocSpanContextToOtel(sc octrace.SpanContext) oteltrace.SpanContext {
	var traceFlags byte
	if sc.IsSampled() {
		traceFlags = oteltrace.FlagsSampled
	}
	return oteltrace.SpanContext{
		TraceID:    oteltrace.ID(sc.TraceID),
		SpanID:     oteltrace.SpanID(sc.SpanID),
		TraceFlags: traceFlags,
	}
}
