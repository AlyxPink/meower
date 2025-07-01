# 🐱 Meower Project Guide

**Welcome to your newly meowed project!** This guide will help you understand your project structure and get productive quickly.

## 🎯 Quick Start

```bash
# Start your development environment
docker-compose up

# In another terminal, generate a new service
meower create handler PostService -m Create,Get,List,Update,Delete

# Your app is now running at:
# - Web: http://localhost:3000
# - API: localhost:50051 (gRPC)
# - Database UI: http://localhost:5430
```

## 📁 Project Structure Deep Dive

### 🖥️ API Server (`/api`)

Your gRPC API server - this is where all your business logic lives.

```
api/
├── proto/              # Protocol Buffer definitions
│   ├── user/v1/       # User service (example)
│   └── meow/v1/       # Meow service (example)
├── server/            # gRPC server implementation
│   ├── handlers/      # Business logic handlers
│   └── server.go      # Server setup and registration
├── db/                # Database layer
│   ├── schema.sql     # Database schema definition
│   ├── *.sql         # SQL queries
│   └── *.go          # SQLC generated type-safe code
└── main.go            # API server entry point
```

**Key Files:**
- `proto/*/v1/*.proto` - Define your API contracts
- `server/handlers/*.go` - Implement your business logic
- `db/schema.sql` - Database tables and constraints
- `db/query.*.sql` - SQL queries for each service

### 🌐 Web Server (`/web`)

Your HTTP web server - handles web requests and renders HTML.

```
web/
├── handlers/          # HTTP request handlers
├── views/             # Templ templates
│   ├── layouts/       # Base page layouts
│   ├── pages/         # Individual pages
│   ├── components/    # Reusable UI components
│   └── services/      # Service-specific UI
├── static/            # Static assets
│   ├── src/css/       # Source CSS files
│   └── public/css/    # Compiled CSS (auto-generated)
├── routes/            # Route definitions
├── routing/           # Route registration
└── main.go            # Web server entry point
```

**Key Files:**
- `handlers/*.go` - Handle HTTP requests (call gRPC API)
- `views/**/*.templ` - HTML templates (type-safe!)
- `routes/routes.go` - Define route constants
- `routing/routing.go` - Register routes with handlers

## 🔄 Development Workflow

### 1. Adding a New Feature

Let's say you want to add a blog feature:

```bash
# 1. Generate the service handler
meower create handler BlogService -m Create,Get,List,Update,Delete

# 2. Define your database schema
# Edit api/db/schema.sql - add your blog table

# 3. Add database queries
# Edit api/db/query.blogs.sql - add your CRUD operations

# 4. Generate type-safe database code
sqlc generate

# 5. Implement business logic
# Edit api/server/handlers/blogservice.go

# 6. Create web handlers
# Edit web/handlers/blog.go

# 7. Create templates
# Edit web/views/pages/blog/index.templ

# 8. Register routes
# Edit web/routing/routing.go
```

### 2. Working with the Database

**Schema Changes:**
```sql
-- api/db/schema.sql
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    user_id UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**Queries:**
```sql
-- api/db/query.posts.sql
-- name: CreatePost :one
INSERT INTO posts (title, content, user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPost :one
SELECT * FROM posts WHERE id = $1;

-- name: ListPosts :many
SELECT * FROM posts ORDER BY created_at DESC;
```

**Generate Code:**
```bash
sqlc generate
```

### 3. gRPC Service Implementation

After generating a handler, implement the business logic:

```go
// api/server/handlers/blogservice.go
func (s *blogserviceServiceServer) CreateBlog(ctx context.Context, req *blogV1.CreateBlogRequest) (*blogV1.CreateBlogResponse, error) {
    // Use the generated SQLC code
    post, err := db.New(s.db).CreatePost(ctx, db.CreatePostParams{
        Title:   req.Title,
        Content: req.Content,
        UserID:  parseUUID(req.UserId), // implement parseUUID helper
    })
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to create post: %v", err)
    }
    
    return &blogV1.CreateBlogResponse{
        Blog: &blogV1.Blog{
            Id:        post.ID.String(),
            Title:     post.Title,
            Content:   post.Content,
            CreatedAt: timestamppb.New(post.CreatedAt),
        },
    }, nil
}
```

### 4. Web Handlers

Create HTTP handlers that call your gRPC service:

```go
// web/handlers/blog.go
func (h *Blog) List(c *fiber.Ctx) error {
    resp, err := h.API.BlogService.ListBlog(c.Context(), &blogV1.ListBlogRequest{})
    if err != nil {
        return err
    }
    
    return renderTempl(c, blog.Index(resp.Blogs))
}
```

### 5. Templ Templates

Create type-safe HTML templates:

```go
// web/views/pages/blog/index.templ
package blog

