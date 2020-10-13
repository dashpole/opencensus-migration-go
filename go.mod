module github.com/dashpole/opencensus-migration-go

go 1.15

require (
	go.opencensus.io v0.22.4
	go.opentelemetry.io/otel v0.13.0
	go.opentelemetry.io/otel/exporters/stdout v0.13.0
	go.opentelemetry.io/otel/sdk v0.13.0
)

replace go.opencensus.io => github.com/dashpole/opencensus-go v0.22.5-0.20201013204313-bea70a30c868
