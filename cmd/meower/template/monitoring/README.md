# Monitoring stack

An opt-in observability stack — Prometheus (metrics), Loki (logs), Tempo
(traces), and Grafana (dashboards) — pre-wired to the telemetry the app already
emits. It ships in the main `docker-compose.yml` behind the `monitoring` compose
profile, so `docker compose --profile monitoring up` brings it up alongside the
app to get dashboards, traces, and log correlation with zero extra setup.

| Component | Purpose | Port |
|-----------|---------|------|
| Grafana | Dashboards & Explore | http://localhost:3001 (admin/admin) |
| Prometheus | Metrics | 9090 |
| Loki | Logs | 3100 |
| Tempo | Traces (OTLP) | 3200, **4317** (gRPC), **4318** (HTTP) |

## Start it

The monitoring services live in the main `docker-compose.yml` but are gated
behind the `monitoring` [compose profile][profiles], so a plain `docker compose
up` leaves them stopped. Bring them up alongside the app with:

```bash
docker compose --profile monitoring up
```

Because they share the project's default network with the app, traces flow with
**zero extra configuration** — the api and web already point at `tempo:4317` /
`tempo:4318`. Grafana comes up at http://localhost:3001 (admin/admin).

To stop just the monitoring containers while leaving the app running:

```bash
docker compose --profile monitoring down
```

[profiles]: https://docs.docker.com/compose/how-tos/profiles/

## Running without the monitoring stack

A plain `docker compose up` (no profile) does **not** start Prometheus/Loki/
Tempo/Grafana. The app still tries to export traces to `tempo:4317` / `:4318`,
but since nothing is listening the exporter degrades silently: it logs a single
warning and then stays quiet (repeats drop to debug) rather than spamming a
`connection refused` line every few seconds. Spans are simply dropped.

To send traces somewhere else instead — a shared collector, an external Tempo,
a vendor endpoint — set `OTEL_EXPORTER_OTLP_ENDPOINT` in a `.env` file at the
project root; Docker Compose reads it automatically and it overrides the
`tempo:*` default for both services. (Note the api uses OTLP/gRPC on `:4317`
and the web uses OTLP/HTTP on `:4318`, so any override must point at a collector
that accepts those protocols on the host:port you give it.)

## How the loop closes

The app's telemetry defaults line up with this stack out of the box:

- **Metrics** — the api serves Prometheus metrics on `:9091` (see
  `api/observability/metrics.go`); Prometheus scrapes the `api` target.
- **Traces** — the api exports OTLP/gRPC to `tempo:4317` and the web exports
  OTLP/HTTP to `tempo:4318` (the Tempo receivers). The endpoint comes from
  `OTEL_EXPORTER_OTLP_ENDPOINT` (default `tempo:4317` / `tempo:4318`), which
  resolves over the shared default network when the `monitoring` profile is up.
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
