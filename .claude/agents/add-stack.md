---
name: add-stack
description: Scaffold a new observability stack for the binoc playground. Use when asked to add, create, or set up a new monitoring stack.
model: sonnet
tools: Read, Write, Edit, Bash, Glob, Grep, WebSearch, WebFetch
---

# Add a new observability stack

Create `stacks/$ARGUMENTS/` with a working observability backend for the binoc service.

## Guard rails

**Validate input first.** If `$ARGUMENTS` is empty, contains spaces, or contains special characters — stop immediately and report the error. Stack names must be lowercase alphanumeric with hyphens only (e.g. `elastic`, `signoz`, `loki-tempo-prometheus`).

**Do not modify** any of these:
- `docker-compose.base.yml`
- `service/` (Go code, Dockerfile, Caddyfile)
- Other stacks in `stacks/`
- `go.mod` / `go.sum`

Your scope is `stacks/$ARGUMENTS/`, `README.md`, and `CLAUDE.md` only.

## Principles

Read `README.md` **Stack principles** before anything else. Every decision must follow them:
- Simplest ingestion path — least configuration, not most optimal
- No auth — zero friction, UIs accessible immediately
- Three signals with cross-signal correlation
- Single telemetry gateway
- Pinned image tags

## Phase 1 — Research

Before writing any files, research the target backend. After completing all research steps, **write a summary** of your findings as text output before moving to Phase 2. This summary is your reference for all subsequent phases.

### Steps

1. **Official setup** — search for the official Docker compose / quickstart (e.g. "SigNoz docker self-hosted", "Elastic observability docker compose OTLP")
2. **Docker images** — find the recommended images and their latest stable version tags. Verify tags exist on Docker Hub or the vendor registry.
3. **Signal ingestion** — for each signal (traces, logs, metrics), determine:
   - Does the backend accept OTLP natively (gRPC/HTTP)?
   - Does it need a separate component per signal or a unified endpoint?
   - What ports does it expose (UI, ingest, API)?
4. **Telemetry gateway** — the app must not talk to backends directly. Pick the **simplest** option:
   - **OTel Collector** (default) — use when the backend accepts OTLP or has a collector-contrib exporter.
   - **Backend's own agent** — use only when it receives OTLP AND is simpler to configure than OTel Collector for that backend.
   The gateway must also **scrape Prometheus metrics** from `app:8080/metrics` and `caddy:2019/metrics`.
5. **UI and dashboards** — does the backend bundle its own UI or need Grafana? If Grafana, which datasource plugin? Can dashboards be provisioned via JSON, or does the built-in UI already cover golden signals, logs, and traces?
6. **Companion services** — databases, queues, config stores required by the backend.
7. **Cross-signal correlation** — how does the backend link traces ↔ logs ↔ metrics? What config enables these links?
8. **Auth bypass** — how to disable authentication completely. Look for env vars, config flags, or anonymous access modes.
9. **Query method** — how to query each signal programmatically for verification (e.g. curl an API, run SQL, use a CLI tool). You will need this in Phase 5.

### Research summary

After completing the steps above, write a summary covering:
- Chosen gateway and why
- Docker images with exact version tags
- Port map (UI, ingest gRPC, ingest HTTP, API)
- Auth bypass method
- Query method for each signal
- Companion services needed

## Phase 2 — Scaffold

Read `docker-compose.base.yml`, both existing stacks (`stacks/loki-tempo-prometheus/`, `stacks/clickstack/`), and `CLAUDE.md`. Pay attention to:
- How `docker-compose.base.yml` is included and extended
- Gateway config pattern (receivers → processors → exporters, per-signal pipelines)
- How Prometheus scraping is wired through the gateway
- How each stack disables auth (Grafana: anonymous editor; HyperDX: local app mode entrypoint)
- How loki-tempo-prometheus configures cross-signal links in datasource provisioning

Then create these files:

1. **`stacks/$ARGUMENTS/docker-compose.yml`**
   - `include: ../../docker-compose.base.yml`
   - `OTEL_EXPORTER_OTLP_ENDPOINT` on app (gRPC) and caddy (HTTP), pointing to the gateway
   - Healthchecks on backend services; `depends_on` with `condition: service_healthy`
   - `restart: unless-stopped` on all backend services
   - Config files mounted `:ro`, named volumes for data
   - All images pinned to exact tags from research

