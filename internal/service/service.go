package service

import (
	"context"
	"fmt"

	"calculator-otel/internal/cache"
	"calculator-otel/internal/logger"
	"calculator-otel/internal/storage"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	logger  logger.Logger
	cache   cache.Cache[int]
	storage storage.Storage
}

func New(logger logger.Logger, cache cache.Cache[int], storage storage.Storage) *Service {
	return &Service{
		logger:  logger,
		cache:   cache,
		storage: storage,
	}
}

func (s *Service) Add(ctx context.Context, a, b int) int {
	trace.SpanFromContext(ctx).AddEvent("Adding numbers", trace.WithAttributes(
		attribute.Float64("a", float64(a)),
		attribute.Float64("b", float64(b)),
		attribute.String("operation", "add"),
	))

	result, err := s.cache.Get(ctx, createCacheKey(a, b, "add"))
	if err == nil {
		trace.SpanFromContext(ctx).AddEvent("Cache hit", trace.WithAttributes(
			attribute.String("key", createCacheKey(a, b, "add")),
			attribute.String("operation", "add"),
		))
		err = s.writeHistory(ctx, a, b, result, "add")
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to write history for add operation from cache", "error", err)
		}

		return result
	}

	result = a + b

	trace.SpanFromContext(ctx).AddEvent("Addition result", trace.WithAttributes(
		attribute.Float64("result", float64(result)),
		attribute.String("operation", "add"),
	))

	err = s.cache.Set(ctx, createCacheKey(a, b, "add"), result)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to set cache value", "error", err, "key", createCacheKey(a, b, "add"))
	}

	err = s.writeHistory(ctx, a, b, result, "add")
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to write history for add operation", "error", err)
	}

	return result
}

func (s *Service) Subtract(ctx context.Context, a, b int) int {
	trace.SpanFromContext(ctx).AddEvent("Subtracting numbers", trace.WithAttributes(
		attribute.Float64("a", float64(a)),
		attribute.Float64("b", float64(b)),
		attribute.String("operation", "subtract"),
	))

	result, err := s.cache.Get(ctx, createCacheKey(a, b, "subtract"))
	if err == nil {
		trace.SpanFromContext(ctx).AddEvent("Cache hit", trace.WithAttributes(
			attribute.String("key", createCacheKey(a, b, "subtract")),
			attribute.String("operation", "subtract"),
		))
		err = s.writeHistory(ctx, a, b, result, "subtract")
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to write history for subtract operation from cache", "error", err)
		}

		return result
	}

	result = a - b

	trace.SpanFromContext(ctx).AddEvent("Subtraction result", trace.WithAttributes(
		attribute.Float64("result", float64(result)),
		attribute.String("operation", "subtract"),
	))

	err = s.cache.Set(ctx, createCacheKey(a, b, "subtract"), result)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to set cache value", "error", err, "key", createCacheKey(a, b, "subtract"))
	}

	err = s.writeHistory(ctx, a, b, result, "subtract")
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to write history for subtract operation", "error", err)
	}

	return result
}

func (s *Service) Multiply(ctx context.Context, a, b int) int {
	trace.SpanFromContext(ctx).AddEvent("Multiplying numbers", trace.WithAttributes(
		attribute.Float64("a", float64(a)),
		attribute.Float64("b", float64(b)),
		attribute.String("operation", "multiply"),
	))

	result, err := s.cache.Get(ctx, createCacheKey(a, b, "multiply"))
	if err == nil {
		trace.SpanFromContext(ctx).AddEvent("Cache hit", trace.WithAttributes(
			attribute.String("key", createCacheKey(a, b, "multiply")),
			attribute.String("operation", "multiply"),
		))

		err = s.writeHistory(ctx, a, b, result, "multiply")
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to write history for multiply operation from cache", "error", err)
		}

		return result
	}

	result = a * b

	trace.SpanFromContext(ctx).AddEvent("Multiplication result", trace.WithAttributes(
		attribute.Float64("result", float64(result)),
		attribute.String("operation", "multiply"),
	))

	err = s.cache.Set(ctx, createCacheKey(a, b, "multiply"), result)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to set cache value", "error", err, "key", createCacheKey(a, b, "multiply"))
	}

	err = s.writeHistory(ctx, a, b, result, "multiply")
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to write history for multiply operation", "error", err)
	}

	return result
}

func (s *Service) Divide(ctx context.Context, a, b int) (int, error) {
	trace.SpanFromContext(ctx).AddEvent("Dividing numbers", trace.WithAttributes(
		attribute.Float64("a", float64(a)),
		attribute.Float64("b", float64(b)),
		attribute.String("operation", "divide"),
	))

	result, err := s.cache.Get(ctx, createCacheKey(a, b, "divide"))
	if err == nil {
		trace.SpanFromContext(ctx).AddEvent("Cache hit", trace.WithAttributes(
			attribute.String("key", createCacheKey(a, b, "divide")),
			attribute.String("operation", "divide"),
		))

		err = s.writeHistory(ctx, a, b, result, "divide")
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to write history for divide operation from cache", "error", err)
		}

		return result, nil
	}

	result = a / b

	trace.SpanFromContext(ctx).AddEvent("Division result", trace.WithAttributes(
		attribute.Float64("result", float64(result)),
		attribute.String("operation", "divide"),
	))

	err = s.cache.Set(ctx, createCacheKey(a, b, "divide"), result)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to set cache value", "error", err, "key", createCacheKey(a, b, "divide"))
	}

	err = s.writeHistory(ctx, a, b, result, "divide")
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to write history for divide operation", "error", err)
	}

	return result, nil
}

func createCacheKey(a, b int, operation string) string {
	return fmt.Sprintf("%d:%d:%s", a, b, operation)
}

func (s *Service) writeHistory(ctx context.Context, input1, input2, result int, operation string) error {
	if err := s.storage.Write(ctx, input1, input2, result, operation); err != nil {
		s.logger.ErrorContext(ctx, "failed to write history", "error", err, "input1", input1, "input2", input2, "result", result, "operation", operation)
		return fmt.Errorf("failed to write history: %w", err)
	}
	return nil
}

func (s *Service) GetHistory(ctx context.Context) ([]*storage.HistoryRecord, error) {
	trace.SpanFromContext(ctx).AddEvent("Retrieving history", trace.WithAttributes(
		attribute.String("operation", "get_history"),
	))

	history, err := s.storage.GetHistory(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get history", "error", err)
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	return history, nil
}
