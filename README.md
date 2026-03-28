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

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/echo?msg=hello` | Returns the message as JSON. Simple request/response for basic traces. |
| `POST` | `/echo` | Echoes the request body back. |
| `GET` | `/notes` | List all notes. Produces DB query spans via `otelsql`. |
| `POST` | `/notes` | Create a note (`{"title":"…","content":"…"}`). DB insert span. |
| `GET` | `/notes/{id}` | Get a note by ID. |
| `DELETE` | `/notes/{id}` | Delete a note by ID. |
| `GET` | `/chain?msg=hello` | Calls `/echo` through Caddy, producing a distributed trace: Caddy → app → Caddy → app. |
| `GET` | `/healthz` | Liveness probe. |
| `GET` | `/readyz` | Readiness probe. |
| `GET` | `/metrics` | Prometheus metrics. |

### Why these routes exist

- **`/echo`** — minimal endpoint for baseline traces and metrics
- **`/notes`** — adds database spans (`sql.conn.exec`, `sql.conn.query`) so traces have depth
- **`/chain`** — creates multi-service distributed traces with trace context propagation across HTTP boundaries (app → Caddy → app)

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
