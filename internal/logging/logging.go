package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/NP-compete/gomcp/internal/logger"
	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// LogLevel represents the severity of a log message
type LogLevel string

const (
	LevelDebug     LogLevel = "debug"
	LevelInfo      LogLevel = "info"
	LevelNotice    LogLevel = "notice"
	LevelWarning   LogLevel = "warning"
	LevelError     LogLevel = "error"
	LevelCritical  LogLevel = "critical"
	LevelAlert     LogLevel = "alert"
	LevelEmergency LogLevel = "emergency"
)

// LogMessage represents a log message to be sent to the client
type LogMessage struct {
	Level     LogLevel               `json:"level"`
	Logger    string                 `json:"logger,omitempty"`
	Data      interface{}            `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

// LogToClient sends a log message to the client via MCP notifications/message
// This allows the server to communicate operational information to the client
func LogToClient(ctx context.Context, session *mcp.ServerSession, level LogLevel, message string, data interface{}) error {
	if session == nil {
		// If no session, fall back to server logs
		logger.Info(fmt.Sprintf("[CLIENT LOG %s] %s", level, message))
		return nil
	}

	// Note: The official SDK may not have direct logging notification support yet
	// This is a placeholder for the proper implementation
	// In production, this would call: session.SendNotification(ctx, "notifications/message", logMsg)
	// with a LogMessage struct containing level, logger, data, and timestamp

	// For now, log to server logs as fallback
	logger.Info(fmt.Sprintf("[CLIENT LOG %s] %s: %v", level, message, data))

	return nil
}

// Debug sends a debug-level log message to the client
func Debug(ctx context.Context, session *mcp.ServerSession, message string, data interface{}) error {
	return LogToClient(ctx, session, LevelDebug, message, data)
}

// Info sends an info-level log message to the client
func Info(ctx context.Context, session *mcp.ServerSession, message string, data interface{}) error {
	return LogToClient(ctx, session, LevelInfo, message, data)
}

// Notice sends a notice-level log message to the client
func Notice(ctx context.Context, session *mcp.ServerSession, message string, data interface{}) error {
	return LogToClient(ctx, session, LevelNotice, message, data)
}

// Warning sends a warning-level log message to the client
func Warning(ctx context.Context, session *mcp.ServerSession, message string, data interface{}) error {
	return LogToClient(ctx, session, LevelWarning, message, data)
}

// Error sends an error-level log message to the client
func Error(ctx context.Context, session *mcp.ServerSession, message string, data interface{}) error {
	return LogToClient(ctx, session, LevelError, message, data)
}

// Critical sends a critical-level log message to the client
func Critical(ctx context.Context, session *mcp.ServerSession, message string, data interface{}) error {
	return LogToClient(ctx, session, LevelCritical, message, data)
}

// Alert sends an alert-level log message to the client
func Alert(ctx context.Context, session *mcp.ServerSession, message string, data interface{}) error {
	return LogToClient(ctx, session, LevelAlert, message, data)
}

// Emergency sends an emergency-level log message to the client
func Emergency(ctx context.Context, session *mcp.ServerSession, message string, data interface{}) error {
	return LogToClient(ctx, session, LevelEmergency, message, data)
}

// RateLimiter provides simple rate limiting for log messages
type RateLimiter struct {
	maxPerSecond int
	lastReset    time.Time
	count        int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxPerSecond int) *RateLimiter {
	return &RateLimiter{
		maxPerSecond: maxPerSecond,
		lastReset:    time.Now(),
		count:        0,
	}
}

// Allow checks if a message should be allowed based on rate limit
func (rl *RateLimiter) Allow() bool {
	now := time.Now()

	// Reset counter every second
	if now.Sub(rl.lastReset) >= time.Second {
		rl.count = 0
		rl.lastReset = now
	}

	// Check if under limit
	if rl.count < rl.maxPerSecond {
		rl.count++
		return true
	}

	return false
}

// RateLimitedLogger wraps logging with rate limiting
type RateLimitedLogger struct {
	limiter *RateLimiter
	dropped int
}

// NewRateLimitedLogger creates a rate-limited logger
func NewRateLimitedLogger(maxPerSecond int) *RateLimitedLogger {
	return &RateLimitedLogger{
		limiter: NewRateLimiter(maxPerSecond),
		dropped: 0,
	}
}

// Log sends a log message if within rate limit
func (rll *RateLimitedLogger) Log(ctx context.Context, session *mcp.ServerSession, level LogLevel, message string, data interface{}) error {
	if !rll.limiter.Allow() {
		rll.dropped++
		// Every 100 dropped messages, send a warning
		if rll.dropped%100 == 0 {
			Warning(ctx, session, fmt.Sprintf("Rate limit reached: %d messages dropped", rll.dropped), nil)
		}
		return fmt.Errorf("rate limit exceeded")
	}

	return LogToClient(ctx, session, level, message, data)
}

// GetDroppedCount returns the number of dropped messages
func (rll *RateLimitedLogger) GetDroppedCount() int {
	return rll.dropped
}

// ResetDroppedCount resets the dropped message counter
func (rll *RateLimitedLogger) ResetDroppedCount() {
	rll.dropped = 0
}

// SecureLog ensures sensitive data is not logged
// It redacts common sensitive fields
func SecureLog(ctx context.Context, session *mcp.ServerSession, level LogLevel, message string, data map[string]interface{}) error {
	// List of sensitive field names to redact
	sensitiveFields := []string{
		"password", "token", "secret", "api_key", "apikey", "auth",
		"authorization", "credential", "private_key", "privatekey",
		"session_id", "sessionid", "cookie", "ssn", "credit_card",
	}

	// Create a copy and redact sensitive fields
	if data != nil {
		secureCopy := make(map[string]interface{})
		for k, v := range data {
			shouldRedact := false
			for _, sensitive := range sensitiveFields {
				if k == sensitive {
					shouldRedact = true
					break
				}
			}

			if shouldRedact {
				secureCopy[k] = "[REDACTED]"
			} else {
				secureCopy[k] = v
			}
		}
		data = secureCopy
	}

	return LogToClient(ctx, session, level, message, data)
}
