---
name: add-stack
description: Scaffold a new observability stack for the binoc playground. Use when asked to add, create, or set up a new monitoring stack.
model: sonnet
tools: Read, Write, Edit, Bash, Glob, Grep
---

# Add a new observability stack

Create `stacks/$ARGUMENTS/` with a working observability backend for the binoc service.

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

## Steps

1. Read `docker-compose.base.yml`, both existing stacks, and `CLAUDE.md`
2. Create `stacks/$ARGUMENTS/docker-compose.yml`
3. Create `stacks/$ARGUMENTS/otel-collector.yml` with OTLP receivers + Prometheus scraper + backend exporters
4. Create backend-specific configs as needed
5. Create `stacks/$ARGUMENTS/index.html`
6. Add provisioned dashboards if the stack uses Grafana
7. Validate compose config: `docker compose -f stacks/$ARGUMENTS/docker-compose.yml config`
8. Update `README.md` stacks table
9. Update `CLAUDE.md` build-and-run section

## Conventions

- Stack directory name should describe the backends (e.g. `elastic`, `clickstack`, `loki-tempo-prometheus`)
- OTel Collector is always the telemetry gateway — app sends OTLP to collector, collector routes to backends
- Prometheus scraping of `app:8080/metrics` and `caddy:2019/metrics` goes through the collector
- Use the same OTel Collector contrib image version as other stacks for consistency
