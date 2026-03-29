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
7. **Cross-signal correlation** — how does the backend link traces ↔ logs ↔ metrics? (e.g. trace ID derived fields in logs, exemplars on metrics, service map data sources). Document what config is needed to enable these links.
8. **Auth and access** — find how to disable or bypass authentication so the UI is accessible immediately without registration or login. This is a playground — zero friction.

## Requirements

Every stack must satisfy these before it ships:

- [ ] `docker-compose.yml` — includes `../../docker-compose.base.yml`, adds backend services
- [ ] `OTEL_EXPORTER_OTLP_ENDPOINT` set on **app** (`http://` scheme, gRPC port) and **caddy** (HTTP port)
- [ ] Single telemetry gateway — app and caddy never talk to backends directly
- [ ] All three signals — traces, logs, metrics — collected and queryable
- [ ] Cross-signal correlation — trace ID links logs to traces, service map or exemplars link metrics to traces
- [ ] `index.html` — landing page with endpoint links (`/api/` prefix) and monitoring tool links
- [ ] All images pinned to exact version tags — no `latest`, no floating tags
- [ ] Healthchecks on backend services; `depends_on` with `condition: service_healthy` where appropriate
- [ ] `restart: unless-stopped` on all backend services
- [ ] Config files mounted read-only (`:ro`)
- [ ] Named volumes for persistent data
- [ ] No auth — UI accessible without login, registration, or API keys
- [ ] Provisioned dashboards covering: request rate, latency percentiles (p50/p95/p99), error rate, in-flight requests, log volume, traces table, logs viewer
- [ ] OTel Collector exporters configured with `retry_on_failure` and `sending_queue` for resilience

## Reference

Read both existing stacks before starting:

- `stacks/loki-tempo-prometheus/` — Grafana + Loki + Tempo + Prometheus, separate backends per signal
- `stacks/clickstack/` — HyperDX all-in-one (ClickHouse), single backend for all signals

Pay attention to:

- How `docker-compose.base.yml` is included and extended
- OTel Collector config pattern (receivers → processors → exporters, per-signal pipelines)
- How the reference stack's `otel-collector.yml` wires Prometheus scraping for app metrics
- Landing page structure matching the endpoint list
- How each stack disables auth (Grafana: anonymous editor; HyperDX: local app mode)
- How loki-tempo-prometheus configures cross-signal links in Grafana datasource provisioning (derived fields, traces-to-logs, traces-to-metrics, service map)

## Phase 2 — Scaffold

1. Read `docker-compose.base.yml`, both existing stacks, and `CLAUDE.md`
2. Create `stacks/$ARGUMENTS/docker-compose.yml`
3. Create `stacks/$ARGUMENTS/otel-collector.yml` with OTLP receivers + Prometheus scraper + backend exporters; include `memory_limiter` processor and `retry_on_failure` on exporters
4. Create backend-specific configs as needed
5. Create `stacks/$ARGUMENTS/index.html`
6. Provision dashboards with the 7 standard panels (request rate, latency p50/p95/p99, error rate, in-flight, log volume, traces table, logs viewer)
7. Configure cross-signal correlation (trace ID in logs → trace view, exemplars or service map in metrics → traces)
8. Disable auth / enable anonymous access on all UIs

## Phase 3 — Self-review

Before validating, re-read every file you created and check:

1. Ports — do the OTel Collector exporter endpoints match the backend service ports?
2. Pipelines — is every signal (traces, logs, metrics) wired from receiver through processor to exporter?
3. Dependencies — does the collector wait for backends to be healthy before starting?
4. Auth — is every UI accessible without login?
5. Dashboard queries — do they reference the correct datasource UIDs and metric/label names?
6. Image tags — are all pinned to exact versions, no `latest`?

Fix any issues found before proceeding.

## Phase 4 — Validate

1. Run `docker compose -f stacks/$ARGUMENTS/docker-compose.yml config` and fix any errors
2. Update `README.md` stacks table
3. Update `CLAUDE.md` build-and-run section

## Phase 5 — End-to-end test

Build and run the full stack, then verify every aspect:

1. **Start the stack**
   ```
   make up STACK=$ARGUMENTS
   ```

2. **Test endpoints** — hit all four and confirm 200 responses:
   ```
   curl -s http://localhost/api/echo?msg=hello
   curl -s http://localhost/api/notes
   curl -s http://localhost/api/random
   curl -s http://localhost/api/chain?msg=hello
   ```

3. **Generate traffic** — send 10+ requests to each endpoint, wait for batch flush (~15s)

4. **Verify traces** — query the backend and confirm traces exist from both `binoc` and `caddy` services. Verify `/api/chain` produces a multi-service distributed trace.

5. **Verify logs** — query the backend and confirm log entries exist with structured fields (method, path, status, duration_ms). Check that TraceId/SpanId are populated for log-to-trace correlation.

6. **Verify metrics** — confirm these metrics are present:
   - `binoc_requests_total` (counter)
   - `binoc_request_duration_seconds` (histogram)
   - `binoc_response_size_bytes` (histogram)
   - `binoc_in_flight_requests` (gauge)

7. **Verify cross-signal correlation** — take a trace ID from a log entry and confirm the same trace ID exists in the traces backend.

8. **Verify UI access** — confirm the monitoring UI returns HTTP 200 without authentication.

9. **Check collector health** — verify no errors in the OTel Collector logs after the initial startup.

10. **Tear down**
    ```
    make down STACK=$ARGUMENTS
    ```

If any check fails, fix the issue and re-run from step 1.

## Conventions

- Stack directory name should describe the backends (e.g. `elastic`, `clickstack`, `loki-tempo-prometheus`)
- OTel Collector is always the telemetry gateway — app sends OTLP to collector, collector routes to backends
- Prometheus scraping of `app:8080/metrics` and `caddy:2019/metrics` goes through the collector
- Use the same OTel Collector contrib image version as other stacks for consistency
