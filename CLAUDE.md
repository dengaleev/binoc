# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and run

```bash
make up                                # build and start default stack (loki-tempo-prometheus)
make up STACK=loki-tempo-prometheus    # explicit stack selection
make up STACK=clickstack              # ClickHouse + HyperDX
make up STACK=signoz                  # SigNoz + ClickHouse
make down                     # stop and remove volumes
make logs                     # tail all service logs
```

The compose setup uses `docker-compose.base.yml` (app + Caddy) included by each stack's own `docker-compose.yml`.

## After changing Go code

Run from `service/`:

```bash
gofmt -w .
go vet ./...
go build ./...
go mod tidy
```

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
.claude/agents/add-stack.md    # agent for scaffolding new stacks
```

**Request flow:** Client → Caddy (`:80`, strips `/api/` prefix, adds trace span) → App (`:8080`) → response.

**Instrumentation wiring:** `main.go` sets up the OTEL TracerProvider and LoggerProvider globally; `otelhttp` middleware auto-instruments all HTTP spans; `otelsql` wraps `database/sql`; `slog` output fans to both stdout and OTLP via `slog.NewMultiHandler` (Go 1.26).

**Middleware order** (outermost first): otel tracing → structured logging → prometheus metrics → mux.

## Key conventions

- Go 1.26; use modern stdlib: `slog.NewMultiHandler`, `net/http` method routing (`GET /echo`), `math/rand/v2`, range-over-int
- No web framework; plain `net/http` with `ServeMux`
- All app config via environment variables; OTEL-specific vars (`OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SERVICE_NAME`) handled by the SDK, not the app
- Functional options pattern for `server.New(opts...)`
- Always use context-aware slog methods (`InfoContext`, `ErrorContext`, etc.) when a `context.Context` is available — the `otelslog` bridge needs the context to propagate `TraceId`/`SpanId` to OTLP log records
- `/metrics` is excluded from tracing and logging via `isInternalPath`
- Distroless container, read-only filesystem
- No tests (playground project)

## Stack principles

See the **Stack** section under **Principles** in `README.md` — that is the source of truth. Do not duplicate the list here.

Use the `/add-stack <name> <description>` agent to scaffold new stacks.

## Proxy environment

Skip this section if `$HTTP_PROXY` is empty. When running behind a TLS-inspecting proxy (e.g. Claude Code web sandbox), all four steps below are required.

**1. Fix `no_proxy`** — if `*.googleapis.com` or `*.google.com` are listed in `$no_proxy`, remove them so traffic routes through the proxy. Without this, Go toolchain downloads and Docker image pulls fail with timeouts.

**2. Go toolchain** — override per command or export in shell:

```bash
no_proxy=localhost,127.0.0.1 go mod tidy
```

**3. Docker daemon** — create `/etc/docker/daemon.json` with proxy and registry mirror, then restart dockerd. Strip `*.googleapis.com` and `*.google.com` from `no-proxy` here too.

```json
{
  "proxies": {
    "http-proxy": "$HTTP_PROXY",
    "https-proxy": "$HTTPS_PROXY",
    "no-proxy": "localhost,127.0.0.1"
  },
  "registry-mirrors": ["https://mirror.gcr.io"]
}
```

**4. Docker build** — the daemon proxy only covers image pulls, not processes inside the build. Pass proxy vars as build args and add the proxy CA cert to the Dockerfile. The proxy does TLS inspection, so without the CA cert `go mod download` fails with `x509: certificate signed by unknown authority`.

Add to `docker-compose.base.yml` → `services.app.build`:

```yaml
args:
  - HTTP_PROXY
  - HTTPS_PROXY
  - NO_PROXY=localhost,127.0.0.1
```

Add to `service/Dockerfile` before `go mod download`:

```dockerfile
ARG HTTP_PROXY HTTPS_PROXY NO_PROXY

COPY proxy-ca.crt* /usr/local/share/ca-certificates/
RUN if [ -f /usr/local/share/ca-certificates/proxy-ca.crt ]; then \
      cat /usr/local/share/ca-certificates/proxy-ca.crt >> /etc/ssl/certs/ca-certificates.crt; \
    fi
```
