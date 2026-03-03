# Contributing to gomcp

First off, thank you for considering contributing to gomcp! It's people like you that make gomcp such a great tool for the MCP community.

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible using our [bug report template](.github/ISSUE_TEMPLATE/bug_report.md).

**Great Bug Reports** tend to have:

- A quick summary and/or background
- Steps to reproduce (be specific!)
- What you expected would happen
- What actually happens
- Notes (possibly including why you think this might be happening)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, use our [feature request template](.github/ISSUE_TEMPLATE/feature_request.md) and include:

- A clear and descriptive title
- A detailed description of the proposed enhancement
- Explain why this enhancement would be useful
- List any alternatives you've considered

### Pull Requests

1. Fork the repo and create your branch from `main`
2. If you've added code that should be tested, add tests
3. If you've changed APIs, update the documentation
4. Ensure the test suite passes
5. Make sure your code lints
6. Issue that pull request!

## Development Setup

### Prerequisites

- Go 1.23 or higher
- Docker/Podman (optional, for containerized development)

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/gomcp.git
cd gomcp

# Add upstream remote
git remote add upstream https://github.com/NP-compete/gomcp.git

# Install dependencies
go mod download

# Install development tools
make tools

# Run tests
make test

# Run linter
make lint
```

### Development Workflow

```bash
# Create a feature branch
git checkout -b feature/my-new-feature

# Make your changes and test
make test
make lint

# Run the server locally
make run

# For Cursor IDE testing
make cursor
```

### Project Structure

```
gomcp/
├── cmd/server/          # Application entry point
├── internal/
│   ├── api/            # HTTP handlers & routing
│   ├── completion/     # Structured outputs
│   ├── config/         # Configuration management
│   ├── logging/        # Server-to-client logs
│   ├── mcp/            # MCP server logic
│   ├── pagination/     # Cursor-based pagination
│   ├── prompts/        # Prompt implementations
│   ├── resources/      # Resource implementations
│   ├── roots/          # Filesystem roots
│   └── tools/          # Tool implementations
├── pkg/mcpprotocol/    # MCP protocol types
├── test/               # Integration tests
└── docs/               # Documentation
```

## Adding New Features

### Adding a New Tool

1. Create `internal/tools/mytool_sdk.go`:

```go
package tools

import (
    "context"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

type MyToolInput struct {
    Param string `json:"param" jsonschema:"required,description=Parameter description"`
}

type MyToolOutput struct {
    Result string `json:"result"`
}

func MyToolSDK(ctx context.Context, req *mcp.CallToolRequest, input MyToolInput) (*mcp.CallToolResult, MyToolOutput, error) {
    output := MyToolOutput{Result: "success"}
    return nil, output, nil
}
```

2. Register in `internal/mcp/server_sdk.go`:

```go
mcp.AddTool(
    server,
    &mcp.Tool{
        Name:        "my_tool",
        Description: "Description of my tool",
    },
    tools.MyToolSDK,
)
```

3. Add tests in `internal/tools/mytool_test.go`

### Adding a New Prompt

1. Create `internal/prompts/myprompt.go`
2. Register in `internal/mcp/server_sdk.go`
3. Add tests

### Adding a New Resource

1. Create `internal/resources/myresource.go`
2. Register in `internal/mcp/server_sdk.go`
3. Add tests

## Style Guidelines

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting (run `make lint`)
- Write meaningful comments for exported functions
- Keep functions focused and small
- Handle errors explicitly

### Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Formatting, missing semicolons, etc.
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding missing tests
- `chore`: Maintenance tasks

Examples:
```
feat(tools): add weather lookup tool
fix(transport): handle SSE reconnection properly
docs(readme): add Cursor IDE setup instructions
```

### Testing

- Write tests for new functionality
- Maintain or improve code coverage
- Use table-driven tests where appropriate
- Mock external dependencies

```bash
# Run all tests
make test

# Run specific tests
go test -v ./internal/tools/...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Review Process

1. All submissions require review
2. We use GitHub pull requests for this purpose
3. Reviewers will look for:
   - Code quality and style
   - Test coverage
   - Documentation updates
   - Backward compatibility

## Community

- Use GitHub Discussions for questions and ideas
- Be respectful and constructive
- Help others when you can

## Recognition

Contributors are recognized in our release notes and README. Thank you for making gomcp better!

## Questions?

Feel free to open an issue with the `question` label or start a discussion.

---

Thank you for contributing!
