package app

import (
	"encoding/json"
	"net/http"

	"calculator-otel/internal/logger"
	"calculator-otel/internal/service"

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

	mux.Handle("GET /ping", http.HandlerFunc(a.pingHandler))
	mux.Handle("POST /ping", http.HandlerFunc(a.pingHandler))

	mux.Handle("POST /calculate", http.HandlerFunc(a.CalculateHandler))

	return mux
}

func (a *app) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func (a *app) CalculateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := a.tracer.Start(r.Context(), "CalculateHandler")
	defer span.End()

	req := &Request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var result int
	switch req.Operation {
	case service.OperandAdd:
		result = a.service.Add(ctx, req.Input1, req.Input2)
	case service.OperandSubtract:
		result = a.service.Subtract(ctx, req.Input1, req.Input2)
	case service.OperandMultiply:
		result = a.service.Multiply(ctx, req.Input1, req.Input2)
	case service.OperandDivide:
		var err error
		result, err = a.service.Divide(ctx, req.Input1, req.Input2)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
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
