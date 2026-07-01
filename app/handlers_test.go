package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newTestApp() *application {
	return &application{
		db: nil,
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "test_requests_total",
		}, []string{"method", "path", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "test_duration_seconds",
		}, []string{"method", "path"}),
	}
}

func TestHealth(t *testing.T) {
	app := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	app.handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]string
	json.NewDecoder(rr.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %q", body["status"])
	}
}
