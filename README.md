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

This entirely replaces the OpenTelemetry SDK by re-implementing it using OpenCensus under the hood.

### User experience

Users that are using entirely OpenCensus can begin switching libraries to OpenTelemetry without changing behavior at all by registering the migration OpenTelemetry TraceProvider:
```golang
global.SetTracerProvider(migration.NewTracerProvider())
```

All OpenTelemetry libraries will use this SDK, which convert spans to OpenCensus.

Once they have migrated all libraries to OpenTelemetry, they remove the migration TraceProvider, and switch to using OpenTelemetry exporters.

### Shortcomings

1. There are some small differences between the OpenTelemetry and OpenCensus libraries.  See comments in the traceprovider for details.  This means that libraries which make use of some OpenTelemetry APIs may not work as expected.  Note that this is _much less bad_ than other options being considered.
2. Even when libraries are using OpenTelemetry, spans are still sent using OpenCensus exporters.  This isn't _neccessarily_ a downside, since users already using OpenCensus presumably had exporters working correctly.  However, it means users won't be able to make use of OpenTelemetry exporters until they have completed the migration, and removed the migration TraceProvider.
