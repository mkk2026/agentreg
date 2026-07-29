// Package client is a thin HTTP client the CLI uses to talk to the daemon.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"syscall"
	"time"

	"github.com/mkk2026/agentreg/internal/agent"
)

// startupGrace is how long the client keeps retrying a "connection refused"
// error, so a command run immediately after `agentctl serve &` succeeds instead
// of racing the daemon's startup.
const startupGrace = 3 * time.Second

// Client talks to a running agentreg daemon.
type Client struct {
	baseURL string
	http    *http.Client
}

// New returns a Client pointed at baseURL (e.g. http://localhost:8080).
func New(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 10 * time.Second}}
}

// Register registers an agent and returns the stored record.
func (c *Client) Register(ctx context.Context, a agent.Agent) (agent.Agent, error) {
	body, err := json.Marshal(a)
	if err != nil {
		return agent.Agent{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/agents", bytes.NewReader(body))
	if err != nil {
		return agent.Agent{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	var out agent.Agent
	if err := c.do(req, &out); err != nil {
		return agent.Agent{}, err
	}
	return out, nil
}

// List returns all registered agents.
func (c *Client) List(ctx context.Context) ([]agent.Agent, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/agents", nil)
	if err != nil {
		return nil, err
	}
	var out []agent.Agent
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Find returns agents that advertise the given capability.
func (c *Client) Find(ctx context.Context, capability string) ([]agent.Agent, error) {
	u := c.baseURL + "/agents/find?capability=" + url.QueryEscape(capability)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	var out []agent.Agent
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.doWithRetry(req)
	if err != nil {
		return fmt.Errorf("cannot reach registry at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		var er struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &er) == nil && er.Error != "" {
			return fmt.Errorf("registry error (%d): %s", resp.StatusCode, er.Error)
		}
		return fmt.Errorf("registry error (%d)", resp.StatusCode)
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// doWithRetry retries only on "connection refused" (the daemon isn't up yet),
// for up to startupGrace, so a command run right after `agentctl serve &` waits
// for the daemon instead of failing the race. Any other error returns
// immediately. Request bodies are rewound via GetBody between attempts.
func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
	deadline := time.Now().Add(startupGrace)
	for attempt := 0; ; attempt++ {
		if attempt > 0 && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			req.Body = body
		}
		resp, err := c.http.Do(req)
		if err == nil {
			return resp, nil
		}
		if errors.Is(err, syscall.ECONNREFUSED) && time.Now().Before(deadline) {
			time.Sleep(150 * time.Millisecond)
			continue
		}
		return nil, err
	}
}
