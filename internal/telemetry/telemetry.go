// Package telemetry configures the application's OpenTelemetry tracing
// pipeline.
//
// The exporter is selected entirely from the environment via the standard
// OTEL_* variables, so the application never hardcodes a destination:
//
//   - OTEL_TRACES_EXPORTER selects the exporter ("otlp" by default, "console"
//     to print spans locally, or "none" to disable exporting).
//   - OTEL_EXPORTER_OTLP_ENDPOINT / OTEL_EXPORTER_OTLP_PROTOCOL point the OTLP
//     exporter at a collector/agent (e.g. Grafana Alloy) that forwards traces
//     to a centralized backend such as Tempo.
//   - OTEL_SERVICE_NAME / OTEL_RESOURCE_ATTRIBUTES describe this service.
//   - OTEL_TRACES_SAMPLER / OTEL_TRACES_SAMPLER_ARG control sampling (defaults
//     to always-on to match the service's low traffic).
package telemetry

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// defaultServiceName is used for the Tempo/LGTM service graph when
// OTEL_SERVICE_NAME has not been provided.
const defaultServiceName = "no-time-to-explain"

// Init configures the global OpenTelemetry tracer provider and text-map
// propagator. It returns a shutdown function that flushes any buffered spans;
// callers should invoke it during graceful shutdown.
func Init(ctx context.Context) (func(context.Context) error, error) {
	// Ensure a service name is always present so traces are attributable in the
	// backend, without overriding an operator-provided value.
	if os.Getenv("OTEL_SERVICE_NAME") == "" {
		_ = os.Setenv("OTEL_SERVICE_NAME", defaultServiceName)
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, fmt.Errorf("building otel resource: %w", err)
	}

	// autoexport builds the span exporter from OTEL_TRACES_EXPORTER (otlp by
	// default; console and none also supported) and honours the standard
	// OTEL_EXPORTER_OTLP_* env vars for endpoint, protocol, and headers.
	exporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("building otel span exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(samplerFromEnv()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

// samplerFromEnv resolves the trace sampler from the standard
// OTEL_TRACES_SAMPLER / OTEL_TRACES_SAMPLER_ARG environment variables,
// defaulting to a parent-based always-on sampler (equivalent to a 1.0 rate).
func samplerFromEnv() sdktrace.Sampler {
	switch os.Getenv("OTEL_TRACES_SAMPLER") {
	case "always_off":
		return sdktrace.NeverSample()
	case "always_on":
		return sdktrace.AlwaysSample()
	case "traceidratio":
		return sdktrace.TraceIDRatioBased(samplerArg(1.0))
	case "parentbased_always_off":
		return sdktrace.ParentBased(sdktrace.NeverSample())
	case "parentbased_traceidratio":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(samplerArg(1.0)))
	case "parentbased_always_on", "":
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	default:
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	}
}

// samplerArg parses OTEL_TRACES_SAMPLER_ARG as a float, falling back to def.
func samplerArg(def float64) float64 {
	if v := os.Getenv("OTEL_TRACES_SAMPLER_ARG"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}
