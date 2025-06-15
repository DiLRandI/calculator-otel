package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"calculator-otel/internal/service"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	endpoint = "http://localhost"
)

type CalculationRequest struct {
	Input1    int    `json:"input1"`
	Input2    int    `json:"input2"`
	Operation string `json:"operation"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	tracer := otel.Tracer("calculator-otel/client")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/ping", nil)
	if err != nil {
		logger.Error("Failed to create ping request", "error", err)
		return
	}

	req.Header.Set("User-Agent", "CalculatorClient/1.0")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send ping request", "error", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Error("Unexpected status code", "status", resp.Status)
		return
	}
	logger.Info("Ping successful", "status", resp.Status)

	operators := []string{service.OperandAdd, service.OperandSubtract, service.OperandMultiply, service.OperandDivide}

	for i := 0; i < 100_000; i++ {
		go func(threadID int) {
			for {
				select {
				case <-ctx.Done():
					logger.Info("Thread exiting", "threadID", threadID)
					return
				default:
					spanCtx, span := tracer.Start(ctx, "sendCalculationRequest", trace.WithAttributes(
						attribute.String("threadID", strconv.Itoa(threadID)),
					), trace.WithSpanKind(trace.SpanKindClient))
					defer span.End()

					reqBody := CalculationRequest{
						Input1:    rand.Intn(100),
						Input2:    rand.Intn(100),
						Operation: operators[rand.Intn(len(operators))],
					}

					body, err := json.Marshal(reqBody)
					if err != nil {
						logger.Error("Failed to marshal request body", "error", err)
						continue
					}

					req, err := http.NewRequestWithContext(spanCtx, http.MethodPost, endpoint+"/calculate", nil)
					if err != nil {
						logger.Error("Failed to create request", "error", err)
						continue
					}
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("User-Agent", "CalculatorClient/1.0")
					req.Body = io.NopCloser(bytes.NewReader(body))

					resp, err := client.Do(req)
					if err != nil {
						logger.Error("Failed to send request", "error", err)
						continue
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						logger.Error("Unexpected status code", "status", resp.Status)
						continue
					}
					logger.Info("Calculation successful", "threadID", threadID, "status", resp.Status)

					time.Sleep(
						time.Duration(rand.Intn(2_000)) * time.Millisecond,
					) // Random sleep between 0 and 2 seconds
				}
			}
		}(i)
	}

	<-ctx.Done()
	logger.Info("Received shutdown signal, exiting...", "signal", ctx.Err())
	logger.Info("Client shutdown complete")
}
