package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/AlyxPink/meower/internal/templates"
)

// HandlerGenerator generates complete gRPC service implementations.
// This generator creates a full gRPC service stack including:
// - Protocol buffer service definitions (.proto files)
// - Server-side handler implementations with TODO comments
// - Web client integration stubs
// - Proper module path handling for generated code
//
// The generator uses Go's text/template package to ensure proper
// code formatting and supports customizable method sets.
type HandlerGenerator struct {
	vars *templates.TemplateVars
}

// NewHandlerGenerator creates a new handler generator
func NewHandlerGenerator(vars *templates.TemplateVars) *HandlerGenerator {
	return &HandlerGenerator{
		vars: vars,
	}
}

// GenerateProto generates the protocol buffer definition
func (g *HandlerGenerator) GenerateProto(methods []string) error {
	// Create proto directory
	protoDir := filepath.Join("api", "proto", g.vars.ServiceNameLower, "v1")
	if err := os.MkdirAll(protoDir, 0o755); err != nil {
		return fmt.Errorf("failed to create proto directory: %w", err)
	}

	// Generate proto file
	protoFile := filepath.Join(protoDir, g.vars.ServiceNameLower+".proto")

	protoTemplate := `syntax = "proto3";

package {{.ServiceNameLower}}.v1;

import "google/protobuf/timestamp.proto";

option go_package = "{{.ModulePath}}/api/proto/{{.ServiceNameLower}}/v1";

service {{.ServiceName}} {
{{- range .Methods}}
  rpc {{.}}{{$.ResourceName}}({{.}}{{$.ResourceName}}Request) returns ({{.}}{{$.ResourceName}}Response) {}
{{- end}}
}

message {{.ResourceName}} {
  string id = 1;
  string name = 2;
  google.protobuf.Timestamp created_at = 3;
  google.protobuf.Timestamp updated_at = 4;
}

{{- range .Methods}}

message {{.}}{{$.ResourceName}}Request {
{{- if eq . "Create"}}
  string name = 1;
{{- else if eq . "Get"}}
  string id = 1;
{{- else if eq . "Update"}}
  string id = 1;
  string name = 2;
{{- else if eq . "Delete"}}
  string id = 1;
{{- else if eq . "List"}}
  int32 limit = 1;
  int32 offset = 2;
{{- end}}
}

message {{.}}{{$.ResourceName}}Response {
{{- if eq . "List"}}
  repeated {{$.ResourceName}} {{$.ResourceNameLower}}s = 1;
{{- else if eq . "Delete"}}
  bool success = 1;
{{- else}}
  {{$.ResourceName}} {{$.ResourceNameLower}} = 1;
{{- end}}
}
{{- end}}
`

	tmpl, err := template.New("proto").Parse(protoTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse proto template: %w", err)
	}

	// Prepare template data
	resourceName := strings.TrimSuffix(g.vars.ServiceName, "Service")
	data := struct {
		*templates.TemplateVars
		Methods           []string
		ResourceName      string
		ResourceNameLower string
	}{
		TemplateVars:      g.vars,
		Methods:           methods,
		ResourceName:      resourceName,
		ResourceNameLower: strings.ToLower(resourceName),
	}

	file, err := os.Create(protoFile)
	if err != nil {
		return fmt.Errorf("failed to create proto file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute proto template: %w", err)
	}

	return nil
}

