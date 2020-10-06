# Migration TraceProvider

This implements the OpenTelemetry TraceProvider interface.  It wraps the input TraceProvider, but does context translation between OpenTelmetry and OpenCensus context.Context.