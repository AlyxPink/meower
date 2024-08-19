# Meower

Meower is a Go boilerplate that moves the logic of your web application to an internal gRPC server. Along with some other tools, it provides a solid foundation for building web applications.

## Features

- [x] **Fiber** - HTTP web framework
- [x] **Protocol Buffers** - Interface definition language, to define the data structure and services
- [x] **gRPC** - High-performance RPC framework, where the web app logic lives
- [x] **PostgreSQL** - Relational database
- [x] **SQLC** - Generate type safe Go from SQL queries and schema
- [x] **Docker** - To containerize the application
- [x] **Docker Compose** - Run every services with a single command, reloading the server when required
- [x] **GitHub Actions** - CI/CD pipeline
- [x] **Templ** - Generate HTML using templates in Go, with hot browser reload
- [x] **Fly.io** - Global edge platform for applications
- [x] **Tailwind CSS** - Utility-first CSS framework
- [x] **gRPC UI** - gRPC Web implementation for testing gRPC servers
- [x] **Mailpit** - Local webmail client to receive emails from your app
- [x] **pgWeb** - Web-based database browser for PostgreSQL

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)

### Installation

```bash
git clone github.com/AlyxPink/meower
cd meower
```

The generated templ and sqlc files are gitignored, so you need to generate them first:

```bash
docker compose run -v $(pwd):/src/ --no-deps --tty --rm web templ generate
```

Then you can start the application along with every service:

```bash
docker compose up
```

You can now access the various services:

- Web app: [http://localhost:7331](http://localhost:7331)
- gRPC UI: [http://localhost:50050](http://localhost:50050)
- Mailpit: [http://localhost:8025](http://localhost:8025)
- pgWeb: [http://localhost:5430](http://localhost:5430)
