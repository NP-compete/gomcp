package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/NP-compete/gomcp/internal/config"
	"github.com/NP-compete/gomcp/internal/logger"
	"github.com/NP-compete/gomcp/internal/mcp"
	"github.com/NP-compete/gomcp/internal/oauth"
	"github.com/NP-compete/gomcp/internal/pagination"
	"github.com/NP-compete/gomcp/internal/prompts"
	"github.com/NP-compete/gomcp/internal/resources"
	"github.com/NP-compete/gomcp/internal/roots"
	"github.com/NP-compete/gomcp/internal/storage"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents the API server
type Server struct {
	mcpServer    *mcp.Server    // Legacy custom protocol server
	mcpServerSDK *mcp.ServerSDK // Official SDK server
	oauthService *oauth.Service
	config       *config.Config

	// Request tracking for cancellation
	activeRequests map[interface{}]context.CancelFunc
	requestsMutex  sync.RWMutex
}

// NewServer creates a new API server with both legacy and SDK servers
func NewServer(mcpServer *mcp.Server, oauthService *oauth.Service, cfg *config.Config) *Server {
	return &Server{
		mcpServer:      mcpServer,
		mcpServerSDK:   mcp.NewServerSDK(), // Initialize official SDK server
		oauthService:   oauthService,
		config:         cfg,
		activeRequests: make(map[interface{}]context.CancelFunc),
	}
}

// HealthCheckHandler handles health check requests
func (s *Server) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":             "healthy",
		"service":            "template-mcp-server",
		"transport_protocol": s.config.MCPTransportProtocol,
		"version":            "0.1.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// MCPHandler handles MCP protocol requests (legacy custom protocol)
func (s *Server) MCPHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to read request body: %v", err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		TrackLegacyRequest("unknown", r.UserAgent(), r.RemoteAddr, time.Since(startTime), true)
		return
	}
	defer r.Body.Close()

	// Extract method from JSON-RPC request for metrics
	var jsonRPCReq map[string]interface{}
	method := "unknown"
	toolName := ""
	if err := json.Unmarshal(body, &jsonRPCReq); err == nil {
		if m, ok := jsonRPCReq["method"].(string); ok {
			method = m
		}
		// Extract tool name if this is a tools/call
		if method == "tools/call" {
			if params, ok := jsonRPCReq["params"].(map[string]interface{}); ok {
				if name, ok := params["name"].(string); ok {
					toolName = name
				}
			}
		}
	}

	// Process MCP request
	response, err := s.mcpServer.HandleRequest(body)
	isError := err != nil

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to handle MCP request: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		TrackLegacyRequest(method, r.UserAgent(), r.RemoteAddr, time.Since(startTime), true)
		return
	}

	// Track metrics
	duration := time.Since(startTime)
	TrackLegacyRequest(method, r.UserAgent(), r.RemoteAddr, duration, isError)

	// Track tool-specific metrics
	if toolName != "" {
		TrackToolCall(toolName, true, isError)
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// GetSSEHandler returns an SSE handler for the official SDK server with metrics tracking
func (s *Server) GetSSEHandler() http.Handler {
	// For Cursor compatibility, use custom handler instead of SDK
	if s.config.CursorCompatibleSSE {
		logger.Info("SSE handler initialized with Cursor custom protocol handler")
		return s.getCursorCompatibleHandler()
	}

	// Standard SDK handler for other clients
	logger.Info("SSE handler initialized for official SDK")
	getServer := func(req *http.Request) *mcpsdk.Server {
		return s.mcpServerSDK.GetServer()
	}
	handler := mcpsdk.NewSSEHandler(getServer, nil)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		handler.ServeHTTP(ww, r)

		method := "sse_" + r.Method
		if r.Method == "POST" {
			method = "sse_message"
		}
		duration := time.Since(startTime)
		isError := ww.statusCode >= 400
		TrackSDKRequest(method, r.UserAgent(), r.RemoteAddr, duration, isError)
	})
}

