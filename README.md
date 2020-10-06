# opencensus-migration-go

This uses an experimental TraceProvider to facilitate opencensus to opentelemetry migrations within a go program.

### The Problem

If you are creating the following spans:

```golang
ctx, ocSpan := opencensus.StartSpan(context.Background(), "OuterSpan")
defer ocSpan.End()
ctx, otSpan := opentelemetryTracer.Start(ctx, "MiddleSpan")
defer otSpan.End()
ctx, ocSpan := opencensus.StartSpan(ctx, "InnerSpan")
defer ocSpan.End()
```

You would have opencensus report:

```
[--------OuterSpan------------]
    [----InnerSpan------]
```

And opentelemetry would report:

```
   [-----MiddleSpan--------]
```

That is even if I send my traces to the same backend.  Instead, I would like:

```
[--------OuterSpan------------]
   [-----MiddleSpan--------]
    [----InnerSpan------]
```

### The attempted solution

In order to accomplish this, we must get opentelemetry and opencensus use the same context.Context key when creating spans.  This is impossible without forking one of those libraries.
In this experiment, we _somewhat_ accomplish this with a wrapper around opentelemetry's TraceProvider, found in the `migration-traceprovider` directory.  There are 3 steps:

1. In tracer.Start(), we first get the OpenCensus SpanContext, and overwrite the OpenTelemetry SpanContext with it.
1. Call the wrapped (i.e. "normal") tracer.Start() function.  It will use the OpenCensus span as its parent.
1. Get the OpenTelemetry SpanContext, and overwrite the OpenCensus SpanContext with it.

### Shortcomings

1. The OpenCensus library doesn't support setting SpanContext directly, and only exposes a StartSpanWithRemoteParent that allows specifying a SpanContext.  This means every time you create an OpenTelemetry span, you also get an OpenCensus span.  We might be able to remedy this by adding to the OpenCensus library.
2. When converting between OpenTelemetry and OpenCensus spans, we can only keep the SpanContext, not the other pieces, such as tags, attributes, tracestate, etc.  In our example above, it means the "InnerSpan" would no longer include tags from the "OuterSpan", since those were removed when creating the "MiddleSpan"


