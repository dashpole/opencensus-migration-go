# opencensus-migration-go

This uses a forked verison of OpenCensus to facilitate opencensus to opentelemetry migrations within a go program.

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

In order to accomplish this, we must get opentelemetry and opencensus use the same context.Context key when creating spans.  We can accomplish this by using a forked version of opencensus.  In the fork, we store and fetch SpanContext using OpenTelemetry's helper functions.

### User journey

Users add the following to their go.mod file:
```
replace go.opencensus.io => github.com/dashpole/opencensus-go v0.22.5-0.20201006212043-82f3a629fd85
```

### Shortcomings

1. When converting between OpenTelemetry and OpenCensus spans, we can only keep the SpanContext, not the other pieces, such as tags, attributes, tracestate, etc.  In our example above, it means the "InnerSpan" would no longer include tags from the "OuterSpan", since those were removed when creating the "MiddleSpan".
2. We would have to maintain a semi-permenant fork of opencensus, and rebase on any new changes.

