# Template Customization Guide

This guide walks you through customizing the gomcp template for your own MCP server.

## Quick Start

### 1. Use This Template

Click **"Use this template"** on GitHub to create your own repository, or:

```bash
# Clone the template
git clone https://github.com/NP-compete/gomcp.git my-mcp-server
cd my-mcp-server

# Remove git history and start fresh
rm -rf .git
git init
git add .
git commit -m "Initial commit from gomcp template"
```

### 2. Rename the Module

Update the Go module name in `go.mod`:

```go
// Before
module github.com/NP-compete/gomcp

// After
module github.com/YOUR_USERNAME/my-mcp-server
```

Then update all imports:

```bash
# Find and replace all import paths
find . -type f -name "*.go" -exec sed -i '' 's|github.com/NP-compete/gomcp|github.com/YOUR_USERNAME/my-mcp-server|g' {} +

# Update Makefile LDFLAGS
sed -i '' 's|github.com/NP-compete/gomcp|github.com/YOUR_USERNAME/my-mcp-server|g' Makefile
```

### 3. Update Server Identity

Edit `internal/mcp/server_sdk.go`:

```go
impl := &mcp.Implementation{
    Name:    "my-server",      // Your server name
    Version: "1.0.0",          // Your version
}
```

### 4. Customize Configuration

Update `.env.example` and `internal/config/config.go` with your settings.

## Adding Your Own Tools

### Step 1: Create the Tool File

Create `internal/tools/weather_sdk.go`:

```go
package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// WeatherInput defines the tool's input parameters
type WeatherInput struct {
    City    string `json:"city" jsonschema:"required,description=City name to get weather for"`
    Units   string `json:"units" jsonschema:"enum=celsius,enum=fahrenheit,default=celsius,description=Temperature units"`
}

// WeatherOutput defines the tool's output structure
type WeatherOutput struct {
    City        string  `json:"city"`
    Temperature float64 `json:"temperature"`
    Units       string  `json:"units"`
    Condition   string  `json:"condition"`
    Humidity    int     `json:"humidity"`
}

// GetWeatherSDK implements the weather lookup tool
func GetWeatherSDK(ctx context.Context, req *mcp.CallToolRequest, input WeatherInput) (*mcp.CallToolResult, WeatherOutput, error) {
    // Set default units
    if input.Units == "" {
        input.Units = "celsius"
    }

    // Your weather API logic here
    // This is a placeholder - replace with actual API call
    output := WeatherOutput{
        City:        input.City,
        Temperature: 22.5,
        Units:       input.Units,
        Condition:   "Sunny",
        Humidity:    65,
    }

    return nil, output, nil
}
```

### Step 2: Register the Tool

Edit `internal/mcp/server_sdk.go`:

```go
func registerToolsSDK(server *mcp.Server) {
    // ... existing tools ...

    // Register weather tool
    mcp.AddTool(
        server,
        &mcp.Tool{
            Name:        "get_weather",
            Description: "Get current weather for a city. Returns temperature, conditions, and humidity.",
        },
        tools.GetWeatherSDK,
    )
}
```

### Step 3: Add Tests

Create `internal/tools/weather_test.go`:

```go
package tools

import (
    "context"
    "testing"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestGetWeatherSDK(t *testing.T) {
    tests := []struct {
        name    string
        input   WeatherInput
        wantErr bool
    }{
        {
            name:    "valid city",
            input:   WeatherInput{City: "London", Units: "celsius"},
            wantErr: false,
        },
        {
            name:    "default units",
            input:   WeatherInput{City: "Paris"},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            req := &mcp.CallToolRequest{}
            
            _, output, err := GetWeatherSDK(ctx, req, tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("GetWeatherSDK() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if output.City != tt.input.City {
                t.Errorf("GetWeatherSDK() city = %v, want %v", output.City, tt.input.City)
            }
        })
    }
}
```

## Adding Your Own Prompts

### Step 1: Create the Prompt File

Create `internal/prompts/summarize.go`:

