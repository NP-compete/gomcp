package constants

import "time"

// HTTP Server Timeouts
const (
	// DefaultRequestTimeout is the timeout for HTTP requests
	DefaultRequestTimeout = 60 * time.Second
	// DefaultReadTimeout is the timeout for reading requests
	DefaultReadTimeout = 30 * time.Second
	// DefaultWriteTimeout is the timeout for writing responses
	DefaultWriteTimeout = 30 * time.Second
	// DefaultIdleTimeout is the timeout for idle connections
	DefaultIdleTimeout = 120 * time.Second
	// DefaultShutdownTimeout is the timeout for server shutdown
	DefaultShutdownTimeout = 30 * time.Second
)

// Pagination
const (
	// DefaultPageSize is the default number of items per page
	DefaultPageSize = 10
	// MaxPageSize is the maximum number of items per page
	MaxPageSize = 100
)

// Long Operation Tool
const (
	// MinOperationSeconds is the minimum seconds for long operation
	MinOperationSeconds = 1
	// MaxOperationSeconds is the maximum seconds for long operation
	MaxOperationSeconds = 60
)

// Rate Limiting (for future use)
const (
	// DefaultRateLimit is the default requests per second
	DefaultRateLimit = 100
	// DefaultBurstSize is the default burst size for rate limiting
	DefaultBurstSize = 20
)

// Compression
const (
	// DefaultCompressionLevel is the default gzip compression level
	DefaultCompressionLevel = 5
)