// getCursorCompatibleHandler returns a custom handler for Cursor IDE
func (s *Server) getCursorCompatibleHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Handle OPTIONS preflight
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Handle GET - SSE connection establishment
		if r.Method == "GET" {
			logger.Debug("Cursor SSE: GET request for connection establishment")

			// Set SSE headers
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("Access-Control-Allow-Origin", "*")

			// Send initial connection message
			fmt.Fprintf(w, "event: endpoint\ndata: /mcp/sse\n\n")

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

			// Keep connection alive
			<-r.Context().Done()
			return
		}

		// Handle POST - JSON-RPC requests
		if r.Method == "POST" {
			logger.Debug("Cursor SSE: POST request for JSON-RPC")

			// Read and parse body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to read request body: %v", err))
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			logger.Debug(fmt.Sprintf("Cursor request body: %s", string(bodyBytes)))

			// Parse JSON-RPC request
			var jsonRPCReq map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &jsonRPCReq); err != nil {
				logger.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
				// Return JSON-RPC parse error
				errorResponse := s.formatErrorResponse(nil, -32700, "Parse error", err.Error())
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK) // JSON-RPC errors use 200 OK
				w.Write(errorResponse)
				return
			}

			// Extract method and ID
			method, _ := jsonRPCReq["method"].(string)
			requestID := jsonRPCReq["id"]

			// Handle cancellation notifications
			if method == "notifications/cancelled" || method == "$/cancelRequest" {
				s.handleCancellation(jsonRPCReq)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Check if this is a long-running operation that needs progress
			params, _ := jsonRPCReq["params"].(map[string]interface{})
			toolName, _ := params["name"].(string)
			needsProgress := (method == "tools/call" && toolName == "long_operation")

			if needsProgress {
				// Stream response with progress for long operations
				s.handleLongOperationWithProgress(w, r, bodyBytes, requestID, startTime)
				return
			}

			// Route based on method (standard request/response)
			var response []byte
			switch method {
			case "initialize":
				// Handle initialize separately with updated capabilities
				response, err = s.handleInitializeForCursor(bodyBytes)
			case "prompts/list", "prompts/get", "resources/list", "resources/read", "roots/list":
				// These methods require SDK server - handle via SDK
				response, err = s.handleSDKMethodForCursor(r.Context(), bodyBytes, method, requestID)
			default:
				// Tools and other methods - use legacy server with cancellation support
				response, err = s.handleWithCancellation(r.Context(), bodyBytes, requestID)
			}

			if err != nil {
				logger.Error(fmt.Sprintf("MCP server error: %v", err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				TrackSDKRequest("cursor_post", r.UserAgent(), r.RemoteAddr, time.Since(startTime), true)
				return
			}

			// Send response
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusOK)
			w.Write(response)

			TrackSDKRequest("cursor_post", r.UserAgent(), r.RemoteAddr, time.Since(startTime), false)
			return
		}

		// Unsupported method
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
}

// handleCancellation handles cancellation requests from Cursor
func (s *Server) handleCancellation(req map[string]interface{}) {
	params, _ := req["params"].(map[string]interface{})
	cancelID := params["id"]

	s.requestsMutex.Lock()
	if cancelFunc, exists := s.activeRequests[cancelID]; exists {
		logger.Info(fmt.Sprintf("Cancelling request: %v", cancelID))
		cancelFunc()
		delete(s.activeRequests, cancelID)
	}
	s.requestsMutex.Unlock()
}

// handleWithCancellation handles requests with cancellation support
func (s *Server) handleWithCancellation(parentCtx context.Context, bodyBytes []byte, requestID interface{}) ([]byte, error) {
	// Create cancellable context
	_, cancel := context.WithCancel(parentCtx)
	defer cancel()

	// Track request for cancellation
	if requestID != nil {
		s.requestsMutex.Lock()
		s.activeRequests[requestID] = cancel
		s.requestsMutex.Unlock()

		defer func() {
			s.requestsMutex.Lock()
			delete(s.activeRequests, requestID)
			s.requestsMutex.Unlock()
		}()
	}

	// Execute request
	// Note: Legacy server doesn't support context cancellation yet
	// But the context is tracked, so external cancellation requests will work
	return s.mcpServer.HandleRequest(bodyBytes)
}

