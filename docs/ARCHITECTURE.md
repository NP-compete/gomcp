# Architecture

This document describes the architecture of gomcp, a production-ready MCP (Model Context Protocol) server template in Go.

## Overview

gomcp implements the [MCP 2025-06-18 specification](https://modelcontextprotocol.io/specification/2025-06-18), providing a complete server implementation that can be used as a template for building custom MCP servers.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              MCP Clients                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Cursor IDE  │  │   Claude    │  │  HTTP/SSE   │  │   Custom Clients    │ │
│  │  (HTTP/SSE) │  │  Desktop    │  │   Clients   │  │                     │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘ │
└─────────┼────────────────┼────────────────┼────────────────────┼────────────┘
          │                │                │                    │
          ▼                ▼                ▼                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Transport Layer                                    │
│  ┌─────────────────────────────┐  ┌─────────────────────────────────────┐   │
│  │      HTTP/SSE Transport     │  │         stdio Transport             │   │
│  │   (Cursor, Web Clients)     │  │      (Claude Desktop, CLI)          │   │
│  └──────────────┬──────────────┘  └─────────────────┬───────────────────┘   │
└─────────────────┼───────────────────────────────────┼───────────────────────┘
                  │                                   │
                  ▼                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              MCP Server Core                                 │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                    Official MCP Go SDK                               │    │
│  │              (github.com/modelcontextprotocol/go-sdk)               │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                    │                                         │
│         ┌──────────────────────────┼──────────────────────────┐             │
│         ▼                          ▼                          ▼             │
│  ┌─────────────┐          ┌─────────────┐          ┌─────────────┐          │
│  │    Tools    │          │   Prompts   │          │  Resources  │          │
│  │  (4 impl)   │          │  (3 impl)   │          │  (6 impl)   │          │
│  └─────────────┘          └─────────────┘          └─────────────┘          │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                        MCP Features                                  │    │
│  │  ┌────────┐ ┌────────┐ ┌──────────┐ ┌────────┐ ┌──────────────────┐ │    │
│  │  │ Roots  │ │Logging │ │Pagination│ │Sampling│ │   Elicitation    │ │    │
│  │  └────────┘ └────────┘ └──────────┘ └────────┘ └──────────────────┘ │    │
│  │  ┌────────┐ ┌────────┐ ┌──────────┐ ┌────────┐                      │    │
│  │  │Progress│ │ Cancel │ │Completion│ │  Ping  │                      │    │
│  │  └────────┘ └────────┘ └──────────┘ └────────┘                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
gomcp/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── router.go            # HTTP router setup
│   │   ├── handlers.go          # HTTP request handlers
│   │   ├── metrics.go           # Metrics endpoint
│   │   └── mcp_metrics.go       # MCP-specific metrics
│   ├── completion/
│   │   ├── completion.go        # Structured output completion
│   │   └── completion_test.go   # Tests
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── elicitation/
│   │   └── elicitation.go       # User data elicitation
│   ├── errors/
│   │   └── errors.go            # Error types and handling
│   ├── logger/
│   │   └── logger.go            # Application logging
│   ├── logging/
│   │   ├── logging.go           # MCP logging (server-to-client)
│   │   └── logging_test.go      # Tests
│   ├── mcp/
│   │   ├── server.go            # MCP server (HTTP transport)
│   │   └── server_sdk.go        # MCP server (SDK integration)
│   ├── middleware/
│   │   ├── auth.go              # Authentication middleware
│   │   ├── cors.go              # CORS middleware
│   │   ├── logging.go           # Request logging
│   │   ├── recovery.go          # Panic recovery
│   │   ├── requestid.go         # Request ID generation
│   │   └── session.go           # Session management
│   ├── oauth/
│   │   └── oauth.go             # OAuth service
│   ├── pagination/
│   │   ├── pagination.go        # Cursor-based pagination
│   │   └── pagination_test.go   # Tests
│   ├── prompts/
│   │   ├── code_review.go       # Code review prompt
│   │   ├── debug_help.go        # Debug help prompt
│   │   └── git_commit.go        # Git commit prompt
│   ├── resources/
│   │   ├── config.go            # Config resources
│   │   ├── docs.go              # Documentation resources
│   │   └── project_info.go      # Project info resources
│   ├── roots/
│   │   ├── roots.go             # Filesystem roots
│   │   └── roots_test.go        # Tests
│   ├── sampling/
│   │   └── sampling.go          # LLM sampling requests
│   ├── storage/
│   │   ├── interface.go         # Storage interface
│   │   ├── memory.go            # In-memory storage
│   │   └── storage.go           # Storage implementation
│   ├── tools/
│   │   ├── codereview.go        # Code review tool
│   │   ├── codereview_sdk.go    # SDK integration
│   │   ├── logo.go              # Logo tool
│   │   ├── logo_sdk.go          # SDK integration
│   │   ├── long_operation_sdk.go# Long operation tool
│   │   ├── multiply.go          # Multiply tool
│   │   └── multiply_sdk.go      # SDK integration
│   └── version/
│       └── version.go           # Version information
├── pkg/
│   └── mcpprotocol/
│       ├── handler.go           # Protocol handler
│       └── types.go             # Protocol types
├── scripts/
│   ├── health-check.sh          # Health check script
│   ├── init-db.sql              # Database initialization
│   ├── run_dev.sh               # Development runner
│   ├── start.sh                 # Start script
│   ├── stop.sh                  # Stop script
│   └── test_all.sh              # Test runner
├── test/
│   └── integration_test.go      # Integration tests
└── docs/
    ├── ARCHITECTURE.md          # This file
    └── CUSTOMIZATION.md         # Template customization guide
```

## Core Components

### 1. Transport Layer

gomcp supports two transport mechanisms:

#### HTTP/SSE Transport
- Used by Cursor IDE and web-based clients
- Server-Sent Events (SSE) for real-time communication
- Cursor compatibility mode for IDE integration
- CORS support for browser clients

#### stdio Transport
- Used by Claude Desktop and CLI tools
- Standard input/output communication
- No network configuration required

### 2. MCP Server Core

The server core is built on the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk):

```go
// Server initialization
impl := &mcp.Implementation{
    Name:    "template",
    Version: "0.1.0",
}
server := mcp.NewServer(impl, nil)
```

### 3. Tools

Tools are functions that AI models can invoke:

| Tool | Description | Type |
|------|-------------|------|
| `multiply_numbers` | Multiply two numbers | CPU-bound |
| `code_review` | Generate code review | I/O-bound |
| `logo` | Return server logo | Resource-as-tool |
| `long_operation` | Demo progress/cancel | Long-running |

**Tool Implementation Pattern:**

```go
type ToolInput struct {
    Param string `json:"param" jsonschema:"required,description=..."`
}

type ToolOutput struct {
    Result string `json:"result"`
}

func ToolSDK(ctx context.Context, req *mcp.CallToolRequest, input ToolInput) (*mcp.CallToolResult, ToolOutput, error) {
    // Implementation
    return nil, output, nil
}
```

### 4. Prompts

Prompts are reusable templates for AI interactions:

| Prompt | Description |
|--------|-------------|
| `code_review` | Structured code review template |
| `git_commit` | Conventional commit message generator |
| `debug_help` | Debugging assistance template |

### 5. Resources

Resources provide data to AI models:

| Resource | URI | Type |
|----------|-----|------|
| Project Info | `project://info` | Dynamic |
| System Status | `system://status` | Dynamic |
| Quick Start | `docs://quickstart` | Static |
| API Reference | `docs://api` | Static |
| Config Template | `config://template` | Static |
| Env Vars | `config://env-vars` | Static |

### 6. MCP Features

| Feature | Description | Implementation |
|---------|-------------|----------------|
| **Roots** | Filesystem root definitions | `internal/roots/` |
| **Completion** | Structured tool outputs | `internal/completion/` |
| **Logging** | Server-to-client notifications | `internal/logging/` |
| **Pagination** | Cursor-based pagination | `internal/pagination/` |
| **Sampling** | Server-to-client LLM requests | `internal/sampling/` |
| **Elicitation** | Server-to-user data requests | `internal/elicitation/` |
| **Progress** | Real-time progress updates | Built into tools |
| **Cancellation** | Request cancellation | Context-based |
| **Ping** | Health checks | Built-in |

## Data Flow

### HTTP/SSE Request Flow

```
Client Request
      │
      ▼
┌─────────────┐
│   Router    │ ─── Middleware Chain ───┐
└─────────────┘                         │
      │                                 │
      ▼                                 ▼
┌─────────────┐                  ┌─────────────┐
│  Handlers   │                  │   Logging   │
└─────────────┘                  │    CORS     │
      │                          │   Auth      │
      ▼                          │  Recovery   │
┌─────────────┐                  └─────────────┘
│ MCP Server  │
└─────────────┘
      │
      ├──────────┬──────────┬──────────┐
      ▼          ▼          ▼          ▼
   Tools     Prompts   Resources    Features
```

### stdio Request Flow

```
stdin ──► MCP SDK Server ──► Tools/Prompts/Resources ──► stdout
```

## Configuration

Configuration is managed through environment variables:

```go
type Config struct {
    // Server
    MCPHost              string
    MCPPort              int
    MCPTransportProtocol string  // "http", "sse", "stdio"
    
    // Security
    EnableAuth           bool
    SessionSecret        string
    
    // Cursor Compatibility
    CursorCompatibleSSE  bool
    
    // SSL/TLS
    MCPSSLKeyfile        string
    MCPSSLCertfile       string
}
```

## Extension Points

### Adding a New Tool

1. Create `internal/tools/newtool_sdk.go`
2. Define input/output structs with JSON schema tags
3. Implement the tool function
4. Register in `internal/mcp/server_sdk.go`
5. Add tests

### Adding a New Prompt

1. Create `internal/prompts/newprompt.go`
2. Define argument struct
3. Implement the prompt function
4. Register in `internal/mcp/server_sdk.go`
5. Add tests

### Adding a New Resource

1. Create `internal/resources/newresource.go`
2. Implement the resource function
3. Register in `internal/mcp/server_sdk.go`
4. Add tests

## Security Considerations

1. **Authentication**: OAuth support with session management
2. **Transport Security**: SSL/TLS support for HTTP transport
3. **Input Validation**: All inputs validated before processing
4. **Error Handling**: Sensitive information not exposed in errors
5. **Container Security**: Non-root user in Docker

## Performance

- **Connection Pooling**: Database connections pooled
- **Caching**: Resource caching where appropriate
- **Pagination**: Cursor-based pagination for large datasets
- **Timeouts**: Configurable request timeouts

## Testing

```bash
# Unit tests
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...

# Integration tests
go test -v ./test/...

# Benchmarks
go test -bench=. ./...
```

## Monitoring

- **Health Check**: `GET /health`
- **Metrics**: `GET /metrics`
- **Logging**: Structured JSON logging with zerolog

## Related Documentation

- [MCP Specification](https://modelcontextprotocol.io/specification/2025-06-18)
- [Official Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [Customization Guide](CUSTOMIZATION.md)
