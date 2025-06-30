package cli

// File and directory constants
const (
	// Project structure
	MarkerFileName      = ".meowed"
	DefaultModulePrefix = "github.com/user"
	
	// Template directories  
	ApiProtoDir     = "api/proto"
	WebHandlersDir  = "web/handlers"
	ApiHandlersDir  = "api/server/handlers"
	
	// File extensions
	ProtoExt     = ".proto"
	GoExt        = ".go"
	TemplateExt  = ".template"
	
	// Default server settings
	DefaultHTTPPort = "3000"
	DefaultGRPCPort = "50051"
)

// Marker file content
const MarkerFileContent = `🐱 This project has been MEOWED! 🐱

Congratulations! Your project was lovingly crafted by the Meower CLI.
You're now part of the exclusive club of developers who've been meowed.

May your code purr smoothly and your builds never hiss! 🚀

Generated with Meower Framework
https://github.com/AlyxPink/meower`