# 🐱 Meower Framework

**An opinionated, production-ready Go web framework that makes building modern applications delightfully simple.**

Meower combines the best of Go's ecosystem into a cohesive, Rails-like developer experience. Build full-stack applications with gRPC APIs, server-side rendering, and type-safe database queries—all with a single CLI command.

## ✨ Features

### 🚀 **Full-Stack Go**
- **API Server**: gRPC-based microservice architecture
- **Web Server**: Fiber-powered HTTP server with server-side rendering
- **Type Safety**: End-to-end type safety from database to frontend

### 🛠️ **Developer Experience**
- **One Command Setup**: `meower new my-app` creates a complete project
- **Code Generators**: Generate services, handlers, and models instantly
- **Hot Reload**: Live reloading for both backend and frontend changes
- **Docker Integration**: Complete development environment in one command

### 🔧 **Tech Stack**
- **Backend**: Go + gRPC + PostgreSQL + SQLC
- **Frontend**: Go Fiber + Templ templates + TailwindCSS
- **Development**: Docker Compose + hot reload
- **API**: Protocol Buffers for type-safe communication

## 🚀 Quick Start

### Installation

```bash
go install github.com/AlyxPink/meower/cmd/meower@latest
```

### Create Your First Project

```bash
# Create a new project
meower new my-app -m github.com/user/my-app

# Start development environment
cd my-app
docker-compose up
```

Your app is now running at:
- **Web Interface**: http://localhost:3000
- **gRPC API**: localhost:50051
- **Database UI**: http://localhost:5430

### Generate Components

```bash
# Generate a complete gRPC service
meower create handler PostService

# Generate with specific methods
meower create handler UserService -m Create,Get,Update,Delete,List
```

## 📁 Project Structure

```
my-app/
├── 🖥️  api/                  # gRPC API Server
│   ├── proto/                # Protocol Buffer definitions
│   │   ├── user/v1/          # Versioned service definitions
│   │   └── post/v1/
│   ├── server/handlers/      # gRPC service implementations
│   ├── db/                   # Database layer (SQLC generated)
│   │   ├── schema.sql        # Database schema
│   │   ├── queries.sql       # SQL queries
│   │   └── *.go              # Generated type-safe code
│   └── main.go               # API server entry point
│
├── 🌐 web/                   # Web Server
│   ├── handlers/             # HTTP request handlers
│   ├── views/                # Templ templates
│   │   ├── layouts/          # Base layouts
│   │   ├── pages/            # Page templates
│   │   └── components/       # Reusable components
│   ├── static/               # CSS, JS, images
│   ├── routes/               # Route definitions
│   └── main.go               # Web server entry point
│
├── 🐳 docker-compose.yml     # Development environment
└── 📜 scripts/               # Build and utility scripts
```

## 🎯 Core Concepts

### **Split Architecture**
Meower uses a clean separation between your API and web layers:
- **API Server**: Pure business logic, database operations, gRPC endpoints
- **Web Server**: HTTP handlers, template rendering, static assets
- **Communication**: Type-safe gRPC calls between services

### **Type Safety Everywhere**
- **Database**: SQLC generates type-safe Go code from SQL queries
- **API**: Protocol Buffers ensure type safety across service boundaries
- **Frontend**: Templ provides type-safe HTML templating

### **Convention Over Configuration**
- **Standard Structure**: Consistent project layout across all Meower apps
- **Naming Conventions**: Predictable file and package naming
- **Code Generation**: Smart generators that follow established patterns

## 🛠️ Commands Reference

### Project Management
```bash
# Create new project
meower new <project-name> [flags]
  -m, --module string   Go module path (e.g. github.com/user/project)
  -f, --force          Force creation even if directory exists
```

### Code Generation
```bash
# Generate gRPC service handler
meower create handler <ServiceName> [flags]
  -m, --methods strings   Methods to generate (default: Create,Get,Update,Delete,List)

# Examples
meower create handler UserService
meower create handler PostService -m Create,Get,List
meower create handler AuthService -m Login,Logout,Register
```

## 🔄 Development Workflow

### 1. **Start Development Environment**
```bash
docker compose up
```
This starts all services with hot reload enabled:
- API server with live recompilation
- Web server with Templ template reloading
- TailwindCSS with file watching
- PostgreSQL database
- Development tools (pgweb, mailpit)

### 2. **Make Changes**
- **API Changes**: Edit files in `api/`, server restarts automatically
- **Frontend Changes**: Edit `.templ` files, browser refreshes automatically
- **Database Changes**: Update `schema.sql`, run migrations
- **Styles**: Edit CSS files, TailwindCSS rebuilds automatically

### 3. **Generate Code**
```bash
# After adding new SQL queries
sqlc generate

# After modifying .proto files
./scripts/generate_protobuf.sh

# Add new services
meower create handler PaymentService
```

## 🔧 Configuration

### Environment Variables
```bash
# API Configuration
DATABASE_URL=postgres://user:pass@localhost:5432/dbname
API_ENDPOINT=localhost:50051

# Web Configuration
COOKIE_SECRET_KEY=your-secret-key
ENV=development  # or production
```

### Database Setup
Meower uses PostgreSQL with SQLC for type-safe queries:

1. **Define Schema**: Edit `api/db/schema.sql`
2. **Write Queries**: Add queries to `api/db/queries.sql`
3. **Generate Code**: Run `sqlc generate`
4. **Use in Handlers**: Import and use generated functions

## 🎨 Frontend Development

### Templ Templates
Meower uses [Templ](https://templ.guide/) for type-safe HTML templating:

```go
// views/pages/home.templ
package pages

templ HomePage(title string, posts []Post) {
    @layouts.Base(title) {
        <div class="container mx-auto px-4">
            <h1 class="text-3xl font-bold">{ title }</h1>
            for _, post := range posts {
                @components.PostCard(post)
            }
        </div>
    }
}
```

### TailwindCSS Integration
- **Automatic Building**: CSS rebuilds on file changes
- **Component Classes**: Organized in `web/static/src/css/`
- **Production Optimization**: Minified builds for deployment

## 🚀 Deployment

### Production Build
```bash
# Build API server
cd api && go build -o api ./cmd/api

# Build web server
cd web && go build -o web ./cmd/web

# Build assets
npm run build-css-prod
```

### Docker Deployment
```bash
# Build production images
docker build -f api/Dockerfile -t my-app-api .
docker build -f web/Dockerfile -t my-app-web .
```

## 🤝 Contributing

We welcome contributions! Here's how to get started:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** and add tests
4. **Run tests**: `go test ./...`
5. **Submit a pull request**

### Development Setup
```bash
git clone https://github.com/AlyxPink/meower.git
cd meower
go mod tidy
go build -o meower ./cmd/meower
```

## 📄 License

GNU Affero General Public License v3.0 - see [LICENSE](LICENSE) file for details.

## 🎉 Why Meower?

**"Building web applications shouldn't require assembling 20 different tools."**

Meower gives you:
- ✅ **Batteries Included**: Everything you need out of the box
- ✅ **Type Safety**: Catch errors at compile time, not runtime
- ✅ **Developer Joy**: Fast feedback loops and intuitive workflows
- ✅ **Production Ready**: Built for real applications, not just demos
- ✅ **Go All The Way**: Pure Go stack with excellent performance

---

**Made with 🩷 by [AlyxPink](https://github.com/AlyxPink/).**

*May your code purr smoothly and your builds never hiss!* 🚀
