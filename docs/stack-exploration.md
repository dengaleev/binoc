# Observability Stack Exploration

Comprehensive survey of monitoring stacks that can be added to the binoc playground. Each stack reuses the same instrumented Go service — only the backend infrastructure changes.

**Current stack:** `loki-tempo-prometheus` (Grafana + OTel Collector + Prometheus + Loki + Tempo)

**App signals available:**
| Signal  | Export method             | Protocol        |
|---------|--------------------------|-----------------|
| Traces  | OTLP gRPC                | OpenTelemetry   |
| Metrics | `/metrics` endpoint + OTLP | Prometheus + OT |
| Logs    | OTLP gRPC                | OpenTelemetry   |

---

## 1. ClickHouse-Based Stacks

### 1a. SigNoz (all-in-one)

| Component     | Tool                         |
|---------------|------------------------------|
| Collector     | OTel Collector (bundled)     |
| Storage       | ClickHouse                   |
| Visualization | SigNoz UI                    |

- **CNCF status:** Not a CNCF project (OSS, Apache 2.0)
- **OTLP support:** Native — SigNoz is built entirely on OpenTelemetry
- **Docker:** Official docker-compose available (`signoz/signoz`)
- **Value:** Single platform for all three signals with built-in alerting, service maps, trace-to-logs correlation, and exceptions tracking. ClickHouse gives excellent query performance on high-cardinality data.
- **Complexity:** Medium — requires ClickHouse, query-service, frontend, and OTel Collector

### 1b. Uptrace (all-in-one)

| Component     | Tool                         |
|---------------|------------------------------|
| Collector     | OTel Collector (bundled)     |
| Storage       | ClickHouse + PostgreSQL      |
| Visualization | Uptrace UI                   |

- **OTLP support:** Native OTLP ingestion (gRPC + HTTP)
- **Docker:** `uptrace/uptrace` image, docker-compose provided
- **Value:** Unified traces/metrics/logs with grouping, alerting, dashboards. Uses ClickHouse for telemetry and PostgreSQL for metadata. Has a generous open-source edition.
- **Complexity:** Medium — ClickHouse + PostgreSQL + Uptrace

### 1c. ClickHouse + Grafana (DIY)

| Component     | Tool                           |
|---------------|--------------------------------|
| Collector     | OTel Collector                 |
| Storage       | ClickHouse                     |
| Visualization | Grafana + ClickHouse plugin    |

- **OTLP support:** OTel Collector has a native ClickHouse exporter
- **Docker:** `clickhouse/clickhouse-server`
- **Value:** Maximum flexibility, extremely fast analytical queries, SQL-based exploration. Great for teams that want full control over schema and retention.
- **Complexity:** Higher — requires schema design, OTel Collector config, Grafana dashboards

---

## 2. VictoriaMetrics-Based Stacks

### 2a. VictoriaMetrics + VictoriaLogs + Tempo

| Component     | Tool                                    |
|---------------|-----------------------------------------|
| Collector     | OTel Collector or vmagent               |
| Metrics       | VictoriaMetrics                         |
| Logs          | VictoriaLogs                            |
| Traces        | Tempo (or Jaeger)                       |
| Visualization | Grafana                                 |

- **OTLP support:** VictoriaMetrics supports OTLP metrics ingestion natively (via `/opentelemetry/v1/metrics`). VictoriaLogs supports OTLP logs ingestion natively.
- **Docker:** `victoriametrics/victoria-metrics`, `victoriametrics/victoria-logs`
- **Value:** Drop-in Prometheus replacement with better compression (up to 10x), faster queries, lower resource usage. VictoriaLogs is purpose-built for log storage with LogsQL query language. Compatible with PromQL and Prometheus remote write.
- **Complexity:** Low-medium — similar architecture to current stack but with VM components

### 2b. VictoriaMetrics Full Stack (VM + VictoriaLogs + VictoriaTraces)

| Component     | Tool                           |
|---------------|--------------------------------|
| Collector     | OTel Collector                 |
| All signals   | VictoriaMetrics ecosystem      |
| Visualization | Grafana or VM UI               |

- **Note:** VictoriaTraces (based on Tempo fork) is available for trace storage
- **Value:** Single-vendor stack with excellent resource efficiency. All components share similar configuration patterns and operational model.

---

## 3. Grafana LGTM Stack Variants

### 3a. Grafana Mimir + Loki + Tempo (cloud-native)

