package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/NP-compete/gomcp/internal/completion"
	"github.com/NP-compete/gomcp/internal/logger"
)

// MultiplyInput defines the input schema for multiply_numbers tool
type MultiplyInput struct {
	A float64 `json:"a" jsonschema:"required,the first number to multiply"`
	B float64 `json:"b" jsonschema:"required,the second number to multiply"`
}

// MultiplyOutput defines the output schema for multiply_numbers tool
type MultiplyOutput struct {
	Status    string  `json:"status" jsonschema:"operation status (success/error)"`
	Operation string  `json:"operation" jsonschema:"type of operation performed"`
	A         float64 `json:"a" jsonschema:"first input number"`
	B         float64 `json:"b" jsonschema:"second input number"`
	Result    float64 `json:"result" jsonschema:"the multiplication result"`
	Message   string  `json:"message" jsonschema:"human-readable status message"`
}

// MultiplyNumbersSDK multiplies two numbers using the official SDK pattern
//
// TOOL_NAME=multiply_numbers
// DISPLAY_NAME=Number Multiplication
// USECASE=Multiply two (floating point) numbers together
// INSTRUCTIONS=1. Provide two numeric values (int or float), 2. Call function, 3. Receive result
// INPUT_DESCRIPTION=Two parameters: a (number), b (number). Examples: (4, 5), (3.14, 2.0), (-1, 10)
// OUTPUT_DESCRIPTION=Dictionary with status, operation, input values (a, b), result, and message
// EXAMPLES=multiply_numbers(4, 5), multiply_numbers(3.14, 2.0)
// PREREQUISITES=None - standalone arithmetic operation
// RELATED_TOOLS=None - basic math operation
func MultiplyNumbersSDK(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input MultiplyInput,
) (*mcp.CallToolResult, MultiplyOutput, error) {
	// Log the tool call
	if req != nil && req.Params != nil {
		logger.WithFields(map[string]interface{}{
			"tool_name": req.Params.Name,
			"arguments": map[string]interface{}{"a": input.A, "b": input.B},
		}).Info().Msg("Tool call invoked")
	}

	result := input.A * input.B

	logger.Info(fmt.Sprintf("Multiply tool called: %v * %v = %v", input.A, input.B, result))

	output := MultiplyOutput{
		Status:    "success",
		Operation: "multiplication",
		A:         input.A,
		B:         input.B,
		Result:    result,
		Message:   fmt.Sprintf("Successfully multiplied %v and %v", input.A, input.B),
	}

	// Create structured output for better client integration
	structuredOutput := completion.NewStructuredOutput(
		fmt.Sprintf("%v × %v = %v", input.A, input.B, result),
		map[string]interface{}{
			"status":    output.Status,
			"operation": output.Operation,
			"a":         output.A,
			"b":         output.B,
			"result":    output.Result,
			"message":   output.Message,
		},
	)

	// Convert completion.ContentItem to mcp.Content
	// Create MCP content slice
	var mcpContent []mcp.Content
	for _, item := range structuredOutput.Content {
		mcpContent = append(mcpContent, &mcp.TextContent{Text: item.Text})
	}

	return &mcp.CallToolResult{
		Content: mcpContent,
		Meta: map[string]interface{}{
			"structuredContent": structuredOutput.StructuredContent,
		},
	}, output, nil
}

// MultiplyNumbersOutputSchema returns the output schema for the multiply_numbers tool
func MultiplyNumbersOutputSchema() *completion.ToolOutputSchema {
	return completion.BuildOutputSchema(
		"Result of multiplying two numbers",
		map[string]completion.PropertySchema{
			"status": {
				Type:        "string",
				Description: "Operation status",
				Enum:        []string{"success", "error"},
			},
			"operation": {
				Type:        "string",
				Description: "Type of operation performed",
			},
			"a": {
				Type:        "number",
				Description: "First input number",
			},
			"b": {
				Type:        "number",
				Description: "Second input number",
			},
			"result": {
				Type:        "number",
				Description: "The multiplication result",
			},
			"message": {
				Type:        "string",
				Description: "Human-readable status message",
			},
		},
		[]string{"status", "operation", "result"},
	)
}
