package sampling

import (
	"context"
	"fmt"

	"github.com/NP-compete/gomcp/internal/logger"
	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// CreateMessageSampling handles sampling requests from the server to the client
// This allows the server to ask the client's LLM to generate text
func CreateMessageSampling(
	ctx context.Context,
	session *mcp.ServerSession,
	messages []*mcp.SamplingMessage,
	modelPreferences *mcp.ModelPreferences,
	systemPrompt string,
	maxTokens int64,
) (*mcp.CreateMessageResult, error) {

	if session == nil {
		return nil, fmt.Errorf("session is nil - cannot send sampling request")
	}

	params := &mcp.CreateMessageParams{
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

	result, err := session.CreateMessage(ctx, params)

	if err != nil {
		logger.Error(fmt.Sprintf("Sampling request failed: %v", err))
		return nil, err
	}

	logger.Info("Sampling response received")
	return result, nil
}

// ExampleSamplingUsage demonstrates how to use sampling in a tool
func ExampleSamplingUsage(ctx context.Context, session *mcp.ServerSession, userQuery string) (string, error) {
	messages := []*mcp.SamplingMessage{
		{
			Role:    mcp.Role("user"),
			Content: &mcp.TextContent{Text: userQuery},
		},
	}

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

	result, err := CreateMessageSampling(
		ctx,
		session,
		messages,
		modelPrefs,
		"You are a helpful assistant.",
		1000,
	)

	if err != nil {
		return "", err
	}

	if result.Content != nil {
		if textContent, ok := result.Content.(*mcp.TextContent); ok {
			return textContent.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in sampling response")
}