| Component     | Tool                           |
|---------------|--------------------------------|
| Collector     | Grafana Alloy (or OTel Collector) |
| Metrics       | Grafana Mimir                  |
| Logs          | Loki                           |
| Traces        | Tempo                          |
| Visualization | Grafana                        |

- **CNCF status:** Not CNCF (Grafana Labs OSS, AGPL-3.0)
- **OTLP support:** Mimir has native OTLP metrics ingestion. Alloy is a full OTel-compatible collector.
- **Docker:** `grafana/mimir`, `grafana/alloy`
- **Value:** Mimir is horizontally scalable Prometheus-compatible long-term storage. Alloy replaces both Prometheus agent and OTel Collector. This is the "production Grafana stack."
- **Complexity:** Medium — Mimir adds more components than Prometheus but gives scalability

### 3b. Grafana Alloy as Collector

Replace OTel Collector with Grafana Alloy in any Grafana-based stack. Alloy is built on OTel Collector but adds Grafana-specific features like native Loki/Mimir/Tempo exporters, a River configuration language, and a built-in UI.

---

## 4. Elastic / OpenSearch Stacks

### 4a. OpenSearch Stack

| Component     | Tool                           |
|---------------|--------------------------------|
| Collector     | OTel Collector or Data Prepper |
| Storage       | OpenSearch                     |
| Visualization | OpenSearch Dashboards          |

- **CNCF status:** Not CNCF (Linux Foundation project)
- **OTLP support:** OpenSearch Data Prepper has native OTLP ingestion for traces, metrics, and logs. OpenSearch also supports direct OTLP ingestion.
- **Docker:** `opensearchproject/opensearch`, `opensearchproject/data-prepper`
- **Value:** Full-text search on logs, trace analytics with service maps, anomaly detection. Familiar Elasticsearch-like query language. Good for teams needing powerful log search.
- **Complexity:** Medium-high — OpenSearch is resource-hungry

### 4b. ELK / Elastic Stack

| Component     | Tool                           |
|---------------|--------------------------------|
| Collector     | Elastic APM Server or OTel Collector |
| Storage       | Elasticsearch                  |
| Visualization | Kibana                         |

- **OTLP support:** Elastic APM Server supports native OTLP ingestion (since 7.x). The OTel Collector also has an Elasticsearch exporter.
- **Docker:** `elasticsearch`, `kibana`, `elastic/apm-server`
- **Value:** Industry-standard log analytics. Kibana has powerful visualization. Elastic APM provides service maps, error tracking, and ML-based anomaly detection.
- **Complexity:** High — Elasticsearch is resource-intensive, licensing considerations (SSPL)

---

## 5. CNCF Tracing Backends

### 5a. Jaeger (CNCF Graduated)

| Component     | Tool                           |
|---------------|--------------------------------|
| Collector     | Jaeger Collector (built-in OTLP) |
| Storage       | Badger (local), Cassandra, Elasticsearch, or ClickHouse |
| Visualization | Jaeger UI                      |

- **CNCF status:** Graduated
- **OTLP support:** Jaeger v2 is built on OTel Collector — native OTLP ingestion on ports 4317/4318
- **Docker:** `jaegertracing/jaeger:latest` (all-in-one) or `jaegertracing/jaeger-v2`
- **Value:** The reference distributed tracing platform. Jaeger v2 unifies all components into a single binary based on OTel Collector. Excellent trace search, comparison, and dependency graphs.
- **Stack:** Jaeger + Prometheus + Loki (traces from Jaeger, metrics/logs from existing)

### 5b. Zipkin

| Component     | Tool                           |
|---------------|--------------------------------|
| Collector     | OTel Collector (Zipkin exporter) |
| Storage       | In-memory, MySQL, Cassandra, Elasticsearch |
| Visualization | Zipkin UI                      |

- **OTLP support:** Via OTel Collector Zipkin exporter (translates OTLP → Zipkin format)
- **Docker:** `openzipkin/zipkin`
- **Value:** Lightweight, simple trace visualization. Good for learning distributed tracing concepts. Minimal resource usage with in-memory storage.
- **Stack:** Zipkin + Prometheus + Loki

### 5c. Apache SkyWalking

| Component     | Tool                           |
|---------------|--------------------------------|
| Collector     | SkyWalking OAP (supports OTLP) |
| Storage       | H2 (embedded), Elasticsearch, BanyanDB |
| Visualization | SkyWalking UI                  |

