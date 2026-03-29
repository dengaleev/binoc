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
