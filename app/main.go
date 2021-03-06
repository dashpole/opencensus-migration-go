package main

import (
	"context"
	"log"

	"github.com/dashpole/opencensus-migration-go/migration"
	oclibrary "github.com/dashpole/opencensus-migration-go/opencensus-library"
	otellibrary "github.com/dashpole/opencensus-migration-go/opentelemetry-library"

	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/trace"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/stdout"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	ctx := context.Background()

	log.Println("Registering opencensus log exporter...")
	// NewLogExporter also registers the exporter
	ocExporter, err := exporter.NewLogExporter(exporter.Options{})
	if err != nil {
		log.Fatal(err)
	}
	ocExporter.Start()
	defer ocExporter.Stop()
	defer ocExporter.Close()

	log.Println("Configuring opencensus...")
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	log.Println("Registering opentelemetry stdout exporter...")
	otExporter, err := stdout.NewExporter()
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(otExporter))
	global.SetTracerProvider(tp)

	log.Println("Emitting opencensus span.\n  It should have no parent, since it is the first span.")
	ctx = oclibrary.ExportExampleSpan(ctx)

	log.Println("Emitting opentelemetry span.\n-- It should have no parent, since the OpenCensus and OpenTelemetry span contexts are different.")
	ctx = otellibrary.ExportExampleSpan(ctx)

	log.Println("Resetting context...")
	ctx = context.Background()

	log.Println("Making OpenCensus libraries write traces using OpenTelemetry...")
	trace.DefaultTracer = migration.NewTracer()

	log.Println("Emitting opencensus span, which should be printed out using the OpenTelemetry stdout exporter.\n-- It should have no parent, since it is the first span.")
	ctx = oclibrary.ExportExampleSpan(ctx)

	log.Println("Emitting opentelemetry span\n-- It should have the OC span as a parent, since the OC span was written with using OpenTelemetry APIs.")
	ctx = otellibrary.ExportExampleSpan(ctx)

	log.Println("Emitting opencensus span, which should be printed out using the OpenTelemetry stdout exporter.\n-- It should have the OTel span as a parent, since it was written using OpenTelemetry APIs")
	ctx = oclibrary.ExportExampleSpan(ctx)
}
