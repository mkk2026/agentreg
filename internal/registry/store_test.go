package registry

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/mkk2026/agentreg/internal/agent"
)

func TestRegisterDefaultsAndGet(t *testing.T) {
	s, err := NewMemoryStore("")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Register(agent.Agent{Name: "search", Capabilities: []string{"search"}, Endpoint: "http://x"}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get("search")
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != agent.SourceLocal {
		t.Errorf("Source = %q, want %q", got.Source, agent.SourceLocal)
	}
	if got.Status != agent.StatusUnknown {
		t.Errorf("Status = %q, want %q", got.Status, agent.StatusUnknown)
	}
	if got.RegisteredAt.IsZero() {
		t.Error("RegisteredAt was not set")
	}
}

func TestFindByCapability(t *testing.T) {
	s, _ := NewMemoryStore("")
	_ = s.Register(agent.Agent{Name: "a", Capabilities: []string{"search", "db-read"}, Endpoint: "http://a"})
	_ = s.Register(agent.Agent{Name: "b", Capabilities: []string{"email"}, Endpoint: "http://b"})

	hit, _ := s.Find("search")
	if len(hit) != 1 || hit[0].Name != "a" {
		t.Fatalf("Find(search) = %+v, want [a]", hit)
	}
	if miss, _ := s.Find("nope"); len(miss) != 0 {
		t.Errorf("Find(nope) = %+v, want empty", miss)
	}
}

func TestGetNotFound(t *testing.T) {
	s, _ := NewMemoryStore("")
	if _, err := s.Get("ghost"); err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestUpdateStatus(t *testing.T) {
	s, _ := NewMemoryStore("")
	_ = s.Register(agent.Agent{Name: "a", Endpoint: "http://a"})
	now := time.Now().UTC()
	if err := s.UpdateStatus("a", agent.StatusHealthy, now); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get("a")
	if got.Status != agent.StatusHealthy {
		t.Errorf("Status = %q, want healthy", got.Status)
	}
	if !got.LastHeartbeat.Equal(now) {
		t.Errorf("LastHeartbeat = %v, want %v", got.LastHeartbeat, now)
	}
	if err := s.UpdateStatus("ghost", agent.StatusHealthy, now); err != ErrNotFound {
		t.Errorf("UpdateStatus(ghost) err = %v, want ErrNotFound", err)
	}
}

func TestReRegisterPreservesRegisteredAt(t *testing.T) {
	s, _ := NewMemoryStore("")
	_ = s.Register(agent.Agent{Name: "a", Endpoint: "http://a"})
	first, _ := s.Get("a")

	time.Sleep(2 * time.Millisecond)
	_ = s.Register(agent.Agent{Name: "a", Endpoint: "http://a2"})
	second, _ := s.Get("a")

	if !second.RegisteredAt.Equal(first.RegisteredAt) {
		t.Error("RegisteredAt changed on re-register")
	}
	if second.Endpoint != "http://a2" {
		t.Errorf("Endpoint = %q, want updated to http://a2", second.Endpoint)
	}
}

func TestPersistenceRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reg.json")
	s1, err := NewMemoryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = s1.Register(agent.Agent{Name: "search", Capabilities: []string{"search"}, Endpoint: "http://x"})

	s2, err := NewMemoryStore(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	got, err := s2.Get("search")
	if err != nil {
		t.Fatalf("reload Get: %v", err)
	}
	if got.Endpoint != "http://x" {
		t.Errorf("Endpoint = %q, want http://x", got.Endpoint)
	}
}

func TestLabelsNormalizeAndRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reg.json")
	s1, err := NewMemoryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	// with labels
	_ = s1.Register(agent.Agent{
		Name: "a", Endpoint: "http://a",
		Labels: map[string]string{"env": "prod", "team": "platform"},
	})
	// without labels -> must normalize to non-nil empty map
	_ = s1.Register(agent.Agent{Name: "b", Endpoint: "http://b"})
	if b, _ := s1.Get("b"); b.Labels == nil {
		t.Error("empty Labels should normalize to a non-nil map, got nil")
	}

	// labels survive a status update
	_ = s1.UpdateStatus("a", agent.StatusHealthy, time.Now().UTC())
	if a, _ := s1.Get("a"); a.Labels["env"] != "prod" {
		t.Errorf("UpdateStatus dropped labels: %+v", a.Labels)
	}

	// labels survive reload from disk (JSON persistence)
	s2, err := NewMemoryStore(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	a, err := s2.Get("a")
	if err != nil {
		t.Fatal(err)
	}
	if a.Labels["env"] != "prod" || a.Labels["team"] != "platform" {
		t.Errorf("labels did not round-trip through persistence: %+v", a.Labels)
	}
	if b, _ := s2.Get("b"); b.Labels == nil {
		t.Error("reloaded record with no labels should have non-nil map")
	}
}

func TestRemove(t *testing.T) {
	s, _ := NewMemoryStore("")
	_ = s.Register(agent.Agent{Name: "a", Endpoint: "http://a"})
	if err := s.Remove("a"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get("a"); err != ErrNotFound {
		t.Errorf("Get after remove err = %v, want ErrNotFound", err)
	}
	if err := s.Remove("a"); err != ErrNotFound {
		t.Errorf("double Remove err = %v, want ErrNotFound", err)
	}
}
