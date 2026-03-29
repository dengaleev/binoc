---
name: add-stack
description: Scaffold a new observability stack for the binoc playground. Use when asked to add, create, or set up a new monitoring stack.
model: sonnet
tools: Read, Write, Edit, Bash, Glob, Grep, WebSearch, WebFetch
---

# Add a new observability stack

Create `stacks/$ARGUMENTS/` with a working observability backend for the binoc service.

## Phase 1 — Research

Before writing any files, research the target backend:

1. **Search** for the official Docker compose / quickstart setup for the backend (e.g. "Elastic observability docker compose OTLP", "SigNoz docker self-hosted")
2. **Find** the recommended Docker images and their latest stable version tags
3. **Understand** how the backend ingests each signal:
   - Does it accept OTLP natively (gRPC/HTTP) or need the OTel Collector with a specific exporter?
   - Does it need a separate component per signal or a unified endpoint?
   - What ports does it expose (UI, ingest, API)?
4. **Check** OTel Collector contrib exporter support — search for the relevant exporter in the collector-contrib docs and note required config fields
5. **Identify** whether the backend bundles its own UI or needs Grafana (and if Grafana, which datasource plugin)
6. **Note** any required companion services (databases, queues, config stores)

## Requirements

Every stack must satisfy these before it ships:

- [ ] `docker-compose.yml` — includes `../../docker-compose.base.yml`, adds backend services
- [ ] `OTEL_EXPORTER_OTLP_ENDPOINT` set on **app** (`http://` scheme, gRPC port) and **caddy** (HTTP port)
- [ ] Single telemetry gateway — app and caddy never talk to backends directly
- [ ] All three signals — traces, logs, metrics — collected and queryable
- [ ] `index.html` — landing page with endpoint links (`/api/` prefix) and monitoring tool links
- [ ] All images pinned to exact version tags — no `latest`, no floating tags
- [ ] Healthchecks on backend services; `depends_on` with `condition: service_healthy` where appropriate
- [ ] `restart: unless-stopped` on all backend services
- [ ] Config files mounted read-only (`:ro`)
- [ ] Named volumes for persistent data

## Reference

Read both existing stacks before starting:

- `stacks/loki-tempo-prometheus/` — Grafana + Loki + Tempo + Prometheus, separate backends per signal
- `stacks/clickstack/` — HyperDX all-in-one (ClickHouse), single backend for all signals

Pay attention to:

- How `docker-compose.base.yml` is included and extended
- OTel Collector config pattern (receivers → processors → exporters, per-signal pipelines)
- How the reference stack's `otel-collector.yml` wires Prometheus scraping for app metrics
- Landing page structure matching the endpoint list

## Phase 2 — Scaffold

1. Read `docker-compose.base.yml`, both existing stacks, and `CLAUDE.md`
2. Create `stacks/$ARGUMENTS/docker-compose.yml`
3. Create `stacks/$ARGUMENTS/otel-collector.yml` with OTLP receivers + Prometheus scraper + backend exporters
4. Create backend-specific configs as needed
5. Create `stacks/$ARGUMENTS/index.html`
6. Add provisioned dashboards if the stack uses Grafana

## Phase 3 — Validate

1. Run `docker compose -f stacks/$ARGUMENTS/docker-compose.yml config` and fix any errors
2. Review each config file against the research findings — verify ports, endpoints, and auth match the backend docs
3. Update `README.md` stacks table
4. Update `CLAUDE.md` build-and-run section

## Conventions

- Stack directory name should describe the backends (e.g. `elastic`, `clickstack`, `loki-tempo-prometheus`)
- OTel Collector is always the telemetry gateway — app sends OTLP to collector, collector routes to backends
- Prometheus scraping of `app:8080/metrics` and `caddy:2019/metrics` goes through the collector
- Use the same OTel Collector contrib image version as other stacks for consistency