// handleLongOperationWithProgress streams progress updates for long operations
func (s *Server) handleLongOperationWithProgress(w http.ResponseWriter, r *http.Request, bodyBytes []byte, requestID interface{}, startTime time.Time) {
	// Set SSE headers for streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Track request for cancellation
	if requestID != nil {
		s.requestsMutex.Lock()
		s.activeRequests[requestID] = cancel
		s.requestsMutex.Unlock()

		defer func() {
			s.requestsMutex.Lock()
			delete(s.activeRequests, requestID)
			s.requestsMutex.Unlock()
		}()
	}

	logger.Info(fmt.Sprintf("Starting long operation with progress streaming for request: %v", requestID))

	// Parse request to get tool parameters
	var jsonRPCReq map[string]interface{}
	json.Unmarshal(bodyBytes, &jsonRPCReq)
	params, _ := jsonRPCReq["params"].(map[string]interface{})
	arguments, _ := params["arguments"].(map[string]interface{})

	seconds, _ := arguments["seconds"].(float64)
	if seconds == 0 {
		seconds = 10
	}
	if seconds > 60 {
		seconds = 60
	}

	task, _ := arguments["task"].(string)
	if task == "" {
		task = "long operation"
	}

	// Simulate long operation with progress updates
	totalSteps := int(seconds)
	for i := 0; i < totalSteps; i++ {
		select {
		case <-ctx.Done():
			// Cancelled - send cancellation response
			cancelledResponse := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      requestID,
				"result": map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": fmt.Sprintf("{\"status\":\"cancelled\",\"task\":\"%s\",\"elapsed_seconds\":%.1f,\"completed\":false,\"message\":\"Operation cancelled after %.1f of %.0f seconds\"}",
								task, float64(i), float64(i), seconds),
						},
					},
				},
			}
			responseBytes, _ := json.Marshal(cancelledResponse)
			fmt.Fprintf(w, "data: %s\n\n", string(responseBytes))
			flusher.Flush()

			logger.Warn(fmt.Sprintf("Long operation cancelled after %d/%d seconds", i, totalSteps))
			TrackSDKRequest("cursor_long_op_cancelled", r.UserAgent(), r.RemoteAddr, time.Since(startTime), false)
			return

		case <-time.After(1 * time.Second):
			// Send progress notification
			progress := float64(i+1) / float64(totalSteps) * 100
			progressMsg := fmt.Sprintf("Processing %s: %d/%d seconds (%.0f%%)", task, i+1, totalSteps, progress)

			progressNotification := map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "notifications/progress",
				"params": map[string]interface{}{
					"progressToken": requestID,
					"progress":      i + 1,
					"total":         totalSteps,
					"message":       progressMsg,
				},
			}

			progressBytes, _ := json.Marshal(progressNotification)
			fmt.Fprintf(w, "data: %s\n\n", string(progressBytes))
			flusher.Flush()

			logger.Debug(fmt.Sprintf("Progress: %s", progressMsg))
		}
	}

	// Send final result
	finalResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      requestID,
		"result": map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("{\"status\":\"success\",\"task\":\"%s\",\"elapsed_seconds\":%.1f,\"completed\":true,\"message\":\"Successfully completed %s in %.1f seconds\"}",
						task, seconds, task, seconds),
				},
			},
		},
	}

	responseBytes, _ := json.Marshal(finalResponse)
	fmt.Fprintf(w, "data: %s\n\n", string(responseBytes))
	flusher.Flush()

	logger.Info(fmt.Sprintf("Long operation completed: %s (%.0f seconds)", task, seconds))
	TrackSDKRequest("cursor_long_op_completed", r.UserAgent(), r.RemoteAddr, time.Since(startTime), false)
}

