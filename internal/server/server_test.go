package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/corebrim/agentreg/internal/agent"
	"github.com/corebrim/agentreg/internal/registry"
	"github.com/corebrim/agentreg/internal/verify"
)

// stubVerifier returns a fixed health result without any network I/O.
type stubVerifier struct{ healthy bool }

func (s stubVerifier) Verify(_ context.Context, _ agent.Agent) (verify.VerificationResult, error) {
	return verify.VerificationResult{Healthy: s.healthy, CheckedAt: time.Now().UTC(), Detail: "stub"}, nil
}

func newTestServer(t *testing.T, healthy bool) (*Server, *httptest.Server) {
	t.Helper()
	store, err := registry.NewMemoryStore("")
	if err != nil {
		t.Fatal(err)
	}
	srv := New(store, stubVerifier{healthy: healthy}, log.New(io.Discard, "", 0))
	return srv, httptest.NewServer(srv.Handler())
}

func post(t *testing.T, url string, v any) *http.Response {
	t.Helper()
	body, _ := json.Marshal(v)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestRegisterListFind(t *testing.T) {
	_, ts := newTestServer(t, true)
	defer ts.Close()

	resp := post(t, ts.URL+"/agents", agent.Agent{Name: "search", Capabilities: []string{"search"}, Endpoint: "http://x"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register status = %d, want 201", resp.StatusCode)
	}
	resp.Body.Close()

	var list []agent.Agent
	getJSON(t, ts.URL+"/agents", &list)
	if len(list) != 1 || list[0].Name != "search" {
		t.Fatalf("list = %+v, want one agent named search", list)
	}
	if list[0].Source != agent.SourceLocal {
		t.Errorf("Source = %q, want %q", list[0].Source, agent.SourceLocal)
	}

	var hit []agent.Agent
	getJSON(t, ts.URL+"/agents/find?capability=search", &hit)
	if len(hit) != 1 {
		t.Fatalf("find(search) = %+v, want one", hit)
	}

	var miss []agent.Agent
	getJSON(t, ts.URL+"/agents/find?capability=nope", &miss)
	if len(miss) != 0 {
		t.Errorf("find(nope) = %+v, want empty", miss)
	}
}

func TestRegisterValidation(t *testing.T) {
	_, ts := newTestServer(t, true)
	defer ts.Close()

	resp := post(t, ts.URL+"/agents", agent.Agent{Name: "x"}) // missing endpoint
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestFindMissingParam(t *testing.T) {
	_, ts := newTestServer(t, true)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/agents/find")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestHeartbeatEndpoint(t *testing.T) {
	_, ts := newTestServer(t, true)
	defer ts.Close()

	post(t, ts.URL+"/agents", agent.Agent{Name: "a", Endpoint: "http://x"}).Body.Close()

	resp := post(t, ts.URL+"/agents/a/heartbeat", nil)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("heartbeat status = %d, want 204", resp.StatusCode)
	}
	resp.Body.Close()

	resp = post(t, ts.URL+"/agents/ghost/heartbeat", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("heartbeat(ghost) status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

// TestVerifyAllUpdatesStatus exercises the heartbeat mechanism: a healthy pass
// marks agents healthy, and a later unhealthy pass flips them.
func TestVerifyAllUpdatesStatus(t *testing.T) {
	store, _ := registry.NewMemoryStore("")
	_ = store.Register(agent.Agent{Name: "a", Endpoint: "http://x"})

	New(store, stubVerifier{healthy: true}, log.New(io.Discard, "", 0)).verifyAll(context.Background())
	if got, _ := store.Get("a"); got.Status != agent.StatusHealthy {
		t.Errorf("after healthy pass Status = %q, want healthy", got.Status)
	}

	New(store, stubVerifier{healthy: false}, log.New(io.Discard, "", 0)).verifyAll(context.Background())
	if got, _ := store.Get("a"); got.Status != agent.StatusUnhealthy {
		t.Errorf("after unhealthy pass Status = %q, want unhealthy", got.Status)
	}
}

func getJSON(t *testing.T, url string, out any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatalf("decode %s: %v", url, err)
	}
}