import "your-module/api/proto/blog/v1"

templ Index(blogs []*blogv1.Blog) {
    @layouts.Main("Blog Posts") {
        <div class="container mx-auto px-4">
            <h1 class="text-3xl font-bold mb-6">Blog Posts</h1>
            <div class="grid gap-4">
                for _, blog := range blogs {
                    <div class="p-4 border rounded">
                        <h2 class="text-xl font-bold">{ blog.Title }</h2>
                        <p class="text-gray-600">{ blog.Content }</p>
                    </div>
                }
            </div>
        </div>
    }
}
```

## 🛠️ Common Patterns

### Error Handling

**gRPC Services:**
```go
if err != nil {
    return nil, status.Errorf(codes.Internal, "operation failed: %v", err)
}
```

**Web Handlers:**
```go
if err != nil {
    return fiber.NewError(fiber.StatusInternalServerError, "Something went wrong")
}
```

### Validation

**Input Validation:**
```go
if req.Title == "" {
    return nil, status.Errorf(codes.InvalidArgument, "title is required")
}
```

### Database Transactions

```go
tx, err := s.db.Begin(ctx)
if err != nil {
    return nil, err
}
defer tx.Rollback()

// Use tx instead of s.db for queries
result, err := db.New(tx).CreatePost(ctx, params)
if err != nil {
    return nil, err
}

if err := tx.Commit(); err != nil {
    return nil, err
}
```

## 🎨 Frontend Development

### TailwindCSS Classes

Your project includes TailwindCSS. Common utility classes:

```html
<!-- Layout -->
<div class="container mx-auto px-4">
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">

<!-- Typography -->
<h1 class="text-3xl font-bold text-gray-900">
<p class="text-gray-600 leading-relaxed">

<!-- Buttons -->
<button class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">

<!-- Forms -->
<input class="border border-gray-300 rounded px-3 py-2 w-full">
```

### Component Organization

```go
// web/views/components/ui/button.templ
package ui

templ Button(text string, href string) {
    <a href={ templ.URL(href) } class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
        { text }
    </a>
}

// Usage in pages:
@ui.Button("Create Post", "/posts/new")
```

## 🚀 Production Deployment

### Building for Production

```bash
# Build API
cd api && go build -o bin/api

# Build Web
cd web && go build -o bin/web

# Build CSS
npm run build-css-prod
```

### Environment Variables

Create production environment files:

```bash
# .env.production
DATABASE_URL=postgres://user:pass@prod-db:5432/myapp
API_ENDPOINT=api:50051
COOKIE_SECRET_KEY=your-production-secret
ENV=production
```

### Docker Production

```dockerfile
# Dockerfile.prod
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY bin/api /
COPY bin/web /
EXPOSE 3000 50051
```

## 🐛 Debugging Tips

### gRPC Debugging

Use grpcui to test your API:
```bash
# Access at http://localhost:50050
grpcui -plaintext localhost:50051
```

### Database Debugging

Use pgweb to inspect your database:
```bash
# Access at http://localhost:5430
```

### Logs

```bash
# View API logs
docker-compose logs api

# View Web logs
docker-compose logs web

# Follow logs
docker-compose logs -f
```

## 🔧 Customization

### Adding Middleware

```go
// web/main.go
app.Use(cors.New())
app.Use(helmet.New())
app.Use(logger.New())
```

### Custom Validation

```go
// internal/validation/validator.go
func ValidateEmail(email string) error {
    if !strings.Contains(email, "@") {
        return errors.New("invalid email format")
    }
    return nil
}
```

### Authentication

```go
// web/handlers/middleware.go
func AuthMiddleware(store *session.Store) fiber.Handler {
    return func(c *fiber.Ctx) error {
        sess, err := store.Get(c)
        if err != nil {
            return c.Redirect("/login")
        }
        
        userID := sess.Get("user_id")
        if userID == nil {
            return c.Redirect("/login")
        }
        
        c.Locals("user_id", userID)
        return c.Next()
    }
}
```

## 📚 Additional Resources

- **Templ Documentation**: https://templ.guide/
- **Fiber Documentation**: https://docs.gofiber.io/
- **SQLC Documentation**: https://docs.sqlc.dev/
- **gRPC Go Tutorial**: https://grpc.io/docs/languages/go/
- **TailwindCSS Documentation**: https://tailwindcss.com/docs

## 🆘 Getting Help

If you run into issues:

1. **Check the logs**: `docker-compose logs <service-name>`
2. **Verify your .proto files**: Run `buf lint` in the api directory
3. **Regenerate code**: Run `sqlc generate` and `./scripts/generate_protobuf.sh`
4. **Join the community**: [GitHub Discussions](https://github.com/AlyxPink/meower/discussions)

---

**Happy coding! May your builds be fast and your bugs few! 🐱✨**