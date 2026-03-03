package mcp

import (
	"context"

	"github.com/NP-compete/gomcp/internal/logger"
	"github.com/NP-compete/gomcp/internal/prompts"
	"github.com/NP-compete/gomcp/internal/resources"
	"github.com/NP-compete/gomcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerSDK wraps the official MCP SDK server
type ServerSDK struct {
	server *mcp.Server
}

// NewServerSDK creates a new MCP server using the official SDK
func NewServerSDK() *ServerSDK {
	// Create server with implementation info
	impl := &mcp.Implementation{
		Name:    "template",
		Version: "0.1.0",
	}

	server := mcp.NewServer(impl, nil)

	// Register all capabilities
	registerToolsSDK(server)
	registerPromptsSDK(server)
	registerResourcesSDK(server)
	registerRootsSDK(server)

	logger.Info("Template MCP Server initialized successfully with official SDK")

	return &ServerSDK{
		server: server,
	}
}

// registerToolsSDK registers all MCP tools using the official SDK pattern
func registerToolsSDK(server *mcp.Server) {
	// Register multiply_numbers tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "multiply_numbers",
			Description: "Multiply two numbers with comprehensive tool metadata. CPU-bound operation.",
		},
		tools.MultiplyNumbersSDK,
	)

	// Register generate_code_review_prompt tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "generate_code_review_prompt",
			Description: "Generate a structured code review prompt with comprehensive metadata. I/O-bound operation.",
		},
		tools.GenerateCodeReviewPromptSDK,
	)

	// Register get_redhat_logo tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "get_redhat_logo",
			Description: "Return the Red Hat logo as a base64 encoded string. Resource-as-tool pattern.",
		},
		tools.GetRedHatLogoSDK,
	)

	// Register long_operation tool (demonstrates cancellation)
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "long_operation",
			Description: "Simulates a long-running operation that supports cancellation.",
		},
		tools.LongOperationSDK,
	)

	logger.Info("MCP tools registered successfully with official SDK")
}

// registerPromptsSDK registers all MCP prompts using the official SDK pattern
func registerPromptsSDK(server *mcp.Server) {
	// Register code_review prompt
	server.AddPrompt(
		&mcp.Prompt{
			Name:        "code_review",
			Description: "Generate a structured code review prompt with analysis criteria",
		},
		func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			var args prompts.CodeReviewPromptArgs
			if req.Params.Arguments != nil {
				// Parse arguments
				if code, ok := req.Params.Arguments["code"]; ok {
					args.Code = code
				}
				if lang, ok := req.Params.Arguments["language"]; ok {
					args.Language = lang
				}
				if focus, ok := req.Params.Arguments["focus"]; ok {
					args.Focus = focus
				}
			}
			return prompts.GetCodeReviewPrompt(ctx, req, args)
		},
	)

	// Register git_commit prompt
	server.AddPrompt(
		&mcp.Prompt{
			Name:        "git_commit",
			Description: "Generate a git commit message from diff analysis",
		},
		func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			var args prompts.GitCommitPromptArgs
			if req.Params.Arguments != nil {
				if diff, ok := req.Params.Arguments["diff"]; ok {
					args.Diff = diff
				}
				if typ, ok := req.Params.Arguments["type"]; ok {
					args.Type = typ
				}
				if scope, ok := req.Params.Arguments["scope"]; ok {
					args.Scope = scope
				}
				if conv, ok := req.Params.Arguments["conventional"]; ok {
					args.Conventional = conv == "true"
				}
			}
			return prompts.GetGitCommitPrompt(ctx, req, args)
		},
	)

	// Register debug_help prompt
	server.AddPrompt(
		&mcp.Prompt{
			Name:        "debug_help",
			Description: "Get debugging assistance with root cause analysis",
		},
		func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			var args prompts.DebugHelpPromptArgs
			if req.Params.Arguments != nil {
				if err, ok := req.Params.Arguments["error"]; ok {
					args.Error = err
				}
				if code, ok := req.Params.Arguments["code"]; ok {
					args.Code = code
				}
				if lang, ok := req.Params.Arguments["language"]; ok {
					args.Language = lang
				}
				if context, ok := req.Params.Arguments["context"]; ok {
					args.Context = context
				}
			}
			return prompts.GetDebugHelpPrompt(ctx, req, args)
		},
	)

	logger.Info("MCP prompts registered successfully with official SDK")
}

// registerResourcesSDK registers all MCP resources using the official SDK pattern
func registerResourcesSDK(server *mcp.Server) {
	// Register project info resource
	server.AddResource(
		&mcp.Resource{
			URI:         "project://info",
			Name:        "Project Information",
			Description: "Information about the gomcp server (version, features, capabilities)",
			MIMEType:    "application/json",
		},
		resources.GetProjectInfo,
	)

	// Register system status resource
	server.AddResource(
		&mcp.Resource{
			URI:         "system://status",
			Name:        "System Status",
			Description: "Current system status and resource usage",
			MIMEType:    "application/json",
		},
		resources.GetSystemStatus,
	)

	// Register quick start guide resource
	server.AddResource(
		&mcp.Resource{
			URI:         "docs://quickstart",
			Name:        "Quick Start Guide",
			Description: "Quick start guide for using gomcp",
			MIMEType:    "text/markdown",
		},
		resources.GetQuickStartGuide,
	)

	// Register API reference resource
	server.AddResource(
		&mcp.Resource{
			URI:         "docs://api",
			Name:        "API Reference",
			Description: "Complete API endpoint documentation",
			MIMEType:    "text/markdown",
		},
		resources.GetAPIReference,
	)

	// Register config template resource
	server.AddResource(
		&mcp.Resource{
			URI:         "config://template",
			Name:        "Configuration Template",
			Description: "Configuration template with all available options",
			MIMEType:    "application/json",
		},
		resources.GetConfigTemplate,
	)

	// Register env vars reference resource
	server.AddResource(
		&mcp.Resource{
			URI:         "config://env-vars",
			Name:        "Environment Variables",
			Description: "Complete reference of environment variables",
			MIMEType:    "application/json",
		},
		resources.GetEnvVarsReference,
	)

	logger.Info("MCP resources registered successfully with official SDK")
}

// registerRootsSDK registers filesystem roots with the official SDK server
func registerRootsSDK(server *mcp.Server) {
	// Note: Roots support is implemented via custom handler in Cursor compatibility mode
	// The official SDK doesn't yet expose a direct roots registration API
	// Roots are available through the Cursor custom handler (getCursorCompatibleHandler)

	logger.Info("MCP roots prepared (available via Cursor handler)")
}

// GetServer returns the underlying official SDK server
func (s *ServerSDK) GetServer() *mcp.Server {
	return s.server
}

// HandleHTTPRequest processes an MCP request via HTTP
// This is a bridge between HTTP transport and SDK's session-based model
func (s *ServerSDK) HandleHTTPRequest(ctx context.Context, requestBody []byte) ([]byte, error) {
	// The official SDK expects session-based communication, not raw HTTP request/response
	// For HTTP transport, we need to bridge this gap
	// This will be implemented when we update the HTTP handler
	return nil, nil
}
