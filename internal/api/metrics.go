package api

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/NP-compete/gomcp/internal/version"
)

// Metrics represents application metrics
type Metrics struct {
	Version       string        `json:"version"`
	Uptime        string        `json:"uptime"`
	UptimeSeconds float64       `json:"uptime_seconds"`
	Goroutines    int           `json:"goroutines"`
	Memory        MemoryMetrics `json:"memory"`
	System        SystemMetrics `json:"system"`
	MCP           MCPMetrics    `json:"mcp"` // MCP endpoint usage metrics
	Timestamp     time.Time     `json:"timestamp"`
}

// MemoryMetrics represents memory usage metrics
type MemoryMetrics struct {
	Alloc      uint64  `json:"alloc_bytes"`
	TotalAlloc uint64  `json:"total_alloc_bytes"`
	Sys        uint64  `json:"sys_bytes"`
	NumGC      uint32  `json:"num_gc"`
	AllocMB    float64 `json:"alloc_mb"`
	SysMB      float64 `json:"sys_mb"`
}

// SystemMetrics represents system metrics
type SystemMetrics struct {
	NumCPU    int    `json:"num_cpu"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

var startTime = time.Now()

// MetricsHandler returns application metrics including MCP endpoint usage
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(startTime)
	versionInfo := version.Get()

	metrics := Metrics{
		Version:       versionInfo.Version,
		Uptime:        uptime.String(),
		UptimeSeconds: uptime.Seconds(),
		Goroutines:    runtime.NumGoroutine(),
		Memory: MemoryMetrics{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			Sys:        m.Sys,
			NumGC:      m.NumGC,
			AllocMB:    float64(m.Alloc) / 1024 / 1024,
			SysMB:      float64(m.Sys) / 1024 / 1024,
		},
		System: SystemMetrics{
			NumCPU:    runtime.NumCPU(),
			GoVersion: runtime.Version(),
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
		},
		MCP:       GetMCPMetrics(), // Include MCP endpoint usage metrics
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// VersionHandler returns version information
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(version.Get())
}