```go
package prompts

import (
    "context"
    "fmt"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// SummarizePromptArgs defines the prompt arguments
type SummarizePromptArgs struct {
    Text   string `json:"text"`
    Style  string `json:"style"`  // "brief", "detailed", "bullet"
    MaxLen int    `json:"max_length"`
}

// GetSummarizePrompt generates a text summarization prompt
func GetSummarizePrompt(ctx context.Context, req *mcp.GetPromptRequest, args SummarizePromptArgs) (*mcp.GetPromptResult, error) {
    // Set defaults
    if args.Style == "" {
        args.Style = "brief"
    }
    if args.MaxLen == 0 {
        args.MaxLen = 200
    }

    styleInstructions := map[string]string{
        "brief":    "Provide a concise summary in 1-2 sentences.",
        "detailed": "Provide a comprehensive summary covering all key points.",
        "bullet":   "Provide a summary as bullet points.",
    }

    promptText := fmt.Sprintf(`Summarize the following text.

Style: %s
Maximum Length: %d words

Instructions: %s

Text to summarize:
---
%s
---

Summary:`,
        args.Style,
        args.MaxLen,
        styleInstructions[args.Style],
        args.Text,
    )

    return &mcp.GetPromptResult{
        Messages: []*mcp.PromptMessage{
            {
                Role: "user",
                Content: &mcp.TextContent{
                    Type: "text",
                    Text: promptText,
                },
            },
        },
    }, nil
}
```

### Step 2: Register the Prompt

Edit `internal/mcp/server_sdk.go`:

```go
func registerPromptsSDK(server *mcp.Server) {
    // ... existing prompts ...

    // Register summarize prompt
    server.AddPrompt(
        &mcp.Prompt{
            Name:        "summarize",
            Description: "Generate a text summarization prompt with configurable style",
        },
        func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
            var args prompts.SummarizePromptArgs
            if req.Params.Arguments != nil {
                if text, ok := req.Params.Arguments["text"]; ok {
                    args.Text = text
                }
                if style, ok := req.Params.Arguments["style"]; ok {
                    args.Style = style
                }
                // Parse max_length as needed
            }
            return prompts.GetSummarizePrompt(ctx, req, args)
        },
    )
}
```

## Adding Your Own Resources

### Step 1: Create the Resource File

Create `internal/resources/database.go`:

```go
package resources

import (
    "context"
    "encoding/json"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// DatabaseSchema represents the database schema resource
type DatabaseSchema struct {
    Tables []TableInfo `json:"tables"`
}

type TableInfo struct {
    Name    string       `json:"name"`
    Columns []ColumnInfo `json:"columns"`
}

type ColumnInfo struct {
    Name     string `json:"name"`
    Type     string `json:"type"`
    Nullable bool   `json:"nullable"`
}

// GetDatabaseSchema returns the database schema as a resource
func GetDatabaseSchema(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
    // In a real implementation, query your database for schema info
    schema := DatabaseSchema{
        Tables: []TableInfo{
            {
                Name: "users",
                Columns: []ColumnInfo{
                    {Name: "id", Type: "uuid", Nullable: false},
                    {Name: "email", Type: "varchar", Nullable: false},
                    {Name: "created_at", Type: "timestamp", Nullable: false},
                },
            },
            // Add more tables...
        },
    }

    data, err := json.MarshalIndent(schema, "", "  ")
    if err != nil {
        return nil, err
    }

    return &mcp.ReadResourceResult{
        Contents: []*mcp.ResourceContents{
            {
                URI:      "database://schema",
                MIMEType: "application/json",
                Text:     string(data),
            },
        },
    }, nil
}
```

### Step 2: Register the Resource

Edit `internal/mcp/server_sdk.go`:

```go
func registerResourcesSDK(server *mcp.Server) {
    // ... existing resources ...

    // Register database schema resource
    server.AddResource(
        &mcp.Resource{
            URI:         "database://schema",
            Name:        "Database Schema",
            Description: "Current database schema with all tables and columns",
            MIMEType:    "application/json",
        },
        resources.GetDatabaseSchema,
    )
}
```

