# Meower

**A CLI that generates a solid, opinionated Go web-app starting point — then gets out of your way.**

Meower is a project *generator*, not a framework. `meower new` scaffolds a
complete, multi-module Go application — gRPC API, server-rendered web frontend,
type-safe database access, observability, and a Docker dev environment — wired
together and ready to run. From there it's plain, idiomatic Go that you own and
change freely. There's no framework to stay inside and no upgrade treadmill;
the generated code is yours to drift from.

This makes it a good fit for LLM-assisted development: a known-good, fully-wired
base an agent can extend without first learning a bespoke framework.

## What you get

A generated project is a Go workspace (`go.work`) with these modules:

| Layer | Tech |
|-------|------|
| Web | Fiber + Templ + HTMX + TailwindCSS |
| API | Go + gRPC + PostgreSQL (SQLC) |
| Observability | OpenTelemetry traces + Prometheus metrics + structured logs |
| Dev | Docker Compose + `wgo` hot reload |

Out of the box it includes:

- **Type-safe URL/routing kit** (`pkg/urls`) — routes are typed structs that are
  the single source of truth for both URLs and Fiber patterns, so a renamed
  route is a compile error, not a silent 404.
- **Observability** (`pkg/observability` + `api/observability`) — OTLP tracing,
  a traced database pool, a Prometheus `/metrics` endpoint, and JSON logging in
  production (Loki-ready), all pre-wired.
- **Server-sent events hub** (`web/sse`) — an in-process pub/sub hub with a
  working HTMX live-update demo.
- **Middleware** — gRPC recovery, rate limiting (Redis), and trace enrichment;
  Fiber trace-ID headers.
- **Session auth** (login/signup/logout) — optional, removable with `--no-auth`.
- **Background workers** — a periodic-worker harness, removable with `--no-workers`.
- **CLAUDE.md** — agent guidance using a 🔴 invariant / 🟡 default / 🟢 preference
  rule taxonomy, documenting the generated structure and the kits above.
- **Monitoring overlay** — an opt-in Grafana/Loki/Tempo/Prometheus stack that
  pairs with the observability wiring.

## Quick start

```bash
go install github.com/AlyxPink/meower/cmd/meower@latest

# Generate a project (module path defaults to github.com/user/<name> if omitted)
meower new my-app -m github.com/you/my-app
cd my-app

# Bring up the full dev environment (hot reload, codegen, db, redis, …)
docker compose up
```

| Service | URL |
|---------|-----|
| Web | http://localhost:3000 |
| gRPC API | localhost:50051 |
| gRPC UI | http://localhost:50050 |
| Metrics | http://localhost:9091/metrics |
| pgweb (DB UI) | http://localhost:5430 |
| Mailpit | http://localhost:8025 |

### Generation flags

```bash
meower new my-app -m github.com/you/my-app   # batteries-included (default)
meower new my-app --no-auth                   # omit the auth scaffold
meower new my-app --no-workers                # omit the worker harness
meower new my-app --no-auth --no-workers      # the lean base
```

### Generate a service handler

```bash
meower create handler PostService
meower create handler UserService -m Create,Get,Update,Delete,List
```

## Generated project layout

```
my-app/
├── api/                 # gRPC API server
│   ├── proto/           #   Protocol Buffer definitions (.proto)
│   ├── db/              #   SQL schema + queries → SQLC-generated Go
│   ├── observability/   #   metrics server, traced DB pool, telemetry
│   └── server/          #   gRPC services, middleware, config, workers
├── web/                 # Fiber web server
│   ├── handlers/        #   request handlers
│   ├── views/           #   Templ templates (.templ)
│   ├── routing/         #   route registration (via the urls kit)
│   ├── sse/             #   server-sent-events hub
│   └── observability/   #   web tracing init
├── pkg/
│   ├── urls/            #   type-safe URL builder (shared)
│   └── observability/   #   shared logging + trace helpers
├── monitoring/          # opt-in Grafana/Loki/Tempo/Prometheus overlay
├── scripts/
├── docker-compose.yml
└── go.work
```

## How it works

The generated project ships source for its codegen (`.proto`, `.sql`, `.templ`)
and generates the Go code on build / first run. Under `docker compose up`,
`wgo` watches and regenerates on change:

- `.proto` → protobuf Go (gRPC services + messages)
- `.sql` → SQLC type-safe query methods
- `.templ` → Templ Go (type-safe HTML)
- CSS/JS → TailwindCSS + bundled assets

The modules resolve each other through `go.work` (local dev) and `replace`
directives (Docker builds), so the workspace builds without publishing anything.

## Development (working on Meower itself)

```bash
git clone https://github.com/AlyxPink/meower.git
cd meower
go build -o meower ./cmd/meower
go test ./internal/...
```

The complete project template lives under `cmd/meower/template/` and is embedded
into the CLI binary via `go:embed`. Template files use `TEMPLATE_MODULE_PATH` /
`TEMPLATE_PROJECT_NAME` placeholders that are substituted at generation time.

> Heads-up: if you previously ran `go install`, an older `meower` may sit on your
> `$PATH` and shadow a local `./meower` build. Run `./meower` during development,
> or re-run `go install ./cmd/meower` to refresh the installed binary.

## License

GNU Affero General Public License v3.0 — see [LICENSE](LICENSE).

---

**Made by [AlyxPink](https://github.com/AlyxPink/)**
