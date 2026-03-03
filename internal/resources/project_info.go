package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/NP-compete/gomcp/internal/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetProjectInfo returns information about the gomcp server
func GetProjectInfo(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	info := version.Get()

	projectInfo := map[string]interface{}{
		"name":        "gomcp - Go MCP Server",
		"description": "Production-ready Model Context Protocol server implementation in Go",
		"version":     info.Version,
		"git_commit":  info.GitCommit,
		"build_time":  info.BuildTime,
		"go_version":  info.GoVersion,
		"platform": map[string]string{
			"os":   info.OS,
			"arch": info.Arch,
		},
		"features": []string{
			"Official MCP Go SDK integration",
			"Dual endpoints (legacy + SDK)",
			"Tool call logging with request IDs",
			"Endpoint usage monitoring",
			"PostgreSQL or in-memory storage",
			"OAuth2 with PKCE",
			"Comprehensive middleware stack",
		},
		"capabilities": map[string]bool{
			"tools":     true,
			"prompts":   true,
			"resources": true,
		},
	}

	jsonData, err := json.MarshalIndent(projectInfo, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal project info: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      "project://info",
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}

// GetSystemStatus returns current system status
func GetSystemStatus(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	status := map[string]interface{}{
		"status": "operational",
		"system": map[string]interface{}{
			"goroutines": runtime.NumGoroutine(),
			"cpu_cores":  runtime.NumCPU(),
			"memory": map[string]interface{}{
				"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
				"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
				"sys_mb":         float64(m.Sys) / 1024 / 1024,
				"num_gc":         m.NumGC,
			},
		},
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	}

	jsonData, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal system status: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      "system://status",
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}
