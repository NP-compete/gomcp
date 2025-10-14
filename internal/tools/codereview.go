package tools

import (
	"fmt"

	"github.com/NP-compete/gomcp/internal/logger"
)

// GenerateCodeReviewPrompt generates a structured code review prompt with comprehensive metadata
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
func GenerateCodeReviewPrompt(args map[string]interface{}) (interface{}, error) {
	// Extract code argument
	codeVal, codeOk := args["code"]
	if !codeOk {
		return map[string]interface{}{
			"status":  "error",
			"error":   "missing required argument",
			"message": "'code' parameter is required",
		}, nil
	}

	code, ok := codeVal.(string)
	if !ok || code == "" {
		return map[string]interface{}{
			"status":  "error",
			"error":   "invalid argument type",
			"message": "Code must be a non-empty string",
		}, nil
	}

	// Extract language argument (optional, defaults to "go")
	language := "go"
	if langVal, langOk := args["language"]; langOk {
		if lang, ok := langVal.(string); ok && lang != "" {
			language = lang
		}
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
`, language, language, code)

	return map[string]interface{}{
		"status":    "success",
		"operation": "code_review_prompt",
		"language":  language,
		"prompt":    promptContent,
		"message":   fmt.Sprintf("Successfully generated code review prompt for %s", language),
	}, nil
}
