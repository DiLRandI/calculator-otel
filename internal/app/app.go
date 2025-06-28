package app

import (
	"encoding/json"
	"net/http"

	"calculator-otel/internal/logger"
	"calculator-otel/internal/service"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

type app struct {
	logger  logger.Logger
	service *service.Service
	tracer  trace.Tracer
}

func New(logger logger.Logger, service *service.Service, tracer trace.Tracer) *app {
	return &app{
		logger:  logger,
		service: service,
		tracer:  tracer,
	}
}

func (a *app) InitializeRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /ping", otelhttp.NewHandler(http.HandlerFunc(a.pingHandler), "PingHandler"))
	mux.Handle("POST /ping", otelhttp.NewHandler(http.HandlerFunc(a.pingHandler), "PingHandler"))

	mux.Handle("POST /calculate", otelhttp.NewHandler(http.HandlerFunc(a.CalculateHandler), "CalculateHandler"))
	mux.Handle("GET /history", otelhttp.NewHandler(http.HandlerFunc(a.HistoryHandler), "HistoryHandler"))

	return mux
}

func (a *app) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func (a *app) CalculateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := &Request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var result int
	switch req.Operation {
	case service.OperandAdd:
		a.logger.InfoContext(ctx, "performing addition", "input1", req.Input1, "input2", req.Input2)
		result = a.service.Add(ctx, req.Input1, req.Input2)
	case service.OperandSubtract:
		a.logger.InfoContext(ctx, "performing subtraction", "input1", req.Input1, "input2", req.Input2)
		result = a.service.Subtract(ctx, req.Input1, req.Input2)
	case service.OperandMultiply:
		a.logger.InfoContext(ctx, "performing multiplication", "input1", req.Input1, "input2", req.Input2)
		result = a.service.Multiply(ctx, req.Input1, req.Input2)
	case service.OperandDivide:
		a.logger.InfoContext(ctx, "performing division", "input1", req.Input1, "input2", req.Input2)
		var err error
		result, err = a.service.Divide(ctx, req.Input1, req.Input2)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		a.logger.ErrorContext(ctx, "invalid operation", "operation", req.Operation)
		http.Error(w, "Invalid operation", http.StatusBadRequest)
		return
	}
	response := Response{
		Result: result,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		a.logger.ErrorContext(ctx, "failed to encode response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	a.logger.InfoContext(ctx, "calculation successful", "operation", req.Operation, "result", result)
}

func (a *app) HistoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	history, err := a.service.GetHistory(ctx)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to get history", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(history) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := json.NewEncoder(w).Encode(history); err != nil {
		a.logger.ErrorContext(ctx, "failed to encode history response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