// handleSDKMethodForCursor handles prompts and resources for Cursor by calling them directly
func (s *Server) handleSDKMethodForCursor(ctx context.Context, bodyBytes []byte, method string, requestID interface{}) ([]byte, error) {
	// Parse the JSON-RPC request
	var jsonRPCReq map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonRPCReq); err != nil {
		return nil, err
	}

	id := jsonRPCReq["id"]
	params, _ := jsonRPCReq["params"].(map[string]interface{})

	switch method {
	case "prompts/list":
		// List all prompts
		promptsList := []map[string]interface{}{
			{
				"name":        "code_review",
				"description": "Generate a structured code review prompt with analysis criteria",
				"arguments": []map[string]interface{}{
					{"name": "code", "description": "The source code to review", "required": true},
					{"name": "language", "description": "Programming language (default: go)", "required": false},
					{"name": "focus", "description": "Specific focus area (optional)", "required": false},
				},
			},
			{
				"name":        "git_commit",
				"description": "Generate a structured git commit message following best practices",
				"arguments": []map[string]interface{}{
					{"name": "diff", "description": "The git diff to summarize", "required": true},
					{"name": "type", "description": "Commit type (feat, fix, docs, etc.)", "required": false},
					{"name": "scope", "description": "Commit scope (optional)", "required": false},
				},
			},
			{
				"name":        "debug_help",
				"description": "Generate debugging assistance for error messages and stack traces",
				"arguments": []map[string]interface{}{
					{"name": "error", "description": "Error message or stack trace", "required": true},
					{"name": "context", "description": "Additional context about the error", "required": false},
				},
			},
		}
		return s.formatSuccessResponse(id, map[string]interface{}{"prompts": promptsList}), nil

	case "prompts/get":
		// Get specific prompt
		name, _ := params["name"].(string)
		arguments, _ := params["arguments"].(map[string]interface{})

		var result *mcpsdk.GetPromptResult
		var err error

		switch name {
		case "code_review":
			args := prompts.CodeReviewPromptArgs{}
			if code, ok := arguments["code"].(string); ok {
				args.Code = code
			}
			if lang, ok := arguments["language"].(string); ok {
				args.Language = lang
			}
			if focus, ok := arguments["focus"].(string); ok {
				args.Focus = focus
			}
			result, err = prompts.GetCodeReviewPrompt(ctx, nil, args)

		case "git_commit":
			args := prompts.GitCommitPromptArgs{}
			if diff, ok := arguments["diff"].(string); ok {
				args.Diff = diff
			}
			if commitType, ok := arguments["type"].(string); ok {
				args.Type = commitType
			}
			if scope, ok := arguments["scope"].(string); ok {
				args.Scope = scope
			}
			result, err = prompts.GetGitCommitPrompt(ctx, nil, args)

		case "debug_help":
			args := prompts.DebugHelpPromptArgs{}
			if errorMsg, ok := arguments["error"].(string); ok {
				args.Error = errorMsg
			}
			if contextInfo, ok := arguments["context"].(string); ok {
				args.Context = contextInfo
			}
			result, err = prompts.GetDebugHelpPrompt(ctx, nil, args)

		default:
			return s.formatErrorResponse(id, -32602, "Invalid params", fmt.Sprintf("Unknown prompt: %s", name)), nil
		}

		if err != nil {
			return s.formatErrorResponse(id, -32603, "Internal error", err.Error()), nil
		}
		return s.formatSuccessResponse(id, result), nil

	case "resources/list":
		// List all resources with pagination support
		cursorStr, _ := params["cursor"].(string)

		// All available resources
		allResources := []map[string]interface{}{
			{
				"uri":         "project://info",
				"name":        "Project Information",
				"description": "Information about the gomcp server (version, features, capabilities)",
				"mimeType":    "application/json",
			},
			{
				"uri":         "project://status",
				"name":        "System Status",
				"description": "Current system status and runtime metrics",
				"mimeType":    "application/json",
			},
			{
				"uri":         "docs://quickstart",
				"name":        "Quick Start Guide",
				"description": "Quick start guide for gomcp server",
				"mimeType":    "text/markdown",
			},
			{
				"uri":         "docs://api-reference",
				"name":        "API Reference",
				"description": "Complete API endpoint documentation",
				"mimeType":    "text/markdown",
			},
			{
				"uri":         "config://template",
				"name":        "Config Template",
				"description": "Configuration file template with all options",
				"mimeType":    "text/plain",
			},
			{
				"uri":         "config://env-vars",
				"name":        "Environment Variables",
				"description": "Complete environment variables reference",
				"mimeType":    "text/markdown",
			},
		}

		// Paginate resources (10 per page)
		paginatedResult, err := pagination.PaginateSlice(
			allResources,
			pagination.Cursor(cursorStr),
			10,
		)

		if err != nil {
			return s.formatErrorResponse(id, -32602, "Invalid cursor", err.Error()), nil
		}

		// Build response with pagination
		response := map[string]interface{}{
			"resources": paginatedResult.Items,
		}

		// Add nextCursor if present
		if paginatedResult.NextCursor != nil {
			response["nextCursor"] = *paginatedResult.NextCursor
		}

		return s.formatSuccessResponse(id, response), nil

	case "resources/read":
		// Read specific resource
		uri, _ := params["uri"].(string)

		var result *mcpsdk.ReadResourceResult
		var err error

		switch uri {
		case "project://info":
			result, err = resources.GetProjectInfo(ctx, nil)
		case "project://status":
			result, err = resources.GetSystemStatus(ctx, nil)
		case "docs://quickstart":
			result, err = resources.GetQuickStartGuide(ctx, nil)
		case "docs://api-reference":
			result, err = resources.GetAPIReference(ctx, nil)
		case "config://template":
			result, err = resources.GetConfigTemplate(ctx, nil)
		case "config://env-vars":
			result, err = resources.GetEnvVarsReference(ctx, nil)
		default:
			return s.formatErrorResponse(id, -32602, "Invalid params", fmt.Sprintf("Unknown resource: %s", uri)), nil
		}

		if err != nil {
			return s.formatErrorResponse(id, -32603, "Internal error", err.Error()), nil
		}
		return s.formatSuccessResponse(id, result), nil

	case "roots/list":
		// List filesystem roots
		result, err := roots.ListRoots(ctx, nil)
		if err != nil {
			return s.formatErrorResponse(id, -32603, "Internal error", err.Error()), nil
		}
		return s.formatSuccessResponse(id, map[string]interface{}{
			"roots": result.Roots,
		}), nil
	}

	return s.formatErrorResponse(id, -32601, "Method not found", method), nil
}

