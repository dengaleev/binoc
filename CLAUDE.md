# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Observability playground — a minimal Go service paired with interchangeable monitoring stacks. The service produces traces, logs, and metrics; each stack collects and visualizes them.

## Build and run

```bash
make up                       # build and start default stack (grafana-lgtm)
make up STACK=grafana-lgtm    # explicit stack selection
make down                     # stop and remove volumes
make logs                     # tail all service logs
```

The compose setup uses `docker-compose.base.yml` (app + Caddy) included by each stack's own `docker-compose.yml`.

## Architecture

```
service/
  cmd/binoc/main.go           # entry point: config → instrument → server → run
  internal/
    config/                    # env-only config via caarlos0/env, OTEL SDK reads its own vars
    instrument/                # logging (slog+otelslog), tracing (otlptracegrpc), metrics (prometheus)
    server/                    # net/http handlers, middleware (otel, logging, metrics), background ticker
    store/                     # in-memory SQLite via otelsql, schema+seed in single SQL constant
stacks/<name>/                 # each stack: docker-compose.yml + backend configs + provisioned dashboards
```

**Request flow:** Client → Caddy (`:80`, strips `/api/` prefix, adds trace span) → App (`:8080`) → response.

**Instrumentation wiring:** `main.go` sets up the OTEL TracerProvider and LoggerProvider globally; `otelhttp` middleware auto-instruments all HTTP spans; `otelsql` wraps `database/sql`; `slog` output fans to both stdout and OTLP via `slog.NewMultiHandler` (Go 1.26).

**Middleware order** (outermost first): otel tracing → structured logging → prometheus metrics → mux.

## Key conventions

- Go 1.26; no web framework, plain `net/http` with method routing (`GET /echo`)
- All app config via environment variables; OTEL-specific vars (`OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SERVICE_NAME`) handled by the SDK, not the app
- Functional options pattern for `server.New(opts...)`
- `/metrics` is excluded from tracing and logging via `isInternalPath`
- Distroless container, read-only filesystem
- No tests (playground project)

## Building behind a proxy

When Docker builds run behind a TLS-inspecting proxy (e.g. Claude Code web sandbox), three things need configuring:

**1. Docker daemon** — set registry mirror and proxy in `/etc/docker/daemon.json`:
```json
{
  "proxies": {
    "http-proxy": "$HTTP_PROXY",
    "https-proxy": "$HTTPS_PROXY",
    "no-proxy": "$NO_PROXY"
  },
  "registry-mirrors": ["https://mirror.gcr.io"]
}
```
Restart dockerd after changes.

**2. Build args** — pass proxy vars so `go mod download` works inside the build:
```yaml
# docker-compose.base.yml → services.app.build.args
args:
  - HTTP_PROXY
  - HTTPS_PROXY
  - NO_PROXY=localhost,127.0.0.1
```

**3. Proxy CA certificate** — if the proxy does TLS inspection, place the CA cert at `service/proxy-ca.crt` and add to Dockerfile before `go mod download`:
```dockerfile
ARG HTTP_PROXY HTTPS_PROXY NO_PROXY

COPY proxy-ca.crt* /usr/local/share/ca-certificates/
RUN if [ -f /usr/local/share/ca-certificates/proxy-ca.crt ]; then \
      cat /usr/local/share/ca-certificates/proxy-ca.crt >> /etc/ssl/certs/ca-certificates.crt; \
    fi
```
