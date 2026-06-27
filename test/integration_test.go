package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NP-compete/gomcp/internal/api"
	"github.com/NP-compete/gomcp/internal/config"
	"github.com/NP-compete/gomcp/internal/constants"
	"github.com/NP-compete/gomcp/internal/mcp"
)

// Test server setup
func setupTestServer(t *testing.T) (*httptest.Server, *api.Server) {
	cfg := &config.Config{
		MCPPort:              8081,
		MCPTransportProtocol: "http",
		CursorCompatibleSSE:  true,
		EnableAuth:           false,
	}

	mcpServer := mcp.NewServer()
	apiServer := api.NewServer(mcpServer, nil, cfg)

	handler := apiServer.GetSSEHandler()
	ts := httptest.NewServer(handler)

	return ts, apiServer
}

func sendJSONRPCRequest(t *testing.T, url string, request map[string]interface{}) map[string]interface{} {
	jsonData, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	return result
}

func TestInitialize(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      0,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test",
				"version": "1.0",
			},
		},
	}

	response := sendJSONRPCRequest(t, ts.URL, request)

	// Check response structure
	if response["jsonrpc"] != "2.0" {
		t.Error("Expected jsonrpc version 2.0")
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result in response")
	}

	// Check capabilities
	capabilities, ok := result["capabilities"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected capabilities in result")
	}

	expectedCapabilities := []string{
		"tools", "prompts", "resources", "roots",
		"sampling", "elicitation", "logging", "pagination",
	}

	for _, cap := range expectedCapabilities {
		if _, exists := capabilities[cap]; !exists {
			t.Errorf("Expected capability '%s' not found", cap)
		}
	}

	// Check tools.completion
	tools, ok := capabilities["tools"].(map[string]interface{})
	if !ok {
		t.Error("Expected tools capability to be an object")
	}

	if _, exists := tools["completion"]; !exists {
		t.Error("Expected tools.completion capability")
	}

	// Check pagination
	pag, ok := capabilities["pagination"].(map[string]interface{})
	if !ok {
		t.Error("Expected pagination capability to be an object")
	}

	if pag["support"] != true {
		t.Error("Expected pagination.support to be true")
	}

	if pag["maxPageSize"] != float64(constants.MaxPageSize) {
		t.Errorf("Expected pagination.maxPageSize to be %d, got %v", constants.MaxPageSize, pag["maxPageSize"])
	}

	// Check logging
	logging, ok := capabilities["logging"].(map[string]interface{})
	if !ok {
		t.Error("Expected logging capability to be an object")
	}

	if logging["level"] != "info" {
		t.Errorf("Expected logging.level to be 'info', got %v", logging["level"])
	}
}

func TestToolsList(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	response := sendJSONRPCRequest(t, ts.URL, request)

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result in response")
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatal("Expected tools array in result")
	}

	if len(tools) == 0 {
		t.Error("Expected at least one tool")
	}

	// Check first tool structure
	tool, ok := tools[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected tool to be an object")
	}

	if tool["name"] == nil {
		t.Error("Expected tool to have name")
	}

	if tool["description"] == nil {
		t.Error("Expected tool to have description")
	}

	if tool["inputSchema"] == nil {
		t.Error("Expected tool to have inputSchema")
	}
}

func TestToolsCall(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "multiply_numbers",
			"arguments": map[string]interface{}{
				"a": 6,
				"b": 7,
			},
		},
	}

	response := sendJSONRPCRequest(t, ts.URL, request)

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result in response")
	}

	content, ok := result["content"].([]interface{})
	if !ok {
		t.Fatal("Expected content array in result")
	}

	if len(content) == 0 {
		t.Error("Expected at least one content item")
	}
}

func TestPromptsList(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "prompts/list",
		"params":  map[string]interface{}{},
	}

	response := sendJSONRPCRequest(t, ts.URL, request)

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result in response")
	}

	prompts, ok := result["prompts"].([]interface{})
	if !ok {
		t.Fatal("Expected prompts array in result")
	}

	if len(prompts) < 3 {
		t.Errorf("Expected at least 3 prompts, got %d", len(prompts))
	}

	// Check first prompt structure
	prompt, ok := prompts[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected prompt to be an object")
	}

	if prompt["name"] == nil {
		t.Error("Expected prompt to have name")
	}

	if prompt["description"] == nil {
		t.Error("Expected prompt to have description")
	}
}

