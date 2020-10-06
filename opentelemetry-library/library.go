package otellibrary

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/label"
)

// ExampleKey is a key used for the example span
var ExampleKey = label.Key("opentelemetrykey")

// ExportExampleSpan exports some spans using the opencensus go libraries.
func ExportExampleSpan(ctx context.Context) context.Context {
	tracer := global.Tracer("otelexample")
	ctx, span := tracer.Start(ctx, "OpenTelemetrySpan")
	span.SetAttributes(ExampleKey.String("otelvalue"))
	time.Sleep(time.Second)
	span.End()
	return ctx
}