## Removing Example Code

To start with a clean slate, remove the example implementations:

```bash
# Remove example tools (keep the files as templates)
rm internal/tools/multiply*.go
rm internal/tools/codereview*.go
rm internal/tools/logo*.go
rm internal/tools/long_operation*.go

# Remove example prompts
rm internal/prompts/code_review.go
rm internal/prompts/debug_help.go
rm internal/prompts/git_commit.go

# Remove example resources
rm internal/resources/config.go
rm internal/resources/docs.go
rm internal/resources/project_info.go
```

Then update `internal/mcp/server_sdk.go` to remove the registrations.

## Configuration Customization

### Adding New Environment Variables

1. Add to `internal/config/config.go`:

```go
type Config struct {
    // ... existing fields ...
    
    // Your custom config
    MyAPIKey     string `mapstructure:"MY_API_KEY"`
    MyAPIBaseURL string `mapstructure:"MY_API_BASE_URL"`
}
```

2. Add to `.env.example`:

```env
# My API Configuration
MY_API_KEY=your-api-key-here
MY_API_BASE_URL=https://api.example.com
```

3. Add validation in `config.Validate()` if needed.

## Docker Customization

### Adding Dependencies

Edit `Dockerfile` to add system dependencies:

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

# Add your build dependencies
RUN apk add --no-cache git ca-certificates gcc musl-dev

# ... rest of build ...

# Final stage
FROM alpine:latest

# Add your runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# ... rest of runtime ...
```

### Multi-stage Builds for Additional Assets

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder
# ... build Go binary ...

# Asset stage (if you need to process assets)
FROM node:20-alpine AS assets
WORKDIR /assets
COPY frontend/ .
RUN npm ci && npm run build

# Final stage
FROM alpine:latest
COPY --from=builder /app/bin/gomcp /app/gomcp
COPY --from=assets /assets/dist /app/static
```

## CI/CD Customization

### Adding Custom Checks

Edit `.github/workflows/ci.yml`:

```yaml
jobs:
  # ... existing jobs ...

  custom-check:
    name: Custom Validation
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run custom validation
        run: |
          # Your custom checks here
          ./scripts/validate.sh
```

### Adding Deployment

Add to `.github/workflows/release.yml`:

```yaml
  deploy:
    name: Deploy to Production
    runs-on: ubuntu-latest
    needs: [build, docker]
    if: github.ref == 'refs/heads/main'
    steps:
      - name: Deploy
        run: |
          # Your deployment logic
```

## Testing Your Customizations

### Local Testing

```bash
# Run all tests
make test

# Test specific package
go test -v ./internal/tools/...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Testing

```bash
# Start server
make run

# In another terminal, test with curl
curl http://localhost:8081/health

# Test MCP endpoint
curl -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

### Testing with Cursor IDE

1. Build the server: `make cursor`
2. Configure `~/.cursor/mcp.json`:
   ```json
   {
     "mcpServers": {
       "my-server": {
         "url": "http://localhost:8081/mcp/sse"
       }
     }
   }
   ```
3. Restart Cursor IDE
4. Test your tools in the chat

## Checklist

Before publishing your customized server:

- [ ] Updated module name in `go.mod`
- [ ] Updated all import paths
- [ ] Changed server name and version
- [ ] Added your own tools/prompts/resources
- [ ] Removed or modified example code
- [ ] Updated README.md with your documentation
- [ ] Updated CHANGELOG.md
- [ ] Configured CI/CD for your repository
- [ ] Added appropriate LICENSE
- [ ] Tested with target clients (Cursor, Claude Desktop)

## Getting Help

- [MCP Specification](https://modelcontextprotocol.io/specification/2025-06-18)
- [Official Go SDK Documentation](https://github.com/modelcontextprotocol/go-sdk)
- [gomcp Issues](https://github.com/NP-compete/gomcp/issues)
- [gomcp Discussions](https://github.com/NP-compete/gomcp/discussions)
