package elicitation

import (
	"context"
	"fmt"

	"github.com/NP-compete/gomcp/internal/logger"
	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ResponseAction represents the user's response to an elicitation request
type ResponseAction string

const (
	ActionAccept  ResponseAction = "accept"  // User explicitly approved and submitted data
	ActionDecline ResponseAction = "decline" // User explicitly declined
	ActionCancel  ResponseAction = "cancel"  // User dismissed without choosing
)

// ElicitationResult contains the user's response
type ElicitationResult struct {
	Action  ResponseAction         `json:"action"`
	Content map[string]interface{} `json:"content,omitempty"`
}

// CreateElicitation requests structured data from the user via the client
// Returns the user's response with action (accept/decline/cancel) and optional content
func CreateElicitation(
	ctx context.Context,
	session *mcp.ServerSession,
	message string,
	requestedSchema map[string]interface{},
) (*ElicitationResult, error) {

	if session == nil {
		return nil, fmt.Errorf("session is nil - cannot send elicitation request")
	}

	logger.Info(fmt.Sprintf("Sending elicitation request to client: %s", message))

	// Send elicitation request to client
	// Note: The official SDK may not have direct elicitation support yet
	// This is a placeholder for the proper implementation

	// For now, return a simulated response
	// In production, this would call: session.CreateElicitation(ctx, &mcp.CreateElicitationRequest{...})

	logger.Warn("Elicitation not yet supported in SDK - returning simulated accept")

	return &ElicitationResult{
		Action:  ActionAccept,
		Content: map[string]interface{}{},
	}, nil
}

// RequestSimpleText requests a simple text value from the user
func RequestSimpleText(
	ctx context.Context,
	session *mcp.ServerSession,
	message string,
	propertyName string,
	description string,
) (string, error) {

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			propertyName: map[string]interface{}{
				"type":        "string",
				"description": description,
			},
		},
		"required": []string{propertyName},
	}

	result, err := CreateElicitation(ctx, session, message, schema)
	if err != nil {
		return "", err
	}

	if result.Action != ActionAccept {
		return "", fmt.Errorf("user %s the request", result.Action)
	}

	value, ok := result.Content[propertyName].(string)
	if !ok {
		return "", fmt.Errorf("expected string value for %s", propertyName)
	}

	return value, nil
}

// RequestStructuredData requests multiple fields from the user
func RequestStructuredData(
	ctx context.Context,
	session *mcp.ServerSession,
	message string,
	schema map[string]interface{},
) (map[string]interface{}, ResponseAction, error) {

	result, err := CreateElicitation(ctx, session, message, schema)
	if err != nil {
		return nil, "", err
	}

	return result.Content, result.Action, nil
}

// BuildSimpleSchema creates a simple elicitation schema
// Example:
//
//	schema := BuildSimpleSchema(map[string]SchemaProperty{
//	  "name": {Type: "string", Description: "Your name", Required: true},
//	  "age": {Type: "number", Description: "Your age", Min: 18},
//	})
func BuildSimpleSchema(properties map[string]SchemaProperty) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	required := []string{}
	props := make(map[string]interface{})

	for name, prop := range properties {
		propSchema := map[string]interface{}{
			"type": prop.Type,
		}

		if prop.Title != "" {
			propSchema["title"] = prop.Title
		}
		if prop.Description != "" {
			propSchema["description"] = prop.Description
		}
		if prop.Min != nil {
			propSchema["minimum"] = *prop.Min
		}
		if prop.Max != nil {
			propSchema["maximum"] = *prop.Max
		}
		if prop.MinLength != nil {
			propSchema["minLength"] = *prop.MinLength
		}
		if prop.MaxLength != nil {
			propSchema["maxLength"] = *prop.MaxLength
		}
		if prop.Format != "" {
			propSchema["format"] = prop.Format
		}
		if len(prop.Enum) > 0 {
			propSchema["enum"] = prop.Enum
		}
		if len(prop.EnumNames) > 0 {
			propSchema["enumNames"] = prop.EnumNames
		}
		if prop.Default != nil {
			propSchema["default"] = prop.Default
		}

		props[name] = propSchema

		if prop.Required {
			required = append(required, name)
		}
	}

	schema["properties"] = props
	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// SchemaProperty defines a property in an elicitation schema
type SchemaProperty struct {
	Type        string      // "string", "number", "integer", "boolean"
	Title       string      // Display name
	Description string      // Help text
	Required    bool        // Is this field required?
	Min         *float64    // Minimum value (for numbers)
	Max         *float64    // Maximum value (for numbers)
	MinLength   *int        // Minimum length (for strings)
	MaxLength   *int        // Maximum length (for strings)
	Format      string      // Format: "email", "uri", "date", "date-time"
	Enum        []string    // Allowed values (for enums)
	EnumNames   []string    // Display names for enum values
	Default     interface{} // Default value
}
