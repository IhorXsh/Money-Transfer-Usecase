package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/IhorXsh/Money-Transfer-Usecase/internal/domain"
	"github.com/IhorXsh/Money-Transfer-Usecase/internal/repository"
	"github.com/IhorXsh/Money-Transfer-Usecase/internal/usecases/transfer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()

	oldRegisterer := prometheus.DefaultRegisterer
	oldGatherer := prometheus.DefaultGatherer
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg
	t.Cleanup(func() {
		prometheus.DefaultRegisterer = oldRegisterer
		prometheus.DefaultGatherer = oldGatherer
	})

	accounts := map[domain.AccountID]*domain.Account{
		"a1": domain.NewAccount("a1", 100, domain.AccountStatusActive),
		"a2": domain.NewAccount("a2", 50, domain.AccountStatusActive),
		"a3": domain.NewAccount("a3", 10, domain.AccountStatusActive),
	}
	repo := repository.NewAccountRepo(accounts)
	uc := transfer.NewInteractor(repo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return New(logger, uc)
}

func doRequest(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(data))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealthz(t *testing.T) {
	srv := newTestServer(t)

	rec := doRequest(t, srv.Handler(), http.MethodGet, "/healthz", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	require.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}

func TestTransferSuccess(t *testing.T) {
	srv := newTestServer(t)

	rec := doRequest(t, srv.Handler(), http.MethodPost, "/transfer", map[string]any{
		"from_account_id": "a1",
		"to_account_id":   "a2",
		"amount":          25,
	})

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	var resp transferResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Mutations, 2)
	require.Equal(t, "a1", resp.Mutations[0].ID)
	require.Equal(t, "a2", resp.Mutations[1].ID)
}

func TestTransferErrors(t *testing.T) {
	srv := newTestServer(t)

	tests := []struct {
		name       string
		method     string
		body       any
		rawBody    string
		wantStatus int
	}{
		{
			name:       "method not allowed",
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "invalid json",
			method:     http.MethodPost,
			rawBody:    "{invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "insufficient funds",
			method: http.MethodPost,
			body: map[string]any{
				"from_account_id": "a3",
				"to_account_id":   "a2",
				"amount":          1000,
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "account not found",
			method: http.MethodPost,
			body: map[string]any{
				"from_account_id": "missing",
				"to_account_id":   "a2",
				"amount":          10,
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var rec *httptest.ResponseRecorder
			if tc.rawBody != "" {
				req := httptest.NewRequest(tc.method, "/transfer", strings.NewReader(tc.rawBody))
				req.Header.Set("Content-Type", "application/json")
				rec = httptest.NewRecorder()
				srv.Handler().ServeHTTP(rec, req)
			} else {
				rec = doRequest(t, srv.Handler(), tc.method, "/transfer", tc.body)
			}

			require.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestMetricsSmoke(t *testing.T) {
	srv := newTestServer(t)

	_ = doRequest(t, srv.Handler(), http.MethodPost, "/transfer", map[string]any{
		"from_account_id": "a1",
		"to_account_id":   "a2",
		"amount":          10,
	})
	_ = doRequest(t, srv.Handler(), http.MethodPost, "/transfer", map[string]any{
		"from_account_id": "a3",
		"to_account_id":   "a2",
		"amount":          1000,
	})

	rec := doRequest(t, srv.MetricsHandler(), http.MethodGet, "/metrics", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "http_request_duration_seconds")
	require.Contains(t, body, "http_requests_total")
	require.Contains(t, body, "http_request_errors_total")
	require.Contains(t, body, `endpoint="/transfer"`)
}
