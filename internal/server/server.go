package server

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/IhorXsh/Money-Transfer-Usecase/internal/usecases/transfer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type metrics struct {
	requestDuration *prometheus.HistogramVec
	requestsTotal   *prometheus.CounterVec
	requestErrors   *prometheus.CounterVec
}

func newMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"endpoint", "method", "status"},
		),
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"endpoint", "method", "status"},
		),
		requestErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_request_errors_total",
				Help: "Total number of HTTP requests that finished with error status.",
			},
			[]string{"endpoint", "method", "status"},
		),
	}
	reg.MustRegister(m.requestDuration, m.requestsTotal, m.requestErrors)
	return m
}

type Server struct {
	logger *slog.Logger
	uc     *transfer.Interactor
	mu     sync.Mutex
	mux    *http.ServeMux
	m      *metrics
}

func New(logger *slog.Logger, uc *transfer.Interactor) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	s := &Server{
		logger: logger,
		uc:     uc,
		mux:    http.NewServeMux(),
		m:      newMetrics(prometheus.DefaultRegisterer),
	}

	s.mux.Handle("/healthz", otelhttp.NewHandler(http.HandlerFunc(s.healthz), "GET /healthz"))
	s.mux.Handle("/transfer", otelhttp.NewHandler(http.HandlerFunc(s.transfer), "POST /transfer"))

	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) MetricsHandler() http.Handler {
	return promhttp.Handler()
}
