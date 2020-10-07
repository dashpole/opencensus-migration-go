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

In a fork of OpenCensus, we make it possible to swap out the implementation for StartSpan and StartSpanFromRemoteParent, mimicing the OpenTelemetry "Tracer" interface.

The /migration package contains an implementation of the new OpenCensus "Tracer" interface that uses OpenTelemetry.  This makes any libraries that use OpenCensus instead use OpenTelemetry, including exporters registered to OpenTelemetry.

### User Journey

Before updating _any_ libraries to use OpenTelemetry, the user adds the following line to their instantiation of OpenCensus:

```golang
trace.DefaultTracer = migration.NewTracer()
```

They must also instantiate the OpenTelemetry SDK and any exporters they want to use.  They can also remove any OpenCensus exporter registration, as the exporters won't be used for anything anymore.

After this point, they can migrate libraries to OpenTelemetry over time without any (significant) behavior changes.

### Shortcomings

1. There are some small differences between the OpenTelemetry and OpenCensus libraries. See comments in the migration library for details. This means that libraries which make use of some OpenCensus APIs may not work as expected. Note that these differences are relatively small.
2. This requires changes to OpenCensus.  In particular, it requires replacing a struct with an interface to make it possible to replace it.  This is not backwards compatible.  See https://github.com/census-instrumentation/opencensus-go/compare/master...dashpole:replaceable_sdk for the complete changes.
