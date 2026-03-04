package constants

import "time"

const (
	// HTTP Server Timeouts

	// DefaultRequestTimeout is the maximum duration for processing a request
	DefaultRequestTimeout = 60 * time.Second

	// DefaultReadTimeout is the maximum duration for reading the request body
	DefaultReadTimeout = 30 * time.Second

	// DefaultWriteTimeout is the maximum duration for writing the response
	DefaultWriteTimeout = 30 * time.Second

	// DefaultIdleTimeout is the maximum duration for idle connections
	DefaultIdleTimeout = 120 * time.Second

	// DefaultShutdownTimeout is the maximum duration for graceful shutdown
	DefaultShutdownTimeout = 30 * time.Second

	// Pagination

	// DefaultPageSize is the default number of items per page
	DefaultPageSize = 10

	// MaxPageSize is the maximum number of items per page
	MaxPageSize = 100

	// Long Operation Tool

	// MinOperationSeconds is the minimum duration for long operation test tool
	MinOperationSeconds = 1

	// MaxOperationSeconds is the maximum duration for long operation test tool
	MaxOperationSeconds = 60

	// Rate Limiting (for future use)

	// DefaultRateLimit is the default rate limit per minute
	DefaultRateLimit = 100

	// DefaultBurstSize is the default burst size for rate limiting
	DefaultBurstSize = 20
)
