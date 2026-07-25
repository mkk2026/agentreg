// Package registry holds the agent records and persists them.
package registry

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/mkk2026/agentreg/internal/agent"
)

// ErrNotFound is returned when an agent name is not registered.
var ErrNotFound = errors.New("agent not found")

// Store is the persistence seam. v1 ships MemoryStore; moving to a database
// later means implementing this interface and nothing else changes.
type Store interface {
	Register(a agent.Agent) error
	List() ([]agent.Agent, error)
	Find(capability string) ([]agent.Agent, error)
	Get(name string) (agent.Agent, error)
	UpdateStatus(name string, status agent.Status, checkedAt time.Time) error
	Remove(name string) error
}

// MemoryStore is an in-memory Store with optional JSON file persistence.
// All exported methods are safe for concurrent use.
type MemoryStore struct {
	mu     sync.RWMutex
	agents map[string]agent.Agent
	path   string // JSON persistence file; empty disables persistence
}

// NewMemoryStore creates a store. If path is non-empty, existing records are
// loaded from it and every mutation is persisted back atomically.
func NewMemoryStore(path string) (*MemoryStore, error) {
	s := &MemoryStore{agents: make(map[string]agent.Agent), path: path}
	if path != "" {
		if err := s.load(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *MemoryStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // first run, nothing to load
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}
	var list []agent.Agent
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	for _, a := range list {
		if a.Labels == nil {
			a.Labels = map[string]string{} // normalize records written before labels existed
		}
		s.agents[a.Name] = a
	}
	return nil
}

// persist writes the current state to disk atomically (temp file + rename).
// Callers already hold the write lock.
func (s *MemoryStore) persist() error {
	if s.path == "" {
		return nil
	}
	data, err := json.MarshalIndent(s.sortedLocked(), "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".agentreg-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, s.path)
}

// sortedLocked returns a name-sorted copy. Caller must hold at least a read lock.
func (s *MemoryStore) sortedLocked() []agent.Agent {
	list := make([]agent.Agent, 0, len(s.agents))
	for _, a := range s.agents {
		list = append(list, a)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}

// Register adds or replaces an agent. Blank Source/Status are defaulted, and the
// original RegisteredAt is preserved across re-registration.
func (s *MemoryStore) Register(a agent.Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if a.Source == "" {
		a.Source = agent.SourceLocal
	}
	if a.Status == "" {
		a.Status = agent.StatusUnknown
	}
	if a.Labels == nil {
		a.Labels = map[string]string{}
	}
	if existing, ok := s.agents[a.Name]; ok && !existing.RegisteredAt.IsZero() {
		a.RegisteredAt = existing.RegisteredAt
	}
	if a.RegisteredAt.IsZero() {
		a.RegisteredAt = time.Now().UTC()
	}

	s.agents[a.Name] = a
	return s.persist()
}

// List returns all agents, sorted by name.
func (s *MemoryStore) List() ([]agent.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sortedLocked(), nil
}

// Find returns all agents that advertise the given capability, sorted by name.
func (s *MemoryStore) Find(capability string) ([]agent.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []agent.Agent
	for _, a := range s.sortedLocked() {
		if a.HasCapability(capability) {
			out = append(out, a)
		}
	}
	return out, nil
}

// Get returns one agent by name, or ErrNotFound.
func (s *MemoryStore) Get(name string) (agent.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.agents[name]
	if !ok {
		return agent.Agent{}, ErrNotFound
	}
	return a, nil
}

// UpdateStatus sets an agent's health status and last-heartbeat time.
func (s *MemoryStore) UpdateStatus(name string, status agent.Status, checkedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.agents[name]
	if !ok {
		return ErrNotFound
	}
	a.Status = status
	a.LastHeartbeat = checkedAt
	s.agents[name] = a
	return s.persist()
}

// Remove deletes an agent by name.
func (s *MemoryStore) Remove(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.agents[name]; !ok {
		return ErrNotFound
	}
	delete(s.agents, name)
	return s.persist()
}
