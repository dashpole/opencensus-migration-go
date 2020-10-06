# Demo application

This application is just a "normal" application using an opencensus and opentelemetry library.  However, because we are using a forked version of opencensus, which stores and retrieves context using opentelemetry's context, the spans are connected as they should be.

The first OpenCensus span has no parent.
The next OpenTelemetry span is a child of the first span.
The final OpenCensus span is a child of the second span.

The TraceIDs should all be the same, and the "ParentSpanID" should match the SpanID of the previous span.

## Usage

```bash
go run main.go
```