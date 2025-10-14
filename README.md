# gomcp - Go MCP Server Template

Production-ready Model Context Protocol (MCP) server template in Go with **complete MCP 2025-06-18 support**.

## ✨ Features

**12 MCP Features Fully Implemented:**

| Feature | Description |
|---------|-------------|
| **Tools** | 4 example tools with structured outputs |
| **Prompts** | 3 reusable prompt templates |
| **Resources** | 6 static/dynamic resources |
| **Roots** | Filesystem root definitions |
| **Completion** | Structured tool outputs with JSON schemas |
| **Logging** | Server→client log notifications (8 levels) |
| **Pagination** | Cursor-based pagination (max 100/page) |
| **Sampling** | Server→client LLM requests |
| **Elicitation** | Server→user data requests |
| **Progress** | Real-time progress notifications |
| **Cancellation** | Request cancellation support |
| **Ping** | Health checks |

## 🚀 Quick Start

### Prerequisites

- Go 1.23+
- (Optional) Docker/Podman

**Important:** Ensure Go's bin directory is in your PATH:

```bash
# Add to PATH (required for air, development tools)
export PATH=$PATH:$(go env GOPATH)/bin

# Make it permanent (add to ~/.zshrc or ~/.bashrc):
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

### 1. Clone & Install

```bash
git clone https://github.com/NP-compete/gomcp.git
cd gomcp
go mod download
```

### 2. Run Server

**For Cursor IDE:**
```bash
make cursor
# Server runs on http://localhost:8081/mcp/sse
```

**For Claude Desktop:**
```bash
export MCP_TRANSPORT_PROTOCOL=stdio
go run cmd/server/main.go
```

**Default (HTTP):**
```bash
make run
# Server runs on http://localhost:8081
```

### 3. Configure Client

**Cursor IDE** (`~/.cursor/mcp.json`):
```json
{
  "mcpServers": {
    "gomcp": {
      "url": "http://localhost:8081/mcp/sse"
    }
  }
}
```

**Claude Desktop** (`~/Library/Application Support/Claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "gomcp": {
      "command": "/path/to/gomcp/bin/gomcp",
      "args": [],
      "env": {
        "MCP_TRANSPORT_PROTOCOL": "stdio"
      }
    }
  }
}
```

## 📦 What's Included

### Example Tools
- `multiply_numbers` - Number multiplication with structured output
- `code_review` - Generate code review analysis
- `logo` - Display server logo
- `long_operation` - Demonstrates progress & cancellation

### Example Prompts
- `code_review` - Code review template
- `git_commit` - Git commit message generator
- `debug_help` - Debugging assistance

### Example Resources
- `project://info` - Server information
- `project://status` - System status
- `docs://quickstart` - Quick start guide
- `docs://api-reference` - API documentation
- `config://template` - Configuration template
- `config://env-vars` - Environment variables

## 🔧 Configuration

**Environment Variables:**

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_TRANSPORT_PROTOCOL` | `http` | Transport: `stdio`, `http`, `sse` |
| `MCP_PORT` | `8081` | Server port |
| `CURSOR_COMPATIBLE_SSE` | `false` | Enable Cursor compatibility |
| `ENABLE_AUTH` | `true` | Enable OAuth authentication |
| `LOG_LEVEL` | `INFO` | Log level |

Create `.env` file:
```bash
MCP_TRANSPORT_PROTOCOL=http
MCP_PORT=8081
CURSOR_COMPATIBLE_SSE=true
ENABLE_AUTH=false
```

## 🧪 Testing

```bash
# Run all tests
./scripts/test_all.sh

# Run specific tests
go test -v ./internal/completion
go test -v ./internal/logging
go test -v ./internal/pagination

# Generate coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

**Test Coverage:** 65 tests, 100% pass rate

## 📁 Project Structure

```
gomcp/
├── cmd/server/          # Main application entry point
├── internal/
│   ├── api/            # HTTP handlers & routing
│   ├── completion/     # Structured outputs
│   ├── logging/        # Server→client logs
│   ├── pagination/     # Cursor-based pagination
│   ├── prompts/        # Prompt implementations
│   ├── resources/      # Resource implementations
│   ├── roots/          # Filesystem roots
│   ├── tools/          # Tool implementations
│   ├── mcp/            # MCP server logic
│   └── config/         # Configuration management
├── pkg/mcpprotocol/    # MCP protocol implementation
├── test/               # Integration tests
└── Makefile           # Build & run commands
```

