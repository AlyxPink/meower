# Monitoring stack

An opt-in observability stack — Prometheus (metrics), Loki (logs), Tempo
(traces), and Grafana (dashboards) — pre-wired to the telemetry the app already
emits. Bring it up alongside the app to get dashboards, traces, and log
correlation with zero extra setup.

| Component | Purpose | Port |
|-----------|---------|------|
| Grafana | Dashboards & Explore | http://localhost:3001 (admin/admin) |
| Prometheus | Metrics | 9090 |
| Loki | Logs | 3100 |
| Tempo | Traces (OTLP) | 3200, **4317** (gRPC), **4318** (HTTP) |

## Start it

```bash
docker compose -f monitoring/docker-compose.monitoring.yml up
```

It runs on its own Docker network, separate from `docker-compose.yml`, so you can
start and stop it independently of the app.

## How the loop closes

The app's telemetry defaults line up with this stack out of the box:

- **Metrics** — the api serves Prometheus metrics on `:9091` (see
  `api/observability/metrics.go`); Prometheus scrapes the `api` target.
- **Traces** — the api exports OTLP/gRPC to `:4317` and the web exports OTLP/HTTP
  to `:4318` (the Tempo receivers). The endpoint comes from
  `OTEL_EXPORTER_OTLP_ENDPOINT`; point it at `tempo:4317` / `tempo:4318` when the
  app and this stack share a network, or at `localhost` for host networking.
- **Logs** — in production the app logs JSON (`observability.ConfigureLogging`),
  which Loki parses. The Loki datasource extracts `trace_id` from log lines and
  links straight to the matching Tempo trace.

## Dashboards

Three generic dashboards are provisioned under
`grafana/provisioning/dashboards/metrics/`:

- **api-performance** — gRPC request rate, latency, and error rate.
- **web-performance** — HTTP request rate and latency (populates once the web
  server exposes a `/metrics` endpoint — the base instruments only the api).
- **infrastructure** — process/runtime metrics (goroutines, memory, GC) across
  scraped jobs.

Add your own dashboards as JSON in that directory; Grafana picks them up on
restart. Replace the example `app_users_total` gauge in `metrics.go` with metrics
that matter to your application and chart them here.
