package sampling

import (
	"context"
	"fmt"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/NP-compete/gomcp/internal/logger"
)

// CreateMessageSampling handles sampling requests from the server to the client
// This allows the server to ask the client's LLM to generate text
func CreateMessageSampling(
	ctx context.Context,
	session *mcp.ServerSession,
	messages []*mcp.SamplingMessage,
	modelPreferences *mcp.ModelPreferences,
	systemPrompt string,
	maxTokens int,
) (*mcp.CreateMessageResult, error) {

	if session == nil {
		return nil, fmt.Errorf("session is nil - cannot send sampling request")
	}

	// Build sampling request
	params := &mcp.CreateMessageRequestParams{
		Messages:  messages,
		MaxTokens: maxTokens,
	}

	if systemPrompt != "" {
		params.SystemPrompt = systemPrompt
	}

	if modelPreferences != nil {
		params.ModelPreferences = modelPreferences
	}

	logger.Info(fmt.Sprintf("Sending sampling request to client (maxTokens=%d)", maxTokens))

	// Send sampling request to client
	result, err := session.CreateMessage(ctx, &mcp.CreateMessageRequest{
		Params: *params,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Sampling request failed: %v", err))
		return nil, err
	}

	logger.Info(fmt.Sprintf("Sampling response received: %d messages", len(result.Content)))
	return result, nil
}

// ExampleSamplingUsage demonstrates how to use sampling in a tool
func ExampleSamplingUsage(ctx context.Context, session *mcp.ServerSession, userQuery string) (string, error) {
	// Build messages for the LLM
	messages := []*mcp.SamplingMessage{
		{
			Role: mcp.RoleUser,
			Content: mcp.TextContent{
				Type: "text",
				Text: userQuery,
			},
		},
	}

	// Optional: Specify model preferences
	modelPrefs := &mcp.ModelPreferences{
		Hints: []*mcp.ModelHint{
			{
				Name: "claude-3-5-sonnet-20241022",
			},
		},
		CostPriority:         0.5,
		SpeedPriority:        0.5,
		IntelligencePriority: 0.8,
	}

	// Request sampling
	result, err := CreateMessageSampling(
		ctx,
		session,
		messages,
		modelPrefs,
		"You are a helpful assistant.",
		1000, // max tokens
	)

	if err != nil {
		return "", err
	}

	// Extract text from response
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			return textContent.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in sampling response")
}