- **OTLP support:** SkyWalking OAP server accepts OTLP gRPC for traces, metrics, and logs
- **Docker:** `apache/skywalking-oap-server`, `apache/skywalking-ui`
- **Value:** Full APM platform with topology maps, service mesh observability, browser monitoring, and alarm system. Built-in profiling support.
- **Complexity:** Medium — OAP server + UI + storage backend

---

## 6. Alternative Metrics Backends

### 6a. Thanos (CNCF Incubating)

| Component     | Tool                           |
|---------------|--------------------------------|
| Metrics       | Prometheus + Thanos Sidecar    |
| Long-term     | Thanos Store + Object Storage  |
| Query         | Thanos Query                   |

- **CNCF status:** Incubating
- **OTLP support:** Via Prometheus (Thanos extends Prometheus)
- **Docker:** `thanosio/thanos`
- **Value:** Long-term metrics storage with object storage (S3/MinIO), global query view across multiple Prometheus instances, downsampling. Extends rather than replaces Prometheus.
- **Stack:** Thanos + Prometheus + Loki + Tempo + Grafana

### 6b. Cortex (CNCF Graduated)

| Component     | Tool                           |
|---------------|--------------------------------|
| Metrics       | Cortex (replaces Prometheus)   |
| Storage       | Object storage (S3/MinIO)      |

- **CNCF status:** Graduated
- **OTLP support:** Via Prometheus remote write from OTel Collector
- **Docker:** `cortexproject/cortex`
- **Value:** Horizontally scalable, multi-tenant Prometheus backend. 100% PromQL compatible. Good for demonstrating multi-tenant metrics isolation.
- **Note:** Mimir is the spiritual successor (Grafana fork of Cortex)

### 6c. InfluxDB

| Component     | Tool                           |
|---------------|--------------------------------|
| Metrics       | InfluxDB                       |
| Visualization | InfluxDB UI or Grafana         |

- **OTLP support:** OTel Collector has an InfluxDB exporter. InfluxDB 3.0 has native OTLP support.
- **Docker:** `influxdb:latest`
- **Value:** Purpose-built time-series database with its own query language (Flux/SQL). Good for high-write workloads and downsampling.

### 6d. TimescaleDB (PostgreSQL-based)

| Component     | Tool                           |
|---------------|--------------------------------|
| Metrics       | TimescaleDB + Promscale        |
| Visualization | Grafana                        |

- **OTLP support:** Via Prometheus remote write
- **Docker:** `timescale/timescaledb`
- **Value:** SQL-based metrics storage. Use standard PostgreSQL tooling for analysis. Good for teams already using PostgreSQL.
- **Note:** Promscale was deprecated; consider the Prometheus PostgreSQL adapter instead

---

## 7. Alternative Log Backends

### 7a. Quickwit

| Component     | Tool                           |
|---------------|--------------------------------|
| Collector     | OTel Collector                 |
| Log storage   | Quickwit                       |
| Visualization | Quickwit UI or Grafana         |

- **OTLP support:** Native OTLP ingestion for both logs and traces (gRPC + HTTP)
- **Docker:** `quickwit/quickwit`
- **Value:** Sub-second search on cloud storage (S3/MinIO). Extremely cost-efficient for large log volumes. Written in Rust, very performant. Also supports trace storage. Tantivy-based full-text search.
- **Stack:** Quickwit (logs+traces) + Prometheus + Grafana

### 7b. Fluentd/Fluentbit + Various Sinks

| Component     | Tool                           |
|---------------|--------------------------------|
| Collector     | Fluentd or Fluent Bit          |
| Storage       | Elasticsearch, S3, Loki, etc. |

- **CNCF status:** Fluentd is CNCF Graduated; Fluent Bit is CNCF sub-project (Graduated)
- **OTLP support:** Fluent Bit has OTLP input plugin (receives OTLP, forwards to any output)
- **Docker:** `fluent/fluent-bit`, `fluent/fluentd`
- **Value:** Demonstrates the CNCF-graduated log collection path. Fluent Bit is extremely lightweight (C-based). Can show an alternative to OTel Collector for log routing.

---

## 8. Emerging / Newer Platforms

### 8a. GreptimeDB

| Component     | Tool                           |
|---------------|--------------------------------|
| Storage       | GreptimeDB (unified)           |
| Visualization | Grafana                        |

