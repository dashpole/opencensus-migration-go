package migration

import (
	"context"
	"fmt"
	"log"

	octrace "go.opencensus.io/trace"
	"go.opentelemetry.io/otel/api/global"
	otel "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/label"
)

// NewTracer returns an implementation of the OpenCensus Tracer interface which
// uses OpenTelemetry APIs.  Using this implementation of Tracer "upgrades"
// libraries that use OpenCensus to OpenTelemetry to facilitate a migration.
func NewTracer() octrace.Tracer {
	return &otelTracer{tracer: global.Tracer("ocmigration")}
}

type otelTracer struct {
	tracer otel.Tracer
}

func (o *otelTracer) StartSpan(ctx context.Context, name string, s ...octrace.StartOption) (context.Context, octrace.Span) {
	ctx, sp := o.tracer.Start(ctx, name, convertStartOptions(s, name)...)
	return ctx, &span{otSpan: sp}
}

func convertStartOptions(optFns []octrace.StartOption, name string) []otel.SpanOption {
	var ocOpts octrace.StartOptions
	for _, fn := range optFns {
		fn(&ocOpts)
	}
	otOpts := []otel.SpanOption{}
	switch ocOpts.SpanKind {
	case octrace.SpanKindClient:
		otOpts = append(otOpts, otel.WithSpanKind(otel.SpanKindClient))
	case octrace.SpanKindServer:
		otOpts = append(otOpts, otel.WithSpanKind(otel.SpanKindServer))
	case octrace.SpanKindUnspecified:
		otOpts = append(otOpts, otel.WithSpanKind(otel.SpanKindUnspecified))
	}

	if ocOpts.Sampler != nil {
		// OTel doesn't allow setting a sampler in SpanOptions
		log.Printf("Ignoring custom sampler for span %q in OpenCensus -> OpenTelemetry migration sdk.\n", name)
	}
	return otOpts
}

func (o *otelTracer) StartSpanWithRemoteParent(ctx context.Context, name string, parent octrace.SpanContext, s ...octrace.StartOption) (context.Context, octrace.Span) {
	// make sure span context is zero'd out so we use the remote parent
	ctx = otel.ContextWithSpan(ctx, nil)
	ctx = otel.ContextWithRemoteSpanContext(ctx, ocSpanContextToOtel(parent))
	return o.StartSpan(ctx, name, s...)
}

func (o *otelTracer) FromContext(ctx context.Context) octrace.Span {
	otSpan := otel.SpanFromContext(ctx)
	return &span{otSpan: otSpan}
}

func (o *otelTracer) NewContext(parent context.Context, s octrace.Span) context.Context {
	if otSpan, ok := s.(*span); ok {
		return otel.ContextWithSpan(parent, otSpan.otSpan)
	}
	// The user must have created the octrace Span using a different tracer, and we don't know how to store it.
	log.Printf("Unable to create context with span %q, since it was created using a different tracer.\n", s.String())
	return parent
}

type span struct {
	// We can't implement the unexported functions, so add the interface here.
	octrace.Span
	otSpan otel.Span
}

func (s *span) IsRecordingEvents() bool {
	return s.otSpan.IsRecording()
}

func (s *span) End() {
	s.otSpan.End()
}

func (s *span) SpanContext() octrace.SpanContext {
	return s.otelSpanContextToOc(s.otSpan.SpanContext())
}

func (s *span) SetName(name string) {
	s.otSpan.SetName(name)
}

func (s *span) SetStatus(status octrace.Status) {
	s.otSpan.SetStatus(codes.Code(status.Code), status.Message)
}

func (s *span) AddAttributes(attributes ...octrace.Attribute) {
	s.otSpan.SetAttributes(convertAttributes(attributes)...)
}

func convertAttributes(attributes []octrace.Attribute) []label.KeyValue {
	otAttributes := make([]label.KeyValue, len(attributes))
	for i, a := range attributes {
		otAttributes[i] = label.KeyValue{
			Key:   label.Key(a.Key()),
			Value: convertValue(a.Value()),
		}
	}
	return otAttributes
}

func convertValue(ocval interface{}) label.Value {
	switch v := ocval.(type) {
	case bool:
		return label.BoolValue(v)
	case int64:
		return label.Int64Value(v)
	case float64:
		return label.Float64Value(v)
	case string:
		return label.StringValue(v)
	default:
		return label.StringValue("unknown")
	}
}

func (s *span) Annotate(attributes []octrace.Attribute, str string) {
	s.otSpan.AddEvent(context.Background(), str, convertAttributes(attributes)...)
}

func (s *span) Annotatef(attributes []octrace.Attribute, format string, a ...interface{}) {
	s.Annotate(attributes, fmt.Sprintf(format, a...))
}

var (
	uncompressedKey = label.Key("uncompressed byte size")
	compressedKey   = label.Key("compressed byte size")
)

func (s *span) AddMessageSendEvent(messageID, uncompressedByteSize, compressedByteSize int64) {
	s.otSpan.AddEvent(context.Background(), "message send",
		label.KeyValue{
			Key:   uncompressedKey,
			Value: label.Int64Value(uncompressedByteSize),
		},
		label.KeyValue{
			Key:   compressedKey,
			Value: label.Int64Value(compressedByteSize),
		})
}

func (s *span) AddMessageReceiveEvent(messageID, uncompressedByteSize, compressedByteSize int64) {
	s.otSpan.AddEvent(context.Background(), "message receive",
		label.KeyValue{
			Key:   uncompressedKey,
			Value: label.Int64Value(uncompressedByteSize),
		},
		label.KeyValue{
			Key:   compressedKey,
			Value: label.Int64Value(compressedByteSize),
		})
}

func (s *span) AddLink(l octrace.Link) {
	// Links may only be specified at creation time
	log.Printf("Unable to add a link for span %q, since OpenTelemetry doesn't support adding links after creation.\n", s.String())
}

func (s *span) String() string {
	return fmt.Sprintf("span %s",
		s.otSpan.SpanContext().SpanID.String())
}

func (s *span) otelSpanContextToOc(sc otel.SpanContext) octrace.SpanContext {
	if sc.IsDebug() || sc.IsDeferred() {
		// OTel don't support these options
		log.Printf("Ignoring OpenCensus Debug or Deferred trace flags for span %q because they are not supported by OpenTelemetry.\n", s.String())
	}
	var to octrace.TraceOptions
	if sc.IsSampled() {
		// OpenCensus doesn't expose functions to directly set sampled
		to = 0x1
	}
	return octrace.SpanContext{
		TraceID:      octrace.TraceID(sc.TraceID),
		SpanID:       octrace.SpanID(sc.SpanID),
		TraceOptions: to,
	}
}

func ocSpanContextToOtel(sc octrace.SpanContext) otel.SpanContext {
	var traceFlags byte
	if sc.IsSampled() {
		traceFlags = otel.FlagsSampled
	}
	return otel.SpanContext{
		TraceID:    otel.ID(sc.TraceID),
		SpanID:     otel.SpanID(sc.SpanID),
		TraceFlags: traceFlags,
	}
}
