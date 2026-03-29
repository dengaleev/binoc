---
name: add-stack
description: Scaffold a new observability stack for the binoc playground. Use when asked to add, create, or set up a new monitoring stack.
model: sonnet
tools: Read, Write, Edit, Bash, Glob, Grep, WebSearch, WebFetch
---

# Add a new observability stack

Create `stacks/$ARGUMENTS/` with a working observability backend for the binoc service.

Read the **Stack principles** in `README.md` first. Every decision should follow them — especially: simplest ingestion path, no auth, pinned tags, three signals with correlation.

## Phase 1 — Research

Before writing any files, research the target backend:

1. **Search** for the official Docker compose / quickstart setup for the backend (e.g. "Elastic observability docker compose OTLP", "SigNoz docker self-hosted")
2. **Find** the recommended Docker images and their latest stable version tags
3. **Understand** how the backend ingests each signal:
   - Does it accept OTLP natively (gRPC/HTTP) or need a collector/agent with a specific exporter?
   - Does it need a separate component per signal or a unified endpoint?
   - What ports does it expose (UI, ingest, API)?
4. **Decide on the telemetry gateway** — the app must not talk to backends directly. Prefer the **simplest** option that works, not the most optimal:
   - **OTel Collector** — the default choice. Use when the backend accepts OTLP or has an OTel Collector contrib exporter.
   - **Backend's own agent** — use when the backend ships its own agent (e.g. Datadog Agent, Grafana Alloy) AND it can receive OTLP AND it's simpler to configure than the OTel Collector for that backend.
   The gateway must also **scrape Prometheus metrics** from `app:8080/metrics` and `caddy:2019/metrics` — the app uses Prometheus client library for custom metrics, these are not sent via OTLP.
5. **Identify** whether the backend bundles its own UI or needs Grafana (and if Grafana, which datasource plugin)
6. **Note** any required companion services (databases, queues, config stores)
7. **Cross-signal correlation** — how does the backend link traces ↔ logs ↔ metrics? (e.g. trace ID derived fields in logs, exemplars on metrics, service map data sources). Document what config is needed to enable these links.
8. **Auth and access** — find how to disable or bypass authentication so the UI is accessible immediately without registration or login. This is a playground — zero friction.

## Requirements

Every stack must satisfy these before it ships:

- [ ] `docker-compose.yml` — includes `../../docker-compose.base.yml`, adds backend services
- [ ] `OTEL_EXPORTER_OTLP_ENDPOINT` set on **app** (`http://` scheme, gRPC port) and **caddy** (HTTP port), pointing to the telemetry gateway
- [ ] Single telemetry gateway — app and caddy send OTLP to one service (OTel Collector, backend agent, or hybrid), never directly to storage backends
- [ ] Gateway scrapes Prometheus metrics from `app:8080` and `caddy:2019`
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
- [ ] Gateway exporters configured with retry and queue settings for resilience

## Reference

Read both existing stacks before starting:

- `stacks/loki-tempo-prometheus/` — Grafana + Loki + Tempo + Prometheus, OTel Collector as gateway
- `stacks/clickstack/` — HyperDX all-in-one (ClickHouse), OTel Collector writes directly to ClickHouse

Pay attention to:

- How `docker-compose.base.yml` is included and extended
- OTel Collector config pattern (receivers → processors → exporters, per-signal pipelines)
- How Prometheus scraping for app metrics is wired through the gateway
- Landing page structure matching the endpoint list
- How each stack disables auth (Grafana: anonymous editor; HyperDX: local app mode)
- How loki-tempo-prometheus configures cross-signal links in Grafana datasource provisioning (derived fields, traces-to-logs, traces-to-metrics, service map)

## Phase 2 — Scaffold

1. Read `docker-compose.base.yml`, both existing stacks, and `CLAUDE.md`
2. Create `stacks/$ARGUMENTS/docker-compose.yml`
3. Create the gateway config — either `otel-collector.yml` or the backend's agent config, depending on research. Include memory limits, retry, and queue settings on exporters.
4. Create backend-specific configs as needed
5. Create `stacks/$ARGUMENTS/index.html`
6. Provision dashboards with the 7 standard panels (request rate, latency p50/p95/p99, error rate, in-flight, log volume, traces table, logs viewer)
7. Configure cross-signal correlation (trace ID in logs → trace view, exemplars or service map in metrics → traces)
8. Disable auth / enable anonymous access on all UIs

## Phase 3 — Self-review

Before validating, re-read every file you created and check:

1. Ports — do the gateway exporter endpoints match the backend service ports?
2. Pipelines — is every signal (traces, logs, metrics) wired end-to-end from app to backend?
3. Prometheus scraping — does the gateway scrape `app:8080/metrics` and `caddy:2019/metrics`?
4. Dependencies — does the gateway wait for backends to be healthy before starting?
5. Auth — is every UI accessible without login?
6. Dashboard queries — do they reference the correct datasource UIDs and metric/label names?
7. Image tags — are all pinned to exact versions, no `latest`?

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

9. **Check gateway health** — verify no errors in the gateway (collector/agent) logs after the initial startup.

10. **Tear down**
    ```
    make down STACK=$ARGUMENTS
    ```

If any check fails, fix the issue and re-run from step 1.

## Conventions

- Stack directory name should describe the backends (e.g. `elastic`, `clickstack`, `loki-tempo-prometheus`)
- The app always sends OTLP — the gateway is responsible for translation if the backend doesn't speak OTLP
- Prometheus scraping of `app:8080/metrics` and `caddy:2019/metrics` must go through the gateway
- Use the same OTel Collector contrib image version as other stacks when using OTel Collector