// formatSuccessResponse creates a JSON-RPC success response
func (s *Server) formatSuccessResponse(id interface{}, result interface{}) []byte {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	data, _ := json.Marshal(response)
	return data
}

// formatErrorResponse creates a JSON-RPC error response
func (s *Server) formatErrorResponse(id interface{}, code int, message string, data interface{}) []byte {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
			"data":    data,
		},
	}
	responseData, _ := json.Marshal(response)
	return responseData
}

// handleInitializeForCursor handles the initialize request for Cursor with updated capabilities
func (s *Server) handleInitializeForCursor(bodyBytes []byte) ([]byte, error) {
	// Parse the JSON-RPC request
	var jsonRPCReq map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonRPCReq); err != nil {
		return nil, err
	}

	id := jsonRPCReq["id"]

	// Build initialize response with all capabilities
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]string{
			"name":    "template",
			"version": "0.1.0",
		},
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"completion": map[string]interface{}{}, // Structured outputs
			},
			"prompts":     map[string]interface{}{},
			"resources":   map[string]interface{}{},
			"roots":       map[string]interface{}{},                // Protocol 2025-06-18
			"sampling":    map[string]interface{}{},                // Protocol 2025-06-18
			"elicitation": map[string]interface{}{},                // Protocol 2025-06-18
			"logging":     map[string]interface{}{"level": "info"}, // Protocol 2025-06-18
			"pagination": map[string]interface{}{
				"support":     true,
				"maxPageSize": 100,
			}, // Protocol 2025-06-18
		},
	}

	return s.formatSuccessResponse(id, result), nil
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// OAuthProtectedResourceHandler returns protected resource metadata
func (s *Server) OAuthProtectedResourceHandler(w http.ResponseWriter, r *http.Request) {
	host := s.getHost()
	response := map[string]interface{}{
		"resource":                 host,
		"authorization_servers":    []string{host},
		"scopes_supported":         []string{"snowflake-mcp-server"},
		"registration_endpoint":    fmt.Sprintf("%s/auth/register", host),
		"bearer_methods_supported": []string{"header"},
		"revocation_endpoint":      fmt.Sprintf("%s/auth/revoke", host),
		"introspection_endpoint":   fmt.Sprintf("%s/auth/introspect", host),
		"introspection_endpoint_auth_methods_supported": []string{
			"client_secret_basic",
			"client_secret_post",
			"none",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// OAuthAuthorizationServerHandler returns authorization server metadata
func (s *Server) OAuthAuthorizationServerHandler(w http.ResponseWriter, r *http.Request) {
	host := s.getHost()
	response := map[string]interface{}{
		"issuer":                   host,
		"authorization_endpoint":   fmt.Sprintf("%s/auth/authorize", host),
		"token_endpoint":           fmt.Sprintf("%s/auth/token", host),
		"registration_endpoint":    fmt.Sprintf("%s/auth/register", host),
		"scopes_supported":         []string{"dataverse-console"},
		"response_types_supported": []string{"code"},
		"response_modes_supported": []string{"query"},
		"grant_types_supported":    []string{"authorization_code", "refresh_token", "client_credentials"},
		"token_endpoint_auth_methods_supported": []string{
			"client_secret_basic",
			"client_secret_post",
			"none",
		},
		"revocation_endpoint": fmt.Sprintf("%s/auth/revoke", host),
		"revocation_endpoint_auth_methods_supported": []string{
			"client_secret_basic",
			"client_secret_post",
			"none",
		},
		"introspection_endpoint": fmt.Sprintf("%s/auth/introspect", host),
		"introspection_endpoint_auth_methods_supported": []string{
			"client_secret_basic",
			"client_secret_post",
			"none",
		},
		"code_challenge_methods_supported": []string{"S256"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterClientHandler handles OAuth client registration
func (s *Server) RegisterClientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req oauth.RegisterClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	resp, err := s.oauthService.RegisterClient(r.Context(), req)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to register client: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// AuthorizeHandler handles OAuth authorization requests
func (s *Server) AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	clientID := query.Get("client_id")
	redirectURI := query.Get("redirect_uri")
	state := query.Get("state")
	codeChallenge := query.Get("code_challenge")
	codeChallengeMethod := query.Get("code_challenge_method")
	scope := query.Get("scope")

	if clientID == "" || redirectURI == "" || codeChallenge == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Validate client
	_, err := s.oauthService.ValidateClient(r.Context(), clientID, "")
	if err != nil {
		http.Error(w, "Invalid client", http.StatusBadRequest)
		return
	}

	// Create authorization code
	code, err := s.oauthService.CreateAuthorizationCode(
		r.Context(),
		clientID,
		redirectURI,
		scope,
		codeChallenge,
		codeChallengeMethod,
		state,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create authorization code: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect back to client with code
	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, state)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// TokenHandler handles OAuth token requests
func (s *Server) TokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		s.handleAuthorizationCodeGrant(w, r)
	case "refresh_token":
		s.handleRefreshTokenGrant(w, r)
	default:
		http.Error(w, "Unsupported grant type", http.StatusBadRequest)
	}
}

func (s *Server) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	clientID := r.FormValue("client_id")
	codeVerifier := r.FormValue("code_verifier")
	redirectURI := r.FormValue("redirect_uri")

	// Validate authorization code
	authCode, err := s.oauthService.ValidateAuthorizationCode(r.Context(), code)
	if err != nil || authCode == nil {
		http.Error(w, "Invalid authorization code", http.StatusBadRequest)
		return
	}

	// Verify PKCE
	if !oauth.VerifyCodeChallenge(codeVerifier, authCode.CodeChallenge) {
		http.Error(w, "Invalid code verifier", http.StatusBadRequest)
		return
	}

	// Verify client and redirect URI
	if authCode.ClientID != clientID || authCode.RedirectURI != redirectURI {
		http.Error(w, "Invalid client or redirect URI", http.StatusBadRequest)
		return
	}

	// Generate tokens
	accessToken, _ := oauth.GenerateToken()
	refreshToken, _ := oauth.GenerateToken()

	// Store tokens
	expiresIn := 3600 // 1 hour
	s.oauthService.StoreAccessToken(r.Context(), accessToken, clientID, authCode.Scope, expiresIn)
	s.oauthService.StoreRefreshToken(r.Context(), refreshToken, clientID, accessToken, authCode.Scope, 86400*30) // 30 days

	// Mark code as used
	s.oauthService.MarkCodeAsUsed(r.Context(), code)

	// Return tokens
	response := map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    expiresIn,
		"refresh_token": refreshToken,
		"scope":         authCode.Scope,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshTokenStr := r.FormValue("refresh_token")
	clientID := r.FormValue("client_id")

	// Validate refresh token
	refreshToken, err := s.oauthService.ValidateRefreshToken(r.Context(), refreshTokenStr)
	if err != nil || refreshToken == nil {
		http.Error(w, "Invalid refresh token", http.StatusBadRequest)
		return
	}

	if refreshToken.ClientID != clientID {
		http.Error(w, "Invalid client", http.StatusBadRequest)
		return
	}

	// Generate new access token
	newAccessToken, _ := oauth.GenerateToken()
	expiresIn := 3600

	s.oauthService.StoreAccessToken(r.Context(), newAccessToken, clientID, refreshToken.Scope, expiresIn)

	response := map[string]interface{}{
		"access_token": newAccessToken,
		"token_type":   "Bearer",
		"expires_in":   expiresIn,
		"scope":        refreshToken.Scope,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) getHost() string {
	endpoint := s.config.MCPHostEndpoint
	if endpoint == "" {
		endpoint = "http://localhost:8080"
	}

	parsedURL, err := url.Parse(endpoint)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		logger.Warn(fmt.Sprintf("Invalid MCP_HOST_ENDPOINT '%s', falling back to default", endpoint))
		return "http://localhost:8080"
	}

	return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
}

// InitializeStorage initializes the storage service (PostgreSQL or in-memory)
func InitializeStorage(cfg *config.Config) (*oauth.Service, error) {
	if !cfg.EnableAuth {
		return nil, nil
	}

	ctx := context.Background()

	// Check if PostgreSQL is configured
	hasPostgres := cfg.PostgresHost != "" &&
		cfg.PostgresDB != "" &&
		cfg.PostgresUser != ""

	var store storage.Store

	if hasPostgres {
		// Use PostgreSQL storage
		storageCfg := StorageConfig{
			Host:           cfg.PostgresHost,
			Port:           cfg.PostgresPort,
			Database:       cfg.PostgresDB,
			Username:       cfg.PostgresUser,
			Password:       cfg.PostgresPassword,
			PoolSize:       cfg.PostgresPoolSize,
			MaxConnections: cfg.PostgresMaxConnections,
		}

		pgService := NewStorageService(storageCfg)
		if err := pgService.Connect(ctx, storageCfg); err != nil {
			return nil, fmt.Errorf("failed to initialize PostgreSQL storage: %w", err)
		}
		store = pgService
		logger.Info("Using PostgreSQL storage")
	} else {
		// Use in-memory storage
		memStore := storage.NewMemoryStore()
		if err := memStore.Connect(ctx, StorageConfig{}); err != nil {
			return nil, fmt.Errorf("failed to initialize in-memory storage: %w", err)
		}
		store = memStore
		logger.Info("Using in-memory storage (PostgreSQL not configured)")
	}

	oauthService := oauth.NewService(store)
	return oauthService, nil
}
