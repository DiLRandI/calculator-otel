package app

import (
	"encoding/json"
	"net/http"

	"calculator-otel/internal/logger"
	"calculator-otel/internal/service"
)

type app struct {
	logger  logger.Logger
	service *service.Service
}

func New(logger logger.Logger, service *service.Service) *app {
	return &app{
		logger:  logger,
		service: service,
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
	req := &Request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var result int
	switch req.Operation {
	case service.OperandAdd:
		result = a.service.Add(r.Context(), req.Input1, req.Input2)
	case service.OperandSubtract:
		result = a.service.Subtract(r.Context(), req.Input1, req.Input2)
	case service.OperandMultiply:
		result = a.service.Multiply(r.Context(), req.Input1, req.Input2)
	case service.OperandDivide:
		var err error
		result, err = a.service.Divide(r.Context(), req.Input1, req.Input2)
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
		a.logger.ErrorContext(r.Context(), "failed to encode response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	a.logger.InfoContext(r.Context(), "calculation successful", "operation", req.Operation, "result", result)
}