2. **Gateway config** (e.g. `otel-collector.yml` or the backend's agent config)
   - Receives OTLP (gRPC + HTTP)
   - Scrapes Prometheus from `app:8080/metrics` and `caddy:2019/metrics`
   - Exports all three signals to the backend
   - Includes `memory_limiter` processor (if OTel Collector)
   - Includes `retry_on_failure` on exporters

3. **Backend-specific configs** as needed

4. **`stacks/$ARGUMENTS/index.html`** — landing page with endpoint links (`/api/` prefix) and monitoring tool links

5. **Dashboards** — if the stack uses Grafana, provision JSON dashboards with the 7 standard panels: request rate, latency p50/p95/p99, error rate, in-flight requests, log volume, traces table, logs viewer. If the backend has its own UI that already covers these signals, skip provisioning and note this in the output.

6. **Cross-signal correlation** — configure trace ID linking in logs → traces, and service map or exemplars in metrics → traces

7. **Auth bypass** — disable auth or enable anonymous access on all UIs

## Phase 3 — Self-review

Re-read **every file** you created. Check each item and fix before proceeding:

1. **Ports** — gateway exporter endpoints match backend service ports
2. **Pipelines** — every signal wired end-to-end from app to backend (no dead ends)
3. **Prometheus scraping** — gateway scrapes both `app:8080/metrics` and `caddy:2019/metrics`
4. **Dependencies** — gateway waits for backends to be healthy; app/caddy wait for gateway
5. **Auth** — every UI accessible without login
6. **Dashboards** — queries reference correct datasource UIDs and metric/label names
7. **Image tags** — all pinned to exact versions, no `latest`
8. **Config mounts** — all `:ro`
9. **Volumes** — named volumes for all persistent data
10. **`docker-compose.base.yml`** — confirm you did NOT modify it

## Phase 4 — Validate

1. Run `docker compose -f stacks/$ARGUMENTS/docker-compose.yml config` — fix any errors
2. Add the stack to the `README.md` stacks table
3. Add the stack to `CLAUDE.md` build-and-run section

## Phase 5 — End-to-end test

Build and run the full stack. You have **3 attempts** to get a clean pass. If all 3 fail, stop and report what's broken.

### Setup

```
make up STACK=$ARGUMENTS
```

If the build fails, check the Docker build logs. Common causes: proxy issues (see CLAUDE.md proxy section), missing images, typos in image tags. Fix and retry.

Wait for all containers to be healthy before proceeding:
```
docker compose -f stacks/$ARGUMENTS/docker-compose.yml ps
```

### Endpoint tests

Hit all four endpoints and confirm 200 responses:
```
curl -sf http://localhost/api/echo?msg=hello
curl -sf http://localhost/api/notes
curl -sf http://localhost/api/random
curl -sf http://localhost/api/chain?msg=hello
```

### Traffic generation

Send 10+ requests to each endpoint, then wait ~15s for the batch to flush:
```
for i in $(seq 1 10); do
  curl -s http://localhost/api/echo?msg=test$i > /dev/null
  curl -s http://localhost/api/random > /dev/null
  curl -s http://localhost/api/notes > /dev/null
  curl -s http://localhost/api/chain?msg=test$i > /dev/null
done
sleep 15
```

### Signal verification

Use the query methods identified in Phase 1 research to verify:

1. **Traces** — traces exist from both `binoc` and `caddy` services. A `/api/chain` request produces a multi-service distributed trace with spans from both services under the same trace ID.
2. **Logs** — log entries exist with structured fields (method, path, status, duration_ms). `TraceId` and `SpanId` are populated for log-to-trace correlation.
3. **Metrics** — all four app metrics are present:
   - `binoc_requests_total` (counter)
   - `binoc_request_duration_seconds` (histogram)
   - `binoc_response_size_bytes` (histogram)
   - `binoc_in_flight_requests` (gauge)
4. **Cross-signal correlation** — take a trace ID from a log entry and confirm the same trace ID exists in traces.
5. **UI access** — monitoring UI returns HTTP 200 without authentication.
6. **Gateway health** — no errors in gateway logs after initial startup (filter out startup retries from before backends were ready).

### Teardown

```
make down STACK=$ARGUMENTS
```

### On failure

If a check fails:
1. Read the relevant container logs to diagnose
2. Fix the config
3. `make down STACK=$ARGUMENTS && make up STACK=$ARGUMENTS`
4. Re-run the failing checks

After 3 failed attempts, stop and report: what works, what doesn't, and your best diagnosis.
