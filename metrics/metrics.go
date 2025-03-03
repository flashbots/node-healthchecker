package metrics

import (
	"context"

	"go.opentelemetry.io/otel/exporters/prometheus"
	otelapi "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

const (
	metricsNamespace = "node-healthchecker"
)

var (
	meter otelapi.Meter
)

func Setup(ctx context.Context) error {
	for _, setup := range []func(context.Context) error{
		setupMeter, // must come first
		setupHealthchecksFlipCount,
		setupHealthchecksNokCount,
		setupHealthchecksOkCount,
		setupHealthchecksUp,
	} {
		if err := setup(ctx); err != nil {
			return err
		}
	}

	return nil
}

func setupMeter(ctx context.Context) error {
	res, err := resource.New(ctx)
	if err != nil {
		return err
	}

	exporter, err := prometheus.New(
		prometheus.WithNamespace(metricsNamespace),
		prometheus.WithoutScopeInfo(),
	)
	if err != nil {
		return err
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(res),
	)

	meter = provider.Meter(metricsNamespace)

	return nil
}

func setupHealthchecksFlipCount(ctx context.Context) error {
	m, err := meter.Int64Counter("healthcheck_flip_count",
		otelapi.WithDescription("count healthchecks that changed from ok to nok and vice versa"),
	)
	if err != nil {
		return err
	}
	HealthchecksFlipCount = m
	return nil
}

func setupHealthchecksNokCount(ctx context.Context) error {
	m, err := meter.Int64Counter("healthcheck_nok_count",
		otelapi.WithDescription("count of unsuccessful healthchecks"),
	)
	if err != nil {
		return err
	}
	HealthchecksNokCount = m
	return nil
}

func setupHealthchecksOkCount(ctx context.Context) error {
	m, err := meter.Int64Counter("healthcheck_ok_count",
		otelapi.WithDescription("count of successful healthchecks"),
	)
	if err != nil {
		return err
	}
	HealthchecksOkCount = m
	return nil
}

func setupHealthchecksUp(ctx context.Context) error {
	m, err := meter.Int64Gauge("healthcheck_up",
		otelapi.WithDescription("healthcheck status"),
	)
	if err != nil {
		return err
	}
	HealthcheckUp = m
	return nil
}
