package api

import (
	"sync"
	"time"
)

// MCPMetrics tracks usage metrics for MCP endpoints
type MCPMetrics struct {
	// Per-endpoint metrics
	LegacyEndpoint EndpointMetrics `json:"legacy_endpoint"` // /mcp
	SDKEndpoint    EndpointMetrics `json:"sdk_endpoint"`    // /mcp/sse

	// Tool-specific metrics
	ToolCalls map[string]*ToolMetrics `json:"tool_calls"`

	// First/Last seen timestamps
	StartTime time.Time `json:"start_time"`
	LastReset time.Time `json:"last_reset"`
}

// mcpMetricsInternal is the internal structure with mutex
type mcpMetricsInternal struct {
	mu      sync.RWMutex
	metrics MCPMetrics
}

// EndpointMetrics tracks metrics for a specific endpoint
type EndpointMetrics struct {
	TotalRequests     int64            `json:"total_requests"`
	ToolCalls         int64            `json:"tool_calls"`
	ToolsListCalls    int64            `json:"tools_list_calls"`
	InitializeCalls   int64            `json:"initialize_calls"`
	Errors            int64            `json:"errors"`
	LastRequestTime   time.Time        `json:"last_request_time"`
	FirstRequestTime  time.Time        `json:"first_request_time"`
	ClientUserAgents  map[string]int64 `json:"client_user_agents"` // User-Agent -> count
	ClientIPs         map[string]int64 `json:"client_ips"`         // IP -> count
	AvgResponseTime   float64          `json:"avg_response_time_ms"`
	TotalResponseTime float64          `json:"total_response_time_ms"`
}

// ToolMetrics tracks per-tool usage across both endpoints
type ToolMetrics struct {
	Name          string    `json:"name"`
	LegacyCalls   int64     `json:"legacy_calls"`
	SDKCalls      int64     `json:"sdk_calls"`
	TotalCalls    int64     `json:"total_calls"`
	Errors        int64     `json:"errors"`
	LastCallTime  time.Time `json:"last_call_time"`
	FirstCallTime time.Time `json:"first_call_time"`
}

// Global metrics instance
var mcpMetrics = &mcpMetricsInternal{
	metrics: MCPMetrics{
		ToolCalls: make(map[string]*ToolMetrics),
		StartTime: time.Now(),
		LastReset: time.Now(),
		LegacyEndpoint: EndpointMetrics{
			ClientUserAgents: make(map[string]int64),
			ClientIPs:        make(map[string]int64),
		},
		SDKEndpoint: EndpointMetrics{
			ClientUserAgents: make(map[string]int64),
			ClientIPs:        make(map[string]int64),
		},
	},
}

// GetMCPMetrics returns a copy of current MCP metrics
func GetMCPMetrics() MCPMetrics {
	mcpMetrics.mu.RLock()
	defer mcpMetrics.mu.RUnlock()

	// Create a deep copy to avoid race conditions
	metrics := MCPMetrics{
		StartTime:      mcpMetrics.metrics.StartTime,
		LastReset:      mcpMetrics.metrics.LastReset,
		LegacyEndpoint: copyEndpointMetrics(mcpMetrics.metrics.LegacyEndpoint),
		SDKEndpoint:    copyEndpointMetrics(mcpMetrics.metrics.SDKEndpoint),
		ToolCalls:      make(map[string]*ToolMetrics),
	}

	for name, tm := range mcpMetrics.metrics.ToolCalls {
		metrics.ToolCalls[name] = &ToolMetrics{
			Name:          tm.Name,
			LegacyCalls:   tm.LegacyCalls,
			SDKCalls:      tm.SDKCalls,
			TotalCalls:    tm.TotalCalls,
			Errors:        tm.Errors,
			LastCallTime:  tm.LastCallTime,
			FirstCallTime: tm.FirstCallTime,
		}
	}

	return metrics
}

func copyEndpointMetrics(em EndpointMetrics) EndpointMetrics {
	copied := EndpointMetrics{
		TotalRequests:     em.TotalRequests,
		ToolCalls:         em.ToolCalls,
		ToolsListCalls:    em.ToolsListCalls,
		InitializeCalls:   em.InitializeCalls,
		Errors:            em.Errors,
		LastRequestTime:   em.LastRequestTime,
		FirstRequestTime:  em.FirstRequestTime,
		AvgResponseTime:   em.AvgResponseTime,
		TotalResponseTime: em.TotalResponseTime,
		ClientUserAgents:  make(map[string]int64),
		ClientIPs:         make(map[string]int64),
	}

	for k, v := range em.ClientUserAgents {
		copied.ClientUserAgents[k] = v
	}
	for k, v := range em.ClientIPs {
		copied.ClientIPs[k] = v
	}

	return copied
}

