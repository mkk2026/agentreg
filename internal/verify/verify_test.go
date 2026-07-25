package verify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mkk2026/agentreg/internal/agent"
)

func TestHealthVerifierHealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	res, err := NewHealthVerifier(2 * time.Second).Verify(context.Background(),
		agent.Agent{Name: "a", Endpoint: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Healthy {
		t.Errorf("expected healthy, got %+v", res)
	}
}

func TestHealthVerifierUnhealthyStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	res, err := NewHealthVerifier(2 * time.Second).Verify(context.Background(),
		agent.Agent{Name: "a", Endpoint: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if res.Healthy {
		t.Errorf("expected unhealthy for HTTP 500, got %+v", res)
	}
}

func TestHealthVerifierUnreachableIsResultNotError(t *testing.T) {
	// Port 1 is not listening; the probe should fail as a result, not an error.
	res, err := NewHealthVerifier(200*time.Millisecond).Verify(context.Background(),
		agent.Agent{Name: "a", Endpoint: "http://127.0.0.1:1"})
	if err != nil {
		t.Fatalf("unreachable should not return an error, got %v", err)
	}
	if res.Healthy {
		t.Errorf("expected unhealthy for unreachable endpoint, got %+v", res)
	}
}

func TestHealthVerifierBadEndpointErrors(t *testing.T) {
	_, err := NewHealthVerifier(time.Second).Verify(context.Background(),
		agent.Agent{Name: "a", Endpoint: "://malformed"})
	if err == nil {
		t.Error("expected an error for a malformed endpoint")
	}
}