- **OTLP support:** Native OTLP ingestion for metrics, logs, and traces
- **Docker:** `greptime/greptimedb`
- **Value:** Unified time-series database handling metrics, logs, and traces in a single system. SQL + PromQL compatible. Cloud-native with distributed architecture. Written in Rust.
- **Stack:** GreptimeDB + OTel Collector + Grafana

### 8b. Coroot

| Component     | Tool                           |
|---------------|--------------------------------|
| Collector     | Coroot node-agent (eBPF)       |
| Storage       | ClickHouse + Prometheus        |
| Visualization | Coroot UI                      |

- **OTLP support:** Accepts OTLP traces
- **Docker:** `coroot/coroot`
- **Value:** Automated observability with eBPF-based auto-instrumentation. Zero-code service map generation. Automatic anomaly detection and root cause analysis. Focuses on reducing alert fatigue.

### 8c. Perses (CNCF Sandbox)

| Component     | Tool                           |
|---------------|--------------------------------|
| Visualization | Perses (dashboard-only)        |

- **CNCF status:** Sandbox
- **OTLP support:** N/A (visualization layer, queries Prometheus/Tempo/etc.)
- **Docker:** `persesdev/perses`
- **Value:** CNCF-native Grafana alternative. Dashboard-as-code with GitOps support. Prometheus-native. Could replace Grafana in any existing stack.

### 8d. Grafana Pyroscope (Continuous Profiling)

| Component     | Tool                           |
|---------------|--------------------------------|
| Profiling     | Pyroscope                      |
| Visualization | Grafana                        |

- **Docker:** `grafana/pyroscope`
- **Value:** Adds a 4th signal — continuous profiling. Flamegraphs linked to traces. Can integrate with any existing stack as an add-on. The Go service could add `pprof` instrumentation to demonstrate profile-trace correlation.
- **Note:** Would require minor app changes to push profiles

---

## 9. All-in-One Platforms

### 9a. Highlight.io

- Open-source full-stack monitoring (sessions, errors, logs, traces)
- Docker-compose available
- OTLP support for backend telemetry

### 9b. HyperDX

- Open-source observability platform
- ClickHouse-based storage
- Native OTLP support
- Docker-compose available
- Unified search across all signal types

---

## Recommended Implementation Priority

Based on unique value, Docker-compose friendliness, and educational diversity:

| Priority | Stack Name                    | Signals           | Why                                          |
|----------|-------------------------------|-------------------|----------------------------------------------|
| 1        | **jaeger-prometheus**         | Traces+Metrics    | CNCF graduated, Jaeger v2 is OTLP-native    |
| 2        | **victoriametrics-loki-tempo**| All three         | Shows VM as Prometheus drop-in replacement   |
| 3        | **signoz**                    | All three         | ClickHouse-based all-in-one, very popular    |
| 4        | **quickwit-prometheus-jaeger**| All three         | Rust-based search engine, cost-efficient logs|
| 5        | **opensearch**                | All three         | Full-text search focus, Data Prepper pipeline|
| 6        | **elk**                       | All three         | Industry standard, Elastic APM with OTLP    |
| 7        | **mimir-loki-tempo**          | All three         | Production Grafana stack with Alloy collector|
| 8        | **clickhouse-grafana**        | All three         | DIY ClickHouse, maximum flexibility          |
| 9        | **skywalking**                | All three         | Full APM with topology and profiling         |
| 10       | **greptimedb**                | All three         | Emerging unified DB, single storage engine   |
| 11       | **thanos-loki-tempo**         | All three         | Long-term storage with object storage        |
| 12       | **zipkin-prometheus-loki**    | All three         | Lightweight tracing, educational             |
| 13       | **uptrace**                   | All three         | ClickHouse all-in-one alternative to SigNoz  |
| 14       | **coroot**                    | All three         | eBPF auto-instrumentation focus              |

---

## Implementation Pattern

Each new stack follows the same pattern:

```
stacks/<stack-name>/
  docker-compose.yml          # includes ../../docker-compose.base.yml
  <backend-configs>.yml       # collector, storage, etc.
  index.html                  # landing page
  provisioning/               # dashboards and datasources (if Grafana-based)
```

The app service requires **zero code changes** — all stacks work by:
1. Setting `OTEL_EXPORTER_OTLP_ENDPOINT` to point at the stack's collector
2. Configuring the collector to route signals to the appropriate backends
3. Optionally scraping `/metrics` for Prometheus-format metrics

For non-Grafana stacks (SigNoz, Uptrace, Jaeger UI, etc.), no provisioning directory is needed — dashboards come built-in.
