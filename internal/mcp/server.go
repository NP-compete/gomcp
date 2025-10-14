package mcp

import (
	"github.com/NP-compete/gomcp/internal/logger"
	"github.com/NP-compete/gomcp/internal/tools"
	"github.com/NP-compete/gomcp/pkg/mcpprotocol"
)

// Server is the main MCP server implementation
type Server struct {
	protocol *mcpprotocol.Server
}

// NewServer creates a new MCP server with all tools registered
func NewServer() *Server {
	// Get logger instance
	loggerInstance := logger.GetLogger()

	// Create protocol server with logger
	protocol := mcpprotocol.NewServer("template", "0.1.0", loggerInstance)

	// Register tools
	registerTools(protocol)

	logger.Info("Template MCP Server initialized successfully")

	return &Server{
		protocol: protocol,
	}
}

// registerTools registers all MCP tools
func registerTools(server *mcpprotocol.Server) {
	// Register multiply_numbers tool
	server.RegisterTool(
		mcpprotocol.ToolDefinition{
			Name:        "multiply_numbers",
			Description: "Multiply two numbers with comprehensive tool metadata. CPU-bound operation - uses def for computational tasks.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{
						"type":        "number",
						"description": "First number to multiply",
					},
					"b": map[string]interface{}{
						"type":        "number",
						"description": "Second number to multiply",
					},
				},
				"required": []string{"a", "b"},
			},
		},
		tools.MultiplyNumbers,
	)

	// Register generate_code_review_prompt tool
	server.RegisterTool(
		mcpprotocol.ToolDefinition{
			Name:        "generate_code_review_prompt",
			Description: "Generate a structured code review prompt with comprehensive metadata. I/O-bound operation - uses async def for external API calls.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "The source code to be reviewed",
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "Programming language of the code (default: go)",
						"default":     "go",
					},
				},
				"required": []string{"code"},
			},
		},
		tools.GenerateCodeReviewPrompt,
	)

	// Register get_redhat_logo tool
	server.RegisterTool(
		mcpprotocol.ToolDefinition{
			Name:        "get_redhat_logo",
			Description: "Return the Red Hat logo as a base64 encoded string. Resource-as-tool pattern - async def for file I/O operations.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		tools.GetRedHatLogo,
	)

	logger.Info("MCP tools registered successfully")
}

// HandleRequest processes an MCP JSON-RPC request
func (s *Server) HandleRequest(reqData []byte) ([]byte, error) {
	return s.protocol.HandleRequest(reqData)
}

// GetProtocolServer returns the underlying protocol server
func (s *Server) GetProtocolServer() *mcpprotocol.Server {
	return s.protocol
}
