package tools

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/NP-compete/gomcp/internal/logger"
)

// GetRedHatLogo returns the Red Hat logo as a base64 encoded string
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
func GetRedHatLogo(args map[string]interface{}) (interface{}, error) {
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
		return map[string]interface{}{
			"status":    "error",
			"operation": "get_redhat_logo",
			"error":     "file_not_found",
			"message":   errorMsg,
		}, nil
	}

	logoPath, _ = filepath.Abs(logoPath)
	logger.Info(fmt.Sprintf("Reading Red Hat logo from: %s", logoPath))

	// Encode to base64
	logoBase64 := base64.StdEncoding.EncodeToString(logoData)

	logger.Info("Successfully read and encoded Red Hat logo")

	return map[string]interface{}{
		"status":      "success",
		"operation":   "get_redhat_logo",
		"name":        "Red Hat Logo",
		"description": "Red Hat logo as base64 encoded PNG",
		"mimeType":    "image/png",
		"data":        logoBase64,
		"size_bytes":  len(logoData),
		"message":     "Successfully retrieved Red Hat logo",
	}, nil
}
