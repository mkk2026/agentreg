# agentreg

**DNS for AI agents.** A single self-hosted Go binary that lets your agents
register themselves and discover each other by capability, with health built in.

> Consul tells you *where* a service is. agentreg tells you *what* an agent can
> do, *whether* it's healthy, and (soon) *how much to trust it.*

No cloud dependency. No account. One binary, 30-second setup.

## Quickstart

```bash
# build
go build -o agentctl .

# run the registry daemon
./agentctl serve --port 8080 &

# register your agents
./agentctl register search-agent -c search      -e http://localhost:3000
./agentctl register db-agent     -c db-read      -e http://localhost:3001
./agentctl register product-api  -c api,email    -e http://localhost:3002

# see everything, with live health
./agentctl list

# discover by capability (use this in your agent startup instead of a hardcoded URL)
./agentctl find search
./agentctl find search --format json
```

`list` prints a table with health status and last heartbeat:

```
NAME          CAPABILITIES  STATUS     SOURCE  LAST HEARTBEAT  ENDPOINT
db-agent      db-read       unhealthy  local   just now        http://localhost:3001
product-api   api,email     healthy    local   just now        http://localhost:3002
search-agent  search        healthy    local   just now        http://localhost:3000
```

The daemon health-checks every registered agent on an interval, so a down agent
shows up as `unhealthy` **before** it breaks something downstream.

## How it works

```
  agentctl (CLI)  --HTTP-->  agentreg daemon
                                 |
                                 +-- Store (in-memory + JSON file)
                                 +-- Verifier (health probe, on a heartbeat loop)
```

- **Store** — `internal/registry`. In-memory map with atomic JSON persistence,
  behind a `Store` interface so a database backend is a drop-in later.
- **Verifier** — `internal/verify`. The trust seam. v1 ships a health probe; the
  interface is designed so ANS-identity, prompt-injection, and `tools/list`
  verifiers are additive, not a rewrite.
- **Source field** — every record carries `source` (`local` today; `peer:<id>`,
  `from:ans`, `from:mcp_registry` later) so federation and external-registry
  ingestion don't require a migration.

## HTTP API

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/agents` | register an agent (JSON body) |
| GET | `/agents` | list all agents |
| GET | `/agents/find?capability=X` | find by capability |
| POST | `/agents/{name}/heartbeat` | agent self-reports healthy |
| GET | `/healthz` | daemon liveness |

## Configuration

```
agentctl serve
  --port 8080                       # listen port
  --store ~/.agentreg/registry.json # persistence file
  --heartbeat-interval 15s          # how often to health-check agents
  --probe-timeout 3s                # per-agent probe timeout
```

## Not in v1 (roadmap)

Federation, PKI/ANS wire-compatibility, TUI dashboard, `doctor`/`diff`/`graph`
commands, zero-config auto-register. v1 is deliberately the smallest lovable
loop. See the design docs.

## Development

```bash
go test ./...   # unit tests for store, verifier, server
go vet ./...
```

## License

TBD.
