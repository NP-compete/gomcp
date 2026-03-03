package tools

import (
	"context"
	"fmt"

	"github.com/NP-compete/gomcp/internal/logger"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CodeReviewInput defines the input schema for generate_code_review_prompt tool
type CodeReviewInput struct {
	Code     string `json:"code" jsonschema:"required,the source code to be reviewed"`
	Language string `json:"language" jsonschema:"programming language of the code (default: go)"`
}

// CodeReviewOutput defines the output schema for generate_code_review_prompt tool
type CodeReviewOutput struct {
	Status    string `json:"status" jsonschema:"operation status (success/error)"`
	Operation string `json:"operation" jsonschema:"type of operation performed"`
	Language  string `json:"language" jsonschema:"programming language of the code"`
	Prompt    string `json:"prompt" jsonschema:"the generated code review prompt"`
	Message   string `json:"message" jsonschema:"human-readable status message"`
}

// GenerateCodeReviewPromptSDK generates a structured code review prompt using the official SDK pattern
//
// TOOL_NAME=generate_code_review_prompt
// DISPLAY_NAME=Code Review Prompt Generator
// USECASE=Analyze code for quality, bugs, and improvements using external AI service
// INSTRUCTIONS=1. Provide source code as string, 2. Specify programming language, 3. Receive formatted review prompt
// INPUT_DESCRIPTION=code (string): source code to review, language (string, optional): programming language (default: "go")
// OUTPUT_DESCRIPTION=Dictionary with status, operation, language, formatted prompt text, and message
// EXAMPLES=generate_code_review_prompt("func hello() { fmt.Println(\"world\") }", "go")
// PREREQUISITES=Have source code ready for analysis
// RELATED_TOOLS=None - generates prompts for external AI analysis
func GenerateCodeReviewPromptSDK(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input CodeReviewInput,
) (*mcp.CallToolResult, CodeReviewOutput, error) {
	// Log the tool call
	if req != nil && req.Params != nil {
		logger.WithFields(map[string]interface{}{
			"tool_name": req.Params.Name,
			"arguments": map[string]interface{}{"code": input.Code, "language": input.Language},
		}).Info().Msg("Tool call invoked")
	}

	// Validate inputs
	if input.Code == "" {
		return nil, CodeReviewOutput{
			Status:  "error",
			Message: "Code must be a non-empty string",
		}, fmt.Errorf("code must be a non-empty string")
	}

	// Default language to go if empty
	language := input.Language
	if language == "" {
		language = "go"
	}

	logger.Info(fmt.Sprintf("Generating code review prompt for %s code", language))

	promptContent := fmt.Sprintf(`Please review the following %s code:

`+"```"+`%s
%s
`+"```"+`

Focus on:
- Code quality and readability
- Potential bugs or issues
- Best practices
- Performance considerations
`, language, language, input.Code)

	output := CodeReviewOutput{
		Status:    "success",
		Operation: "code_review_prompt",
		Language:  language,
		Prompt:    promptContent,
		Message:   fmt.Sprintf("Successfully generated code review prompt for %s", language),
	}

	return nil, output, nil
}