// GenerateServerHandler generates the server-side gRPC handler
func (g *HandlerGenerator) GenerateServerHandler(methods []string) error {
	// Create handler directory
	handlerDir := filepath.Join("api", "server", "handlers")
	if err := os.MkdirAll(handlerDir, 0o755); err != nil {
		return fmt.Errorf("failed to create handler directory: %w", err)
	}

	// Generate handler file
	handlerFile := filepath.Join(handlerDir, g.vars.ServiceNameLower+".go")

	handlerTemplate := `package handlers

import (
	"context"
	"fmt"

	"{{.ModulePath}}/api/db"
	{{.ServiceNameLower}}V1 "{{.ModulePath}}/api/proto/{{.ServiceNameLower}}/v1"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type {{.ServiceNameLower}}ServiceServer struct {
	{{.ServiceNameLower}}V1.Unimplemented{{.ServiceName}}Server
	db *pgxpool.Pool
}

func New{{.ServiceName}}Server(db *pgxpool.Pool) {{.ServiceNameLower}}V1.{{.ServiceName}}Server {
	return &{{.ServiceNameLower}}ServiceServer{db: db}
}

{{- $resourceName := .ResourceName}}
{{- $resourceNameLower := .ResourceNameLower}}
{{- $serviceLower := .ServiceNameLower}}
{{- range .Methods}}

func (s *{{$serviceLower}}ServiceServer) {{.}}{{$resourceName}}(ctx context.Context, req *{{$serviceLower}}V1.{{.}}{{$resourceName}}Request) (*{{$serviceLower}}V1.{{.}}{{$resourceName}}Response, error) {
{{- if eq . "Create"}}
	// TODO: Implement create logic
	// Example:
	// result, err := db.New(s.db).Create{{$resourceName}}(ctx, db.Create{{$resourceName}}Params{
	//     Name: req.Name,
	// })
	// if err != nil {
	//     return nil, status.Errorf(codes.Internal, "failed to create {{$resourceNameLower}}: %v", err)
	// }

	return &{{$serviceLower}}V1.{{.}}{{$resourceName}}Response{
		{{$resourceName}}: &{{$serviceLower}}V1.{{$resourceName}}{
			Id:        "generated-id",
			Name:      req.Name,
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
		},
	}, nil
{{- else if eq . "Get"}}
	// TODO: Implement get logic
	// Example:
	// uuid, err := parseUUID(req.Id)
	// if err != nil {
	//     return nil, status.Errorf(codes.InvalidArgument, "invalid ID: %v", err)
	// }
	//
	// result, err := db.New(s.db).Get{{$resourceName}}ById(ctx, uuid)
	// if err != nil {
	//     return nil, status.Errorf(codes.NotFound, "{{$resourceNameLower}} not found: %v", err)
	// }

	return &{{$serviceLower}}V1.{{.}}{{$resourceName}}Response{
		{{$resourceName}}: &{{$serviceLower}}V1.{{$resourceName}}{
			Id:        req.Id,
			Name:      "Sample {{$resourceName}}",
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
		},
	}, nil
{{- else if eq . "Update"}}
	// TODO: Implement update logic
	return &{{$serviceLower}}V1.{{.}}{{$resourceName}}Response{
		{{$resourceName}}: &{{$serviceLower}}V1.{{$resourceName}}{
			Id:        req.Id,
			Name:      req.Name,
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
		},
	}, nil
{{- else if eq . "Delete"}}
	// TODO: Implement delete logic
	return &{{$serviceLower}}V1.{{.}}{{$resourceName}}Response{
		Success: true,
	}, nil
{{- else if eq . "List"}}
	// TODO: Implement list logic
	return &{{$serviceLower}}V1.{{.}}{{$resourceName}}Response{
		{{$resourceName}}s: []*{{$serviceLower}}V1.{{$resourceName}}{
			{
				Id:        "sample-1",
				Name:      "Sample {{$resourceName}} 1",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: timestamppb.Now(),
			},
		},
	}, nil
{{- end}}
}
{{- end}}
`

	tmpl, err := template.New("handler").Parse(handlerTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse handler template: %w", err)
	}

	// Prepare template data
	resourceName := strings.TrimSuffix(g.vars.ServiceName, "Service")
	data := struct {
		*templates.TemplateVars
		Methods           []string
		ResourceName      string
		ResourceNameLower string
	}{
		TemplateVars:      g.vars,
		Methods:           methods,
		ResourceName:      resourceName,
		ResourceNameLower: strings.ToLower(resourceName),
	}

	file, err := os.Create(handlerFile)
	if err != nil {
		return fmt.Errorf("failed to create handler file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute handler template: %w", err)
	}

	return nil
}

// GenerateWebHandler generates the web-side handler (stub for now)
func (g *HandlerGenerator) GenerateWebHandler(methods []string) error {
	// Create web handler file (optional, could be a simple client wrapper)
	handlerDir := filepath.Join("web", "handlers")
	handlerFile := filepath.Join(handlerDir, g.vars.ServiceNameLower+".go")

	webHandlerTemplate := `package handlers

import (
	"github.com/gofiber/fiber/v2"
	{{.ServiceNameLower}}V1 "{{.ModulePath}}/api/proto/{{.ServiceNameLower}}/v1"
)

type {{.ServiceName}} struct {
	*App
}

// TODO: Implement web handlers that call the gRPC service
// Example:
// func (h *{{.ServiceName}}) List{{.ResourceName}}(c *fiber.Ctx) error {
//     resp, err := h.API.{{.ServiceName}}.List{{.ResourceName}}(c.Context(), &{{.ServiceNameLower}}V1.List{{.ResourceName}}Request{})
//     if err != nil {
//         return err
//     }
//
//     return c.JSON(resp.{{.ResourceName}}s)
// }
`

	tmpl, err := template.New("webhandler").Parse(webHandlerTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse web handler template: %w", err)
	}

	resourceName := strings.TrimSuffix(g.vars.ServiceName, "Service")
	data := struct {
		*templates.TemplateVars
		ResourceName string
	}{
		TemplateVars: g.vars,
		ResourceName: resourceName,
	}

	file, err := os.Create(handlerFile)
	if err != nil {
		return fmt.Errorf("failed to create web handler file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute web handler template: %w", err)
	}

	return nil
}

// UpdateRoutes updates the route registration (stub for now)
func (g *HandlerGenerator) UpdateRoutes() error {
	// TODO: Parse and update the routing/routing.go file to register new routes
	// For now, just return a helpful error suggesting manual updates
	return fmt.Errorf("automatic route registration not yet implemented")
}
