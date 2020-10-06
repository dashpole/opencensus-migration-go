# Demo application

This application first creates an OpenCensus and OpenTelemetry span to demostrate the problem.  Notice that the OpenTelemetry span has the parent 0000000000000000.

Then, it installs the wrapped trace provider, and creates an OpenCensus, OpenTelemetry, and another OpenCensus span.

The first OpenCensus span has no parent.
The next OpenTelemetry span is a child of the first span.
The "extra" OpenCensus span is a child of the OpenTelemetry span.
The final OpenCensus span is a child of the "extra" span.

The TraceIDs should all be the same, and the "ParentSpanID" should match the SpanID of the previous span.

## Usage

```bash
go run main.go
```