package service

import (
	"context"

	"calculator-otel/internal/logger"
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
	s.logger.InfoContext(ctx, "Adding numbers", "a", a, "b", b)
	return a + b
}

func (s *Service) Subtract(ctx context.Context, a, b int) int {
	s.logger.InfoContext(ctx, "Subtracting numbers", "a", a, "b", b)
	return a - b
}

func (s *Service) Multiply(ctx context.Context, a, b int) int {
	s.logger.InfoContext(ctx, "Multiplying numbers", "a", a, "b", b)
	return a * b
}

func (s *Service) Divide(ctx context.Context, a, b int) (int, error) {
	s.logger.InfoContext(ctx, "Dividing numbers", "a", a, "b", b)
	return a / b, nil
}
