<p align="center">
  <img src="https://img.shields.io/badge/MCP-2025--06--18-blue?style=for-the-badge" alt="MCP Version">
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/github/license/NP-compete/gomcp?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/github/v/release/NP-compete/gomcp?style=for-the-badge" alt="Release">
</p>

<p align="center">
  <a href="https://github.com/NP-compete/gomcp/actions/workflows/ci.yml">
    <img src="https://github.com/NP-compete/gomcp/actions/workflows/ci.yml/badge.svg" alt="CI">
  </a>
  <a href="https://codecov.io/gh/NP-compete/gomcp">
    <img src="https://codecov.io/gh/NP-compete/gomcp/branch/main/graph/badge.svg" alt="Coverage">
  </a>
  <a href="https://goreportcard.com/report/github.com/NP-compete/gomcp">
    <img src="https://goreportcard.com/badge/github.com/NP-compete/gomcp" alt="Go Report Card">
  </a>
</p>

<h1 align="center">gomcp</h1>

<p align="center">
  <strong>Production-ready Model Context Protocol (MCP) server template in Go</strong>
</p>

<p align="center">
  <a href="#features">Features</a> |
  <a href="#quick-start">Quick Start</a> |
  <a href="#use-as-template">Use as Template</a> |
  <a href="#documentation">Documentation</a> |
  <a href="#contributing">Contributing</a>
</p>

---

## Overview

**gomcp** is a complete, production-ready MCP server template built with Go and the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk). It implements the full [MCP 2025-06-18 specification](https://modelcontextprotocol.io/specification/2025-06-18) and is designed to be used as a starting point for building your own MCP servers.

### Why gomcp?

- **Complete Implementation** - All 12 MCP features fully implemented
- **Production Ready** - Docker support, CI/CD, security best practices
- **Well Documented** - Comprehensive docs, examples, and customization guide
- **Easy to Customize** - Clean architecture, modular design, extensive comments
- **Multiple Transports** - HTTP/SSE for Cursor IDE, stdio for Claude Desktop

## Features

<table>
<tr>
<td width="50%">

### MCP Features (12/12)

| Feature | Status | Description |
|---------|:------:|-------------|
| Tools | Yes | 4 example tools with structured outputs |
| Prompts | Yes | 3 reusable prompt templates |
| Resources | Yes | 6 static/dynamic resources |
| Roots | Yes | Filesystem root definitions |
| Completion | Yes | Structured outputs with JSON schemas |
| Logging | Yes | Server-to-client notifications (8 levels) |
| Pagination | Yes | Cursor-based pagination |
| Sampling | Yes | Server-to-client LLM requests |
| Elicitation | Yes | Server-to-user data requests |
| Progress | Yes | Real-time progress notifications |
| Cancellation | Yes | Request cancellation support |
| Ping | Yes | Health checks |

</td>
<td width="50%">

### Infrastructure

| Feature | Description |
|---------|-------------|
| **Multi-Transport** | HTTP/SSE + stdio support |
| **Cursor Compatible** | First-class Cursor IDE support |
| **Claude Desktop** | stdio transport for local use |
| **Docker Ready** | Multi-stage Dockerfile included |
| **CI/CD** | GitHub Actions workflows |
| **Hot Reload** | Development mode with air |
| **OAuth Support** | Authentication ready |
| **Metrics** | Built-in monitoring endpoints |
| **Security** | Non-root containers, SSL/TLS |

</td>
</tr>
</table>

## Quick Start

### Prerequisites

- Go 1.24+
- (Optional) Docker/Podman

### Installation

```bash
# Clone the repository
git clone https://github.com/NP-compete/gomcp.git
cd gomcp

# Download dependencies
go mod download

# Build and run
make run
```

### Running the Server

<details>
<summary><strong>For Cursor IDE</strong> (HTTP/SSE)</summary>

```bash
make cursor
# Server runs on http://localhost:8081/mcp/sse
```

Configure `~/.cursor/mcp.json`:
```json
{
  "mcpServers": {
    "gomcp": {
      "url": "http://localhost:8081/mcp/sse"
    }
  }
}
```

Restart Cursor IDE to connect.

</details>

<details>
<summary><strong>For Claude Desktop</strong> (stdio)</summary>

```bash
# Build the binary
make build-prod

# Configure Claude Desktop
# ~/Library/Application Support/Claude/claude_desktop_config.json (macOS)
# %APPDATA%\Claude\claude_desktop_config.json (Windows)
```

```json
{
  "mcpServers": {
    "gomcp": {
      "command": "/absolute/path/to/gomcp/bin/gomcp",
      "args": [],
      "env": {
        "MCP_TRANSPORT_PROTOCOL": "stdio"
      }
    }
  }
}
```

Restart Claude Desktop to connect.

</details>

<details>
<summary><strong>With Docker</strong></summary>

```bash
# Build and run with Docker
make docker-build
make docker-run

# Or with Docker Compose
docker-compose up -d
```

</details>

## Use as Template

This repository is designed to be used as a template for your own MCP server.

### Option 1: GitHub Template (Recommended)

Click the **"Use this template"** button on GitHub to create a new repository.

### Option 2: Manual Clone

```bash
# Clone and remove git history
git clone https://github.com/NP-compete/gomcp.git my-mcp-server
cd my-mcp-server
rm -rf .git
git init

# Update module name
# Edit go.mod: module github.com/YOUR_USERNAME/my-mcp-server
# Then update all imports
```

### Customization Guide

See [docs/CUSTOMIZATION.md](docs/CUSTOMIZATION.md) for detailed instructions on:

- Adding your own tools, prompts, and resources
- Configuring authentication
- Customizing the Docker setup
- Setting up CI/CD for your repository

## Project Structure

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
├── docs/               # Documentation
├── scripts/            # Utility scripts
├── .github/            # GitHub Actions & templates
├── Dockerfile          # Container build
├── docker-compose.yml  # Local development
└── Makefile           # Build commands
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_TRANSPORT_PROTOCOL` | `http` | Transport: `stdio`, `http`, `sse` |
| `MCP_PORT` | `8081` | Server port |
| `CURSOR_COMPATIBLE_SSE` | `true` | Enable Cursor compatibility |
| `ENABLE_AUTH` | `true` | Enable OAuth authentication |
| `LOG_LEVEL` | `INFO` | Log level |

See [.env.example](.env.example) for all options.

## Testing

```bash
# Run all tests
make test

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run linter
make lint

# Run security scan
make security-scan
```

## Documentation

| Document | Description |
|----------|-------------|
| [Architecture](docs/ARCHITECTURE.md) | System design and component overview |
| [Customization](docs/CUSTOMIZATION.md) | Guide to customizing the template |
| [Contributing](CONTRIBUTING.md) | How to contribute |
| [Security](SECURITY.md) | Security policy |

### External Resources

- [MCP Specification](https://modelcontextprotocol.io/specification/2025-06-18)
- [Official Go SDK](https://github.com/modelcontextprotocol/go-sdk)

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Model Context Protocol](https://modelcontextprotocol.io/) - The MCP specification
- [Official Go SDK](https://github.com/modelcontextprotocol/go-sdk) - The foundation of this server
- All [contributors](https://github.com/NP-compete/gomcp/graphs/contributors) who help improve this project

---

<p align="center">
  <strong>Built for the MCP community</strong>
</p>

<p align="center">
  <a href="https://github.com/NP-compete/gomcp/stargazers">Star us on GitHub</a> |
  <a href="https://github.com/NP-compete/gomcp/issues">Report Bug</a> |
  <a href="https://github.com/NP-compete/gomcp/discussions">Discussions</a>
</p>
