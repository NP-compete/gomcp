package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/NP-compete/gomcp/internal/logger"
)

// LongOperationInput defines the input schema for long_operation tool
type LongOperationInput struct {
	Seconds int    `json:"seconds" jsonschema:"required,number of seconds to run (1-60)"`
	Task    string `json:"task" jsonschema:"task description (optional)"`
}

// LongOperationOutput defines the output schema for long_operation tool
type LongOperationOutput struct {
	Status      string  `json:"status" jsonschema:"operation status (success/cancelled/error)"`
	Task        string  `json:"task" jsonschema:"task description"`
	ElapsedSecs float64 `json:"elapsed_seconds" jsonschema:"actual elapsed time"`
	Completed   bool    `json:"completed" jsonschema:"whether task completed or was cancelled"`
	Message     string  `json:"message" jsonschema:"human-readable status message"`
}

// LongOperationSDK demonstrates proper cancellation and progress handling for long-running operations
//
// TOOL_NAME=long_operation
// DISPLAY_NAME=Long Running Operation (Cancellable with Progress)
// USECASE=Demonstrates cancellation and progress reporting for long-running operations
// INSTRUCTIONS=1. Specify duration in seconds, 2. Call function with progressToken, 3. Receive progress notifications, 4. Can cancel mid-operation
// INPUT_DESCRIPTION=seconds (1-60): operation duration, task: optional description
// OUTPUT_DESCRIPTION=Status with elapsed time and completion state, sends progress notifications during execution
// EXAMPLES=long_operation(10, "processing data"), long_operation(30, "heavy computation")
// PREREQUISITES=None - demo tool for cancellation and progress
// RELATED_TOOLS=None - educational example
func LongOperationSDK(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input LongOperationInput,
) (*mcp.CallToolResult, LongOperationOutput, error) {
	// Log the tool call
	if req != nil && req.Params != nil {
		logger.WithFields(map[string]interface{}{
			"tool_name": req.Params.Name,
			"arguments": map[string]interface{}{"seconds": input.Seconds, "task": input.Task},
		}).Info().Msg("Tool call invoked")
	}

	// Validate input
	if input.Seconds < 1 {
		input.Seconds = 1
	}
	if input.Seconds > 60 {
		input.Seconds = 60
	}
	if input.Task == "" {
		input.Task = "long operation"
	}

	logger.Info(fmt.Sprintf("Starting long operation: %s (%d seconds)", input.Task, input.Seconds))

	// Get progress token if provided
	var progressToken interface{}
	if req != nil && req.Params != nil {
		progressToken = req.Params.GetProgressToken()
	}

	startTime := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	completed := 0
	total := input.Seconds

	// Simulate long-running operation with cancellation checks and progress reporting
	for completed < total {
		select {
		case <-ctx.Done():
			// Context cancelled - client requested cancellation
			elapsed := time.Since(startTime).Seconds()
			logger.Warn(fmt.Sprintf("Long operation cancelled after %.2f seconds: %s", elapsed, input.Task))

			return nil, LongOperationOutput{
				Status:      "cancelled",
				Task:        input.Task,
				ElapsedSecs: elapsed,
				Completed:   false,
				Message:     fmt.Sprintf("Operation cancelled after %.2f of %d seconds", elapsed, total),
			}, nil

		case <-ticker.C:
			// Progress tick
			completed++
			logger.Debug(fmt.Sprintf("Long operation progress: %d/%d seconds - %s", completed, total, input.Task))

			// Send progress notification if progress token provided
			if progressToken != nil && req != nil && req.Session != nil {
				progressMsg := fmt.Sprintf("Processing %s: %d/%d seconds", input.Task, completed, total)
				err := req.Session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
					ProgressToken: progressToken,
					Progress:      float64(completed),
					Total:         float64(total),
					Message:       progressMsg,
				})
				if err != nil {
					logger.Warn(fmt.Sprintf("Failed to send progress notification: %v", err))
				}
			}
		}
	}

	// Operation completed successfully
	elapsed := time.Since(startTime).Seconds()
	logger.Info(fmt.Sprintf("Long operation completed in %.2f seconds: %s", elapsed, input.Task))

	return nil, LongOperationOutput{
		Status:      "success",
		Task:        input.Task,
		ElapsedSecs: elapsed,
		Completed:   true,
		Message:     fmt.Sprintf("Successfully completed %s in %.2f seconds", input.Task, elapsed),
	}, nil
}
