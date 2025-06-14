package service

import (
	"context"

	"calculator-otel/internal/logger"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	logger logger.Logger
}

func New(logger logger.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

func (s *Service) Add(ctx context.Context, a, b int) int {
	trace.SpanFromContext(ctx).AddEvent("Adding numbers", trace.WithAttributes(
		attribute.Float64("a", float64(a)),
		attribute.Float64("b", float64(b)),
		attribute.String("operation", "add"),
	))

	result := a + b

	trace.SpanFromContext(ctx).AddEvent("Addition result", trace.WithAttributes(
		attribute.Float64("result", float64(result)),
		attribute.String("operation", "add"),
	))

	return result
}

func (s *Service) Subtract(ctx context.Context, a, b int) int {
	trace.SpanFromContext(ctx).AddEvent("Subtracting numbers", trace.WithAttributes(
		attribute.Float64("a", float64(a)),
		attribute.Float64("b", float64(b)),
		attribute.String("operation", "subtract"),
	))

	result := a - b

	trace.SpanFromContext(ctx).AddEvent("Subtraction result", trace.WithAttributes(
		attribute.Float64("result", float64(result)),
		attribute.String("operation", "subtract"),
	))

	return result
}

func (s *Service) Multiply(ctx context.Context, a, b int) int {
	trace.SpanFromContext(ctx).AddEvent("Multiplying numbers", trace.WithAttributes(
		attribute.Float64("a", float64(a)),
		attribute.Float64("b", float64(b)),
		attribute.String("operation", "multiply"),
	))

	result := a * b

	trace.SpanFromContext(ctx).AddEvent("Multiplication result", trace.WithAttributes(
		attribute.Float64("result", float64(result)),
		attribute.String("operation", "multiply"),
	))

	return result
}

func (s *Service) Divide(ctx context.Context, a, b int) (int, error) {
	trace.SpanFromContext(ctx).AddEvent("Dividing numbers", trace.WithAttributes(
		attribute.Float64("a", float64(a)),
		attribute.Float64("b", float64(b)),
		attribute.String("operation", "divide"),
	))

	result := a / b

	trace.SpanFromContext(ctx).AddEvent("Division result", trace.WithAttributes(
		attribute.Float64("result", float64(result)),
		attribute.String("operation", "divide"),
	))

	return result, nil
}