## 🛠️ Development

### Hot Reload
```bash
make dev  # Auto-restarts on file changes
```

### Build Commands
```bash
make build          # Development build
make build-prod     # Production build with optimization
make clean          # Clean artifacts
make deps           # Update dependencies
```

### Docker
```bash
make docker-build   # Build image
make docker-run     # Run container
```

## 🎯 Adding Features

### 1. Add a Tool

Create `internal/tools/mytool_sdk.go`:
```go
type MyToolInput struct {
    Param string `json:"param" jsonschema:"required,parameter description"`
}

type MyToolOutput struct {
    Result string `json:"result"`
}

func MyTool(ctx context.Context, req *mcp.CallToolRequest, input MyToolInput) (*mcp.CallToolResult, MyToolOutput, error) {
    // Your logic here
    output := MyToolOutput{Result: "success"}
    return nil, output, nil
}
```

Register in `internal/mcp/server_sdk.go`:
```go
server.AddTool(mcp.NewTool("mytool", "Description", MyTool))
```

### 2. Add a Prompt

Create `internal/prompts/myprompt.go`:
```go
func MyPrompt(ctx context.Context, args mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
    return &mcp.GetPromptResult{
        Messages: []*mcp.PromptMessage{
            {Role: "user", Content: &mcp.TextContent{Text: "Prompt text"}},
        },
    }, nil
}
```

Register in `internal/mcp/server_sdk.go`.

### 3. Add a Resource

Create `internal/resources/myresource.go`:
```go
func MyResource(ctx context.Context, params mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
    return &mcp.ReadResourceResult{
        Contents: []*mcp.ResourceContents{{
            URI:  "my://resource",
            Text: "Content here",
        }},
    }, nil
}
```

Register in `internal/mcp/server_sdk.go`.

## 🔐 Authentication

Enable OAuth:
```bash
export ENABLE_AUTH=true
export POSTGRES_HOST=localhost
export POSTGRES_DB=mcp_db
```

Disable for development:
```bash
export ENABLE_AUTH=false
```

## 🚢 Deployment

### Production Build
```bash
make build-prod
# Binary: bin/gomcp
```

### Docker Deployment
```bash
docker build -t gomcp-server .
docker run -p 8081:8081 --env-file .env gomcp-server
```

### Environment Setup
```bash
# Set transport
export MCP_TRANSPORT_PROTOCOL=http  # or stdio
export MCP_PORT=8081
export CURSOR_COMPATIBLE_SSE=true   # For Cursor
export ENABLE_AUTH=false            # For development
```

## 📊 Monitoring

### Metrics Endpoint
```bash
curl http://localhost:8081/metrics
```

Returns:
- Request counts
- Tool usage
- Error rates
- Response times
- Client info

### Health Check
```bash
curl http://localhost:8081/health
```

## 🐛 Troubleshooting

**Port already in use:**
```bash
lsof -ti:8081 | xargs kill -9
```

**Cursor not connecting:**
1. Set `CURSOR_COMPATIBLE_SSE=true`
2. Use `make cursor`
3. Verify `~/.cursor/mcp.json` has correct URL

**Claude Desktop not working:**
1. Set `MCP_TRANSPORT_PROTOCOL=stdio`
2. Use absolute path to binary
3. Restart Claude Desktop

**Build errors:**
```bash
go mod tidy
go mod download
make clean && make build
```

## 📚 Resources

- [MCP Specification](https://modelcontextprotocol.io/specification/2025-06-18)
- [Official Go SDK](https://github.com/modelcontextprotocol/go-sdk)

## 📝 License

MIT License - see LICENSE file

## 🤝 Contributing

1. Fork the repository
2. Create feature branch
3. Add tests
4. Submit pull request

---

**Built with ❤️ using Go and the official MCP SDK**

Template ready for production use with all MCP 2025-06-18 features!
