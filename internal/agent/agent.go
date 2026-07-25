// Package agent defines the core registry record: an agent (MCP server) and its
// discovery, health, and provenance metadata.
package agent

import "time"

// Status is the health state of an agent as last observed by the registry.
type Status string

const (
	// StatusUnknown means the agent has been registered but not yet probed.
	StatusUnknown Status = "unknown"
	// StatusHealthy means the last verification pass succeeded.
	StatusHealthy Status = "healthy"
	// StatusUnhealthy means the last verification pass failed or was unreachable.
	StatusUnhealthy Status = "unhealthy"
)

// SourceLocal marks a record registered directly against this registry.
//
// Source is the federation/ingestion seam. v1 only ever writes "local", but the
// field exists from day one so later work is additive, not a migration:
//   - "local"            registered directly here
//   - "peer:<id>"        synced from a federated peer registry
//   - "from:ans"         ingested from the Agent Name Service
//   - "from:mcp_registry" ingested from the MCP Registry
//
// Discovery (Find) never branches on Source; it is provenance metadata only.
const SourceLocal = "local"

// Agent is a registered agent / MCP server.
type Agent struct {
	Name          string    `json:"name"`
	Capabilities  []string  `json:"capabilities"`
	Endpoint      string    `json:"endpoint"`
	Source        string    `json:"source"`
	Status        Status    `json:"status"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	RegisteredAt  time.Time `json:"registered_at"`
}

// HasCapability reports whether the agent advertises the given capability.
func (a Agent) HasCapability(capability string) bool {
	for _, c := range a.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}
