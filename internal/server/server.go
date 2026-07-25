// Package server exposes the registry over HTTP and runs the heartbeat loop.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/mkk2026/agentreg/internal/agent"
	"github.com/mkk2026/agentreg/internal/registry"
	"github.com/mkk2026/agentreg/internal/verify"
)

// Server wires the store, verifier, and HTTP handlers together.
type Server struct {
	store    registry.Store
	verifier verify.Verifier
	logger   *log.Logger
}

// New creates a Server.
func New(store registry.Store, verifier verify.Verifier, logger *log.Logger) *Server {
	return &Server{store: store, verifier: verifier, logger: logger}
}

// Handler returns the registry HTTP API.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /agents", s.handleRegister)
	mux.HandleFunc("GET /agents", s.handleList)
	mux.HandleFunc("GET /agents/find", s.handleFind)
	mux.HandleFunc("POST /agents/{name}/heartbeat", s.handleHeartbeat)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var a agent.Agent
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if a.Name == "" || a.Endpoint == "" {
		writeError(w, http.StatusBadRequest, "name and endpoint are required")
		return
	}
	if err := s.store.Register(a); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	stored, err := s.store.Get(a.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, stored)
}

func (s *Server) handleList(w http.ResponseWriter, _ *http.Request) {
	list, err := s.store.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleFind(w http.ResponseWriter, r *http.Request) {
	capability := r.URL.Query().Get("capability")
	if capability == "" {
		writeError(w, http.StatusBadRequest, "capability query parameter is required")
		return
	}
	list, err := s.store.Find(capability)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.store.UpdateStatus(name, agent.StatusHealthy, time.Now().UTC()); err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			writeError(w, http.StatusNotFound, "agent not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// StartHeartbeat verifies every agent on an interval until ctx is cancelled.
// It runs one pass immediately so status is fresh right after boot.
func (s *Server) StartHeartbeat(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.verifyAll(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.verifyAll(ctx)
		}
	}
}

func (s *Server) verifyAll(ctx context.Context) {
	list, err := s.store.List()
	if err != nil {
		s.logger.Printf("heartbeat: list failed: %v", err)
		return
	}
	for _, a := range list {
		res, err := s.verifier.Verify(ctx, a)
		if err != nil {
			s.logger.Printf("heartbeat: verify %q errored: %v", a.Name, err)
		}
		status := agent.StatusUnhealthy
		if res.Healthy {
			status = agent.StatusHealthy
		}
		if a.Status != status {
			s.logger.Printf("heartbeat: %s %s -> %s (%s)", a.Name, a.Status, status, res.Detail)
		}
		if err := s.store.UpdateStatus(a.Name, status, res.CheckedAt); err != nil {
			s.logger.Printf("heartbeat: update %q failed: %v", a.Name, err)
		}
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, errorResponse{Error: msg})
}
