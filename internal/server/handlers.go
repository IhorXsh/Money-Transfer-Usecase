package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/IhorXsh/Money-Transfer-Usecase/internal/contracts"
	"github.com/IhorXsh/Money-Transfer-Usecase/internal/domain"
	"github.com/IhorXsh/Money-Transfer-Usecase/internal/repository"
	"github.com/IhorXsh/Money-Transfer-Usecase/internal/usecases/transfer"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
)

type transferRequest struct {
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Amount        int64  `json:"amount"`
}

type transferResponse struct {
	Mutations []*contracts.Mutation `json:"mutations"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := http.StatusOK
	s.writeJSON(w, status, map[string]string{"status": "ok"})
	s.observeRequest(r.Context(), "/healthz", r.Method, status, start)
}

func (s *Server) transfer(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := http.StatusOK

	if r.Method != http.MethodPost {
		status = http.StatusMethodNotAllowed
		s.writeJSON(w, status, errorResponse{Error: "method not allowed"})
		s.observeRequest(r.Context(), "/transfer", r.Method, status, start)
		return
	}

	var in transferRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		status = http.StatusBadRequest
		s.writeJSON(w, status, errorResponse{Error: "invalid json"})
		s.logger.Warn("transfer request decode failed", "error", err, "trace_id", traceIDFromContext(r.Context()))
		s.observeRequest(r.Context(), "/transfer", r.Method, status, start)
		return
	}

	req := &transfer.TransferRequest{
		FromAccountID: domain.AccountID(in.FromAccountID),
		ToAccountID:   domain.AccountID(in.ToAccountID),
		Amount:        in.Amount,
	}

	s.mu.Lock()
	plan, err := s.uc.Execute(r.Context(), req)
	s.mu.Unlock()
	if err != nil {
		status = mapErrorToStatus(err)
		s.writeJSON(w, status, errorResponse{Error: err.Error()})
		s.logger.Warn(
			"transfer failed",
			"from_account_id", in.FromAccountID,
			"to_account_id", in.ToAccountID,
			"amount", in.Amount,
			"status", status,
			"error", err,
			"trace_id", traceIDFromContext(r.Context()),
		)
		s.observeRequest(r.Context(), "/transfer", r.Method, status, start)
		return
	}

	s.writeJSON(w, status, transferResponse{Mutations: plan.Mutations()})
	s.logger.Info(
		"transfer succeeded",
		"from_account_id", in.FromAccountID,
		"to_account_id", in.ToAccountID,
		"amount", in.Amount,
		"status", status,
		"duration_ms", time.Since(start).Milliseconds(),
		"trace_id", traceIDFromContext(r.Context()),
	)
	s.observeRequest(r.Context(), "/transfer", r.Method, status, start)
}

func (s *Server) observeRequest(ctx context.Context, endpoint, method string, status int, start time.Time) {
	statusLabel := http.StatusText(status)
	duration := time.Since(start).Seconds()
	durationObserver := s.m.requestDuration.WithLabelValues(endpoint, method, statusLabel)
	totalCounter := s.m.requestsTotal.WithLabelValues(endpoint, method, statusLabel)

	traceID := traceIDFromContext(ctx)
	if traceID != "" {
		exemplar := prometheus.Labels{"trace_id": traceID}
		if observer, ok := durationObserver.(prometheus.ExemplarObserver); ok {
			observer.ObserveWithExemplar(duration, exemplar)
		} else {
			durationObserver.Observe(duration)
		}
		if adder, ok := totalCounter.(prometheus.ExemplarAdder); ok {
			adder.AddWithExemplar(1, exemplar)
		} else {
			totalCounter.Inc()
		}
	} else {
		durationObserver.Observe(duration)
		totalCounter.Inc()
	}

	if status >= http.StatusBadRequest {
		errorCounter := s.m.requestErrors.WithLabelValues(endpoint, method, statusLabel)
		if traceID != "" {
			if adder, ok := errorCounter.(prometheus.ExemplarAdder); ok {
				adder.AddWithExemplar(1, prometheus.Labels{"trace_id": traceID})
				return
			}
		}
		errorCounter.Inc()
	}
}

func traceIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return ""
	}
	return spanCtx.TraceID().String()
}

func mapErrorToStatus(err error) int {
	switch {
	case errors.Is(err, transfer.ErrInvalidRequest),
		errors.Is(err, transfer.ErrInvalidAmount),
		errors.Is(err, transfer.ErrMissingAccount),
		errors.Is(err, transfer.ErrSameAccount),
		errors.Is(err, domain.ErrInsufficient),
		errors.Is(err, domain.ErrAccountInactive):
		return http.StatusBadRequest
	case errors.Is(err, repository.ErrAccountNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.logger.Error("response encode failed", "error", err)
	}
}
