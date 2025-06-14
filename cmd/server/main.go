package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeyotel"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"

	"calculator-otel/internal/app"
	"calculator-otel/internal/cache"
	"calculator-otel/internal/observability"
	"calculator-otel/internal/service"
)

const (
	appName        = "calculator-otel"
	appVersion     = "1.0.0"
	appEnvironment = "development"
)

func main() {
	ctx := context.Background()
	signCtx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	otelShutdown, err := observability.InitOpenTelemetry(
		ctx,
		appName,
		appVersion,
		appEnvironment,
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to initialize OpenTelemetry", "error", err)
	}

	defer otelShutdown(ctx)

	otelSlogHandler := otelslog.NewHandler(appName)
	logger := slog.New(otelSlogHandler)

	logger.InfoContext(ctx, "starting calculator server")
	defer logger.InfoContext(ctx, "shutting down calculator server")

	valkyClient, err := valkeyotel.NewClient(valkey.ClientOption{InitAddress: []string{"valkey:6379"}})
	if err != nil {
		logger.ErrorContext(ctx, "failed to create Valkey client", "error", err)
		return
	}

	cache := cache.New[int](valkyClient)

	service := service.New(logger, cache)

	tracer := otel.Tracer(appName)

	app := app.New(logger, service, tracer)
	mux := app.InitializeRoutes()

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
		BaseContext: func(net.Listener) context.Context {
			return signCtx
		},
	}

	go func() {
		<-signCtx.Done()
		logger.InfoContext(ctx, "received shutdown signal, shutting down server")
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown server gracefully", "error", err)
		}
	}()

	logger.InfoContext(ctx, "listening on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.ErrorContext(ctx, "failed to start server", "error", err)
		return
	}

	logger.InfoContext(ctx, "server shutdown complete")
}