func TestResourcesList(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "resources/list",
		"params":  map[string]interface{}{},
	}

	response := sendJSONRPCRequest(t, ts.URL, request)

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result in response")
	}

	resources, ok := result["resources"].([]interface{})
	if !ok {
		t.Fatal("Expected resources array in result")
	}

	if len(resources) == 0 {
		t.Error("Expected at least one resource")
	}

	// Check resource structure
	resource, ok := resources[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected resource to be an object")
	}

	if resource["uri"] == nil {
		t.Error("Expected resource to have uri")
	}

	if resource["name"] == nil {
		t.Error("Expected resource to have name")
	}
}

func TestResourcesListWithPagination(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	// First page
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      5,
		"method":  "resources/list",
		"params":  map[string]interface{}{},
	}

	response := sendJSONRPCRequest(t, ts.URL, request)

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result in response")
	}

	resources, ok := result["resources"].([]interface{})
	if !ok {
		t.Fatal("Expected resources array in result")
	}

	if len(resources) == 0 {
		t.Error("Expected at least one resource")
	}

	// Check if nextCursor is present (if there are more than 10 resources)
	// In our case, we have 6 resources, so no nextCursor expected
	if nextCursor, exists := result["nextCursor"]; exists && nextCursor != nil {
		t.Log("nextCursor present (more pages available)")
	}
}

func TestResourcesRead(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      6,
		"method":  "resources/read",
		"params": map[string]interface{}{
			"uri": "project://info",
		},
	}

	response := sendJSONRPCRequest(t, ts.URL, request)

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result in response")
	}

	contents, ok := result["contents"].([]interface{})
	if !ok {
		t.Fatal("Expected contents array in result")
	}

	if len(contents) == 0 {
		t.Error("Expected at least one content item")
	}

	// Check content structure
	content, ok := contents[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected content to be an object")
	}

	if content["uri"] == nil {
		t.Error("Expected content to have uri")
	}
}

func TestRootsList(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      7,
		"method":  "roots/list",
		"params":  map[string]interface{}{},
	}

	response := sendJSONRPCRequest(t, ts.URL, request)

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result in response")
	}

	roots, ok := result["roots"].([]interface{})
	if !ok {
		t.Fatal("Expected roots array in result")
	}

	if len(roots) < 3 {
		t.Errorf("Expected at least 3 roots, got %d", len(roots))
	}

	// Check root structure
	root, ok := roots[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected root to be an object")
	}

	if root["uri"] == nil {
		t.Error("Expected root to have uri")
	}

	if root["name"] == nil {
		t.Error("Expected root to have name")
	}

	// Check URI format
	uri, ok := root["uri"].(string)
	if !ok || len(uri) < 7 || uri[:7] != "file://" {
		t.Errorf("Expected root URI to start with 'file://', got '%v'", root["uri"])
	}
}

func TestPing(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      8,
		"method":  "ping",
		"params":  map[string]interface{}{},
	}

	response := sendJSONRPCRequest(t, ts.URL, request)

	if response["jsonrpc"] != "2.0" {
		t.Error("Expected jsonrpc version 2.0")
	}

	// Ping should return empty result
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result in response")
	}

	// Result should be empty or minimal
	_ = result
}

func TestInvalidMethod(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      9,
		"method":  "invalid/method",
		"params":  map[string]interface{}{},
	}

	response := sendJSONRPCRequest(t, ts.URL, request)

	// Should have error
	if _, ok := response["error"]; !ok {
		t.Error("Expected error for invalid method")
	}
}

func TestInvalidJSON(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Post(ts.URL, "application/json", bytes.NewBufferString("{invalid json}"))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var response map[string]interface{}
	json.Unmarshal(body, &response)

	// Should have parse error
	if _, ok := response["error"]; !ok {
		t.Error("Expected error for invalid JSON")
	}
}
