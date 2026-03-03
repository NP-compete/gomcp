package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/NP-compete/gomcp/internal/logger"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LogoInput defines the input schema for get_redhat_logo tool (no parameters)
type LogoInput struct {
}

// LogoOutput defines the output schema for get_redhat_logo tool
type LogoOutput struct {
	Status      string `json:"status" jsonschema:"operation status (success/error)"`
	Operation   string `json:"operation" jsonschema:"type of operation performed"`
	Name        string `json:"name" jsonschema:"display name for the logo"`
	Description string `json:"description" jsonschema:"description of the logo"`
	MimeType    string `json:"mimeType" jsonschema:"MIME type of the image"`
	Data        string `json:"data" jsonschema:"base64 encoded PNG data"`
	SizeBytes   int    `json:"size_bytes" jsonschema:"size of the logo file in bytes"`
	Message     string `json:"message" jsonschema:"human-readable status message"`
}

// GetRedHatLogoSDK returns the Red Hat logo as a base64 encoded string using the official SDK pattern
//
// TOOL_NAME=get_redhat_logo
// DISPLAY_NAME=Get Red Hat Logo
// USECASE=Retrieve Red Hat logo for presentations, documentation, or branding
// INSTRUCTIONS=1. Call function (no parameters needed), 2. Receive base64-encoded logo data
// INPUT_DESCRIPTION=No input parameters required
// OUTPUT_DESCRIPTION=Dictionary with status, operation, logo metadata (name, description, mimeType), base64 data, size info, and message
// EXAMPLES=get_redhat_logo()
// PREREQUISITES=None - logo file must exist in assets directory
// RELATED_TOOLS=None - standalone asset retrieval
func GetRedHatLogoSDK(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input LogoInput,
) (*mcp.CallToolResult, LogoOutput, error) {
	// Log the tool call
	if req != nil && req.Params != nil {
		logger.WithFields(map[string]interface{}{
			"tool_name": req.Params.Name,
			"arguments": map[string]interface{}{},
		}).Info().Msg("Tool call invoked")
	}

	// Determine the path to the logo file
	// Look for assets directory relative to the executable or working directory
	possiblePaths := []string{
		"assets/redhat.png",
		"../assets/redhat.png",
		"../../assets/redhat.png",
		"gomcp/assets/redhat.png",
	}

	var logoPath string
	var logoData []byte
	var err error

	for _, path := range possiblePaths {
		logoData, err = os.ReadFile(path)
		if err == nil {
			logoPath = path
			break
		}
	}

	if err != nil {
		errorMsg := fmt.Sprintf("Could not find logo file: %v", err)
		logger.Error(errorMsg)
		return nil, LogoOutput{
			Status:    "error",
			Operation: "get_redhat_logo",
			Message:   errorMsg,
		}, fmt.Errorf("%s", errorMsg)
	}

	logoPath, _ = filepath.Abs(logoPath)
	logger.Info(fmt.Sprintf("Reading Red Hat logo from: %s", logoPath))

	// Encode to base64
	logoBase64 := base64.StdEncoding.EncodeToString(logoData)

	logger.Info("Successfully read and encoded Red Hat logo")

	output := LogoOutput{
		Status:      "success",
		Operation:   "get_redhat_logo",
		Name:        "Red Hat Logo",
		Description: "Red Hat logo as base64 encoded PNG",
		MimeType:    "image/png",
		Data:        logoBase64,
		SizeBytes:   len(logoData),
		Message:     "Successfully retrieved Red Hat logo",
	}

	return nil, output, nil
}
