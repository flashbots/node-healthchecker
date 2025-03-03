package metrics

import (
	otelapi "go.opentelemetry.io/otel/metric"
)

var (
	HealthchecksFlipCount otelapi.Int64Counter
	HealthchecksNokCount  otelapi.Int64Counter
	HealthchecksOkCount   otelapi.Int64Counter
	HealthcheckUp         otelapi.Int64Gauge
)
