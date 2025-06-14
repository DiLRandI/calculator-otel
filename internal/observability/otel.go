package observability

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"
)

func InitOpenTelemetry(
	ctx context.Context,
	appName, version, environment string,
) (func(context.Context) error, error) {
	resource, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(appName),
			semconv.ServiceVersionKey.String(version),
			semconv.DeploymentEnvironmentName(environment),
		),
	)
	if err != nil {
		return nil, err
	}

	logExporter, err := otlploggrpc.New(ctx, otlploggrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}

	logProcessor := log.NewBatchProcessor(logExporter, log.WithExportInterval(5*time.Second))

	logProvider := log.NewLoggerProvider(
		log.WithResource(resource),
		log.WithProcessor(logProcessor),
	)

	global.SetLoggerProvider(logProvider)

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(resource),
	)

	otel.SetTracerProvider(traceProvider)

	// metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure()) // Use WithInsecure for local collector
	// if err != nil {
	// 	return nil, err
	// }

	// reader := metric.NewPeriodicReader(metricExporter, metric.WithInterval(30*time.Second))

	// mp := metric.NewMeterProvider(
	// 	metric.WithResource(resource),
	// 	metric.WithReader(reader),
	// )

	// otel.SetMeterProvider(mp)

	return func(ctx context.Context) error {
		shutdownErr := logProvider.Shutdown(ctx)
		if shutdownErr != nil {
			slog.Error("Error shutting down log provider", "error", shutdownErr)
		}

		shutdownErr = traceProvider.Shutdown(ctx)
		if shutdownErr != nil {
			slog.Error("Error shutting down trace provider", "error", shutdownErr)
		}

		// shutdownErr = mp.Shutdown(ctx)
		// if shutdownErr != nil {
		// 	slog.Error("Error shutting down metric provider", "error", shutdownErr)
		// }

		return shutdownErr
	}, nil
}
