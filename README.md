# binoc

Observability playground — a minimal Go service paired with different monitoring stacks.

## Quickstart

```bash
make up                       # starts the default stack (grafana-lgtm)
make up STACK=grafana-lgtm    # or pick one explicitly
make list                     # show available stacks
```

Open http://localhost to see the navigation page, or go directly:

- **Grafana** http://localhost:3000 (admin / admin)
- **Prometheus** http://localhost:9090

## Service

A single Go binary (`service/`) that exposes several HTTP endpoints behind a Caddy reverse proxy. Each endpoint is designed to exercise a different observability pattern.

### Routes

All app endpoints are published through Caddy under the `/api` prefix. Caddy strips the prefix before proxying to the app.

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/echo?msg=hello` | Returns the message as JSON. Simple request/response for basic traces. |
| `POST` | `/api/echo` | Echoes the request body back. |
| `GET` | `/api/notes` | List notes from in-memory SQLite (seeded on startup). Produces DB query spans via `otelsql`. |
| `GET` | `/api/chain?msg=hello` | Calls `/api/echo` through Caddy, producing a distributed trace: Caddy → app → Caddy → app. |
| `GET` | `/api/random` | Random 0-500ms delay, ~10% error rate. Makes latency and error panels interesting. |
| `GET` | `/time` | Returns current time (served directly by Caddy, not proxied to app). |

Technical endpoints (`/healthz`, `/readyz`, `/metrics`) are only on the app container (`:8080`), not exposed through Caddy.

### Why these routes exist

- **`/echo`** — minimal endpoint for baseline traces and metrics
- **`/notes`** — adds database query spans (`sql.conn.query`) so traces have depth
- **`/chain`** — creates multi-service distributed traces with trace context propagation across HTTP boundaries (app → Caddy → app)
- **`/random`** — generates realistic latency spread and error rate so dashboard panels aren't flat
- **`/time`** — Caddy-only handler; polled by a background ticker in the app every 1s to produce continuous background traces and logs

### Instrumentation

| Signal | Library | How |
|--------|---------|-----|
| Traces | `otelhttp` | Auto-instruments HTTP server spans |
| Traces | `otelsql` | Auto-instruments `database/sql` operations |
| Traces | Caddy `tracing` | Caddy creates spans for reverse proxy requests |
| Metrics | `prometheus/client_golang` | Request rate, latency, in-flight, response size |
| Metrics | `otelsql` | DB connection pool stats |
| Logs | `otelslog` bridge | Sends `slog` output to OTLP collector → Loki |

## Stacks

| Stack | Components |
|-------|------------|
| `grafana-lgtm` | Loki + Grafana + Tempo + Prometheus, with OTel Collector |

Each stack lives in `stacks/<name>/` and includes `docker-compose.base.yml` for shared services.

## Make Targets

| Target | Description |
|--------|-------------|
| `up` | Build and start the stack |
| `down` | Stop the stack and remove volumes |
| `logs` | Tail logs from all services |
| `build` | Build the service image |
| `list` | List available stacks |
