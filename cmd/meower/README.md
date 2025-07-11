# Meower CLI

🐱 **Meower** is an opinionated Go web framework that provides a complete, production-ready foundation for building modern web applications.

## Features

- **🚀 Full-Stack Go**: Web server and API server in Go
- **🔗 gRPC Communication**: Type-safe service-to-service communication
- **🗄️ PostgreSQL + SQLC**: Type-safe database queries
- **📡 Protocol Buffers**: API-first development with Protobuf
- **🎨 Modern Frontend**: Server-side rendering with Templ + TailwindCSS
- **🐳 Docker Development**: Complete development environment
- **⚡ Hot Reload**: Live reloading for both frontend and backend

## Installation

```bash
go install github.com/AlyxPink/meower/cmd/meower@latest
```

## Quick Start

### Create a New Project

```bash
meower new my-app -m github.com/user/my-app
cd my-app
docker-compose up
```

Your app will be running at http://localhost:3000

### Generate Components

```bash
# Generate a gRPC service handler
meower create handler UserService

# Generate with specific methods
meower create handler PostService -m Create,Get,List
```

## Commands

- `meower new [project-name]` - Create a new Meower project
- `meower create handler [service-name]` - Generate gRPC service handlers
- `meower create model [model-name]` - Generate database models (coming soon)

## Project Structure

```
my-app/
├── api/                    # gRPC API server
│   ├── proto/             # Protocol buffer definitions
│   ├── server/            # gRPC handlers
│   ├── db/                # Database queries (SQLC)
│   └── main.go
├── web/                   # Web server
│   ├── handlers/          # HTTP handlers
│   ├── views/             # Templ templates
│   ├── static/            # CSS/JS assets
│   └── main.go
├── docker-compose.yml     # Development environment
└── scripts/               # Build scripts
```

## Development Workflow

1. **Start Development Environment**
   ```bash
   docker-compose up
   ```

2. **Generate Protocol Buffers**
   ```bash
   ./scripts/generate_protobuf.sh
   ```

3. **Generate Database Code**
   ```bash
   sqlc generate
   ```

4. **Hot Reload**
   - Backend changes trigger automatic rebuilds
   - Frontend changes trigger Templ and TailwindCSS rebuilds

## Technology Stack

- **Backend**: Go 1.23+ with gRPC
- **Frontend**: Go Fiber + Templ templates
- **Database**: PostgreSQL with SQLC
- **Styling**: TailwindCSS
- **Development**: Docker + Hot Reload
- **API**: Protocol Buffers + gRPC

## Contributing

This framework is built with developer experience in mind. If you have suggestions or find issues, please open an issue or pull request.

## License

MIT License - see LICENSE file for details.