// Package verify defines the trust seam of agentreg.
//
// v1 ships a single implementation (HealthVerifier). This interface is the seam
// the entire trust story grows from: future ANS-identity, prompt-injection, and
// MCP tools/list verifiers each implement Verify, and a trust engine collects
// their results. Treat this interface's shape as the most important design
// surface in the product, even though today it only carries a health probe.
package verify

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/corebrim/agentreg/internal/agent"
)

// VerificationResult is the outcome of a single verification pass.
type VerificationResult struct {
	Healthy   bool      `json:"healthy"`
	CheckedAt time.Time `json:"checked_at"`
	Detail    string    `json:"detail"`
}

// Verifier decides whether an agent is currently usable / trustworthy.
type Verifier interface {
	Verify(ctx context.Context, a agent.Agent) (VerificationResult, error)
}

// HealthVerifier verifies an agent by issuing an HTTP GET to its endpoint and
// treating any 2xx response as healthy.
type HealthVerifier struct {
	Client *http.Client
}

// NewHealthVerifier returns a HealthVerifier with the given per-probe timeout.
func NewHealthVerifier(timeout time.Duration) *HealthVerifier {
	return &HealthVerifier{Client: &http.Client{Timeout: timeout}}
}

// Verify implements Verifier.
//
// Design choice: an unreachable endpoint or non-2xx status is a *result*
// (Healthy=false), not a returned error. A non-nil error is reserved for cases
// where the check itself could not be formed (e.g. a malformed endpoint). This
// keeps the heartbeat loop simple: it reads res.Healthy and never has to
// distinguish "agent is down" from "verifier is broken."
func (h *HealthVerifier) Verify(ctx context.Context, a agent.Agent) (VerificationResult, error) {
	res := VerificationResult{CheckedAt: time.Now().UTC()}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.Endpoint, nil)
	if err != nil {
		res.Detail = fmt.Sprintf("bad endpoint: %v", err)
		return res, fmt.Errorf("build request for %q: %w", a.Name, err)
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		res.Detail = fmt.Sprintf("unreachable: %v", err)
		return res, nil
	}
	defer resp.Body.Close()

	res.Detail = fmt.Sprintf("HTTP %d", resp.StatusCode)
	res.Healthy = resp.StatusCode >= 200 && resp.StatusCode < 300
	return res, nil
}
