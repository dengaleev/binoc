# binoc

Observability playground — a minimal Go service paired with different monitoring stacks.

## Principles

### Service

- Minimal — few endpoints, in-memory state, no external dependencies
- Env-only configuration; OTEL standard env vars for instrumentation
- Fully instrumented — every request produces a trace, a log line, and a metric
- Diverse workloads — varying latency, error rates, DB spans, distributed calls

### Stack

- Single-node, no HA — playground, not production
- Minimal configuration — sensible defaults, works out of the box
- Simplest ingestion path — prefer the approach with least configuration, not the most optimal
- Three signals — metrics, logs, traces with cross-signal correlation
- Single telemetry gateway — app never talks to backends directly
- End-to-end traces — reverse proxy propagates trace context
- Provisioned dashboards covering golden signals, logs, and traces
- No auth — UIs accessible without login, registration, or API keys
- Pinned image tags — all images use exact version tags, no `latest`

## Quickstart

```bash
make up                                # starts the default stack (loki-tempo-prometheus)
make up STACK=loki-tempo-prometheus    # or pick one explicitly
make up STACK=clickstack              # ClickHouse + HyperDX
make up STACK=signoz                  # SigNoz + ClickHouse
make list                             # show available stacks
```

Open http://localhost to see the navigation page with links to all endpoints and tools.

## Service

A single Go binary (`service/`) behind a Caddy reverse proxy. Each endpoint exercises a different observability pattern.

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/echo?msg=hello` | Returns the message as JSON — baseline traces and metrics |
| `POST` | `/api/echo` | Echoes the request body back |
| `GET` | `/api/notes` | Lists notes from in-memory SQLite — adds DB query spans |
| `GET` | `/api/chain?msg=hello` | Calls `/api/echo` through Caddy — distributed trace across services |
| `GET` | `/api/random` | Random 0-500ms delay, ~10% error rate — realistic latency and errors |

A background ticker pokes `/api/random` every second for continuous telemetry.

## Stacks

Each stack lives in `stacks/<name>/` and extends `docker-compose.base.yml` for shared services.

| Stack | Description |
|-------|-------------|
| `loki-tempo-prometheus` | Grafana + Loki + Tempo + Prometheus — separate backend per signal |
| `clickstack` | HyperDX + ClickHouse — unified backend for all signals |
| `signoz` | SigNoz + ClickHouse — OpenTelemetry-native unified backend |

## Adding a Stack

Use the `/add-stack <name> <description>` agent. It follows the stack principles above and handles research, scaffolding, self-review, validation, and end-to-end testing.

## Make Targets

| Target | Description |
|--------|-------------|
| `up` | Build and start the stack |
| `down` | Stop the stack and remove volumes |
| `logs` | Tail logs from all services |
| `build` | Build the service image |
| `list` | List available stacks |
