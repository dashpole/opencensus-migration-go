package oclibrary

import (
	"context"
	"log"
	"time"

	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
)

// ExampleKey is a key used for the example span
var ExampleKey = tag.MustNewKey("opencensuskey")

// ExportExampleSpan exports some spans using the opencensus go libraries.
func ExportExampleSpan(ctx context.Context) context.Context {
	ctx, err := tag.New(ctx,
		tag.Insert(ExampleKey, "opencensusvalue"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx, span := trace.StartSpan(ctx, "OpenCensusSpan")
	time.Sleep(time.Second)
	span.End()
	return ctx
}
