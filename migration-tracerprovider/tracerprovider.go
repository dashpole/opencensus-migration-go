package tracerprovider

import (
	"context"

	octrace "go.opencensus.io/trace"
	oteltrace "go.opentelemetry.io/otel/api/trace"
)

// NewTracerProvider returns a TracerProvider that wraps the input and helps with migration
func NewTracerProvider(provider oteltrace.TracerProvider) oteltrace.TracerProvider {
	return &wrappedTraceProvider{traceProvider: provider}
}

type wrappedTraceProvider struct {
	traceProvider oteltrace.TracerProvider
}

func (w *wrappedTraceProvider) Tracer(instrumentationName string, opts ...oteltrace.TracerOption) oteltrace.Tracer {
	return wrappedTracer{tracer: w.traceProvider.Tracer(instrumentationName, opts...)}
}

type wrappedTracer struct {
	tracer oteltrace.Tracer
}

func (w wrappedTracer) Start(ctx context.Context, spanName string, opts ...oteltrace.SpanOption) (context.Context, oteltrace.Span) {
	ctx = ocToOtel(ctx)
	ctx, span := w.tracer.Start(ctx, spanName, opts...)
	ctx = otelToOc(ctx)
	return ctx, span
}

func otelToOc(ctx context.Context) context.Context {
	otelSpan := oteltrace.SpanFromContext(ctx)
	otelSc := otelSpan.SpanContext()
	ocSc := octrace.SpanContext{
		TraceID:      octrace.TraceID(otelSc.TraceID),
		SpanID:       octrace.SpanID(otelSc.SpanID),
		TraceOptions: octrace.TraceOptions(otelSc.TraceFlags),
	}
	// We have to export an extra span to embed the OC SpanContext in the golang context
	ctx, span := octrace.StartSpanWithRemoteParent(ctx, "opencensus conversion", ocSc)
	span.End()
	return ctx
}

func ocToOtel(ctx context.Context) context.Context {
	ocSpan := octrace.FromContext(ctx)
	if ocSpan == nil {
		return ctx
	}
	ocSc := ocSpan.SpanContext()
	var traceFlags byte
	if ocSc.IsSampled() {
		traceFlags = oteltrace.FlagsSampled
	}
	otelSc := oteltrace.SpanContext{
		TraceID:    oteltrace.ID(ocSc.TraceID),
		SpanID:     oteltrace.SpanID(ocSc.SpanID),
		TraceFlags: traceFlags,
	}
	// make sure span context is zero'd out
	ctx = oteltrace.ContextWithSpan(ctx, nil)
	ctx = oteltrace.ContextWithRemoteSpanContext(ctx, otelSc)
	return ctx
}
