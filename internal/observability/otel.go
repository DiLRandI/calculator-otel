package observability

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
)

func InitOpenTelemetry(
	ctx context.Context,
	appName, version, environment string,
) (func(context.Context) error, error) {
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(appName),
			semconv.ServiceVersionKey.String(version),
			semconv.DeploymentEnvironmentKey.String(environment),
		),
	)
	if err != nil {
		return nil, err
	}

	logExporter, err := otlploggrpc.New(ctx, otlploggrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}

	logProcessor := log.NewBatchProcessor(logExporter)

	logProvider := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(logProcessor),
	)

	global.SetLoggerProvider(logProvider)

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(traceProvider)

	return func(ctx context.Context) error {
		shutdownErr := logProvider.Shutdown(ctx)
		if shutdownErr != nil {
			slog.Error("Error shutting down log provider", "error", shutdownErr)
		}
		shutdownErr = traceProvider.Shutdown(ctx) // Shutdown trace provider as well
		if shutdownErr != nil {
			slog.Error("Error shutting down trace provider", "error", shutdownErr)
		}
		return shutdownErr
	}, nil
}
