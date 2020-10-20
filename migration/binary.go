package migration

import (
	"context"

	"go.opencensus.io/trace/propagation"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/api/trace"
)

// traceContextKey is the same as opencensus:
// https://github.com/census-instrumentation/opencensus-go/blob/3fb168f674736c026e623310bfccb0691e6dec8a/plugin/ocgrpc/trace_common.go#L42
const traceContextKey = "grpc-trace-bin"

// Binary is an OpenTelemetry implementation of the OpenCensus grpc binary format.
// Needed because of https://github.com/open-telemetry/opentelemetry-specification/issues/437
type Binary struct{}

var _ otel.TextMapPropagator = Binary{}

// Inject injects context into the TextMapCarrier
func (b Binary) Inject(ctx context.Context, carrier otel.TextMapCarrier) {
	binaryContext := ctx.Value(traceContextKey)
	if state, ok := binaryContext.(string); binaryContext != nil && ok {
		carrier.Set(traceContextKey, state)
	}
	sc := trace.SpanFromContext(ctx).SpanContext()
	if !sc.IsValid() {
		return
	}
	h := propagation.Binary(otelSpanContextToOc(sc))
	carrier.Set(traceContextKey, string(h))
}

// Extract extracts the SpanContext from the TextMapCarrier
func (b Binary) Extract(ctx context.Context, carrier otel.TextMapCarrier) context.Context {
	state := carrier.Get(traceContextKey)
	if state != "" {
		ctx = context.WithValue(ctx, traceContextKey, state)
	}

	sc := b.extract(carrier)
	if !sc.IsValid() {
		return ctx
	}
	return trace.ContextWithRemoteSpanContext(ctx, sc)
}

func (b Binary) extract(carrier otel.TextMapCarrier) trace.SpanContext {
	h := carrier.Get(traceContextKey)
	if h == "" {
		return trace.SpanContext{}
	}
	ocContext, ok := propagation.FromBinary([]byte(h))
	if !ok {
		return trace.SpanContext{}
	}
	return ocSpanContextToOtel(ocContext)
}

// Fields returns the fields that this propagator modifies.
func (b Binary) Fields() []string {
	return []string{traceContextKey}
}