// TrackLegacyRequest records a request to the legacy /mcp endpoint
func TrackLegacyRequest(method, userAgent, clientIP string, duration time.Duration, isError bool) {
	mcpMetrics.mu.Lock()
	defer mcpMetrics.mu.Unlock()

	now := time.Now()
	mcpMetrics.metrics.LegacyEndpoint.TotalRequests++
	mcpMetrics.metrics.LegacyEndpoint.LastRequestTime = now

	if mcpMetrics.metrics.LegacyEndpoint.FirstRequestTime.IsZero() {
		mcpMetrics.metrics.LegacyEndpoint.FirstRequestTime = now
	}

	// Track by method
	switch method {
	case "tools/call":
		mcpMetrics.metrics.LegacyEndpoint.ToolCalls++
	case "tools/list":
		mcpMetrics.metrics.LegacyEndpoint.ToolsListCalls++
	case "initialize":
		mcpMetrics.metrics.LegacyEndpoint.InitializeCalls++
	}

	if isError {
		mcpMetrics.metrics.LegacyEndpoint.Errors++
	}

	// Track client info
	if userAgent != "" {
		mcpMetrics.metrics.LegacyEndpoint.ClientUserAgents[userAgent]++
	}
	if clientIP != "" {
		mcpMetrics.metrics.LegacyEndpoint.ClientIPs[clientIP]++
	}

	// Update response time
	durationMs := float64(duration.Milliseconds())
	mcpMetrics.metrics.LegacyEndpoint.TotalResponseTime += durationMs
	if mcpMetrics.metrics.LegacyEndpoint.TotalRequests > 0 {
		mcpMetrics.metrics.LegacyEndpoint.AvgResponseTime = mcpMetrics.metrics.LegacyEndpoint.TotalResponseTime / float64(mcpMetrics.metrics.LegacyEndpoint.TotalRequests)
	}
}

// TrackSDKRequest records a request to the SDK /mcp/sse endpoint
func TrackSDKRequest(method, userAgent, clientIP string, duration time.Duration, isError bool) {
	mcpMetrics.mu.Lock()
	defer mcpMetrics.mu.Unlock()

	now := time.Now()
	mcpMetrics.metrics.SDKEndpoint.TotalRequests++
	mcpMetrics.metrics.SDKEndpoint.LastRequestTime = now

	if mcpMetrics.metrics.SDKEndpoint.FirstRequestTime.IsZero() {
		mcpMetrics.metrics.SDKEndpoint.FirstRequestTime = now
	}

	// Track by method
	switch method {
	case "tools/call":
		mcpMetrics.metrics.SDKEndpoint.ToolCalls++
	case "tools/list":
		mcpMetrics.metrics.SDKEndpoint.ToolsListCalls++
	case "initialize":
		mcpMetrics.metrics.SDKEndpoint.InitializeCalls++
	}

	if isError {
		mcpMetrics.metrics.SDKEndpoint.Errors++
	}

	// Track client info
	if userAgent != "" {
		mcpMetrics.metrics.SDKEndpoint.ClientUserAgents[userAgent]++
	}
	if clientIP != "" {
		mcpMetrics.metrics.SDKEndpoint.ClientIPs[clientIP]++
	}

	// Update response time
	durationMs := float64(duration.Milliseconds())
	mcpMetrics.metrics.SDKEndpoint.TotalResponseTime += durationMs
	if mcpMetrics.metrics.SDKEndpoint.TotalRequests > 0 {
		mcpMetrics.metrics.SDKEndpoint.AvgResponseTime = mcpMetrics.metrics.SDKEndpoint.TotalResponseTime / float64(mcpMetrics.metrics.SDKEndpoint.TotalRequests)
	}
}

// TrackToolCall records a tool call (from either endpoint)
func TrackToolCall(toolName string, isLegacy bool, isError bool) {
	mcpMetrics.mu.Lock()
	defer mcpMetrics.mu.Unlock()

	now := time.Now()

	// Get or create tool metrics
	tm, exists := mcpMetrics.metrics.ToolCalls[toolName]
	if !exists {
		tm = &ToolMetrics{
			Name:          toolName,
			FirstCallTime: now,
		}
		mcpMetrics.metrics.ToolCalls[toolName] = tm
	}

	// Update counts
	if isLegacy {
		tm.LegacyCalls++
	} else {
		tm.SDKCalls++
	}
	tm.TotalCalls++
	tm.LastCallTime = now

	if isError {
		tm.Errors++
	}
}

// ResetMCPMetrics resets all MCP metrics
func ResetMCPMetrics() {
	mcpMetrics.mu.Lock()
	defer mcpMetrics.mu.Unlock()

	mcpMetrics.metrics.LegacyEndpoint = EndpointMetrics{
		ClientUserAgents: make(map[string]int64),
		ClientIPs:        make(map[string]int64),
	}
	mcpMetrics.metrics.SDKEndpoint = EndpointMetrics{
		ClientUserAgents: make(map[string]int64),
		ClientIPs:        make(map[string]int64),
	}
	mcpMetrics.metrics.ToolCalls = make(map[string]*ToolMetrics)
	mcpMetrics.metrics.LastReset = time.Now()
}
