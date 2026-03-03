package api

import (
	"net/http"
	"time"

	"github.com/NP-compete/gomcp/internal/config"
	"github.com/NP-compete/gomcp/internal/logger"
	"github.com/NP-compete/gomcp/internal/mcp"
	"github.com/NP-compete/gomcp/internal/middleware"
	"github.com/NP-compete/gomcp/internal/oauth"
	"github.com/NP-compete/gomcp/internal/storage"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// StorageService is an alias for storage.Service (PostgreSQL)
type StorageService = storage.Service

// StorageConfig is an alias for storage.Config
type StorageConfig = storage.Config

// NewStorageService is an alias for storage.NewService
var NewStorageService = storage.NewService

// Store interface alias
type Store = storage.Store

// NewRouter creates and configures the Chi router
func NewRouter(cfg *config.Config, mcpServer *mcp.Server, oauthService *oauth.Service) http.Handler {
	r := chi.NewRouter()

	// Get logger instance
	log := logger.GetLogger()

	// Apply standard middleware (order matters!)
	r.Use(middleware.RequestID)                    // Add request ID first
	r.Use(chimiddleware.RealIP)                    // Get real IP
	r.Use(middleware.Recovery(*log))               // Recover from panics with logging
	r.Use(middleware.LoggingMiddleware)            // Log all requests
	r.Use(chimiddleware.Compress(5))               // Compress responses
	r.Use(chimiddleware.Timeout(60 * time.Second)) // Request timeout

	// Apply session middleware
	sessionSecret := cfg.GetSessionSecret()
	r.Use(middleware.SessionMiddleware(sessionSecret))

	// Apply CORS middleware if enabled
	corsMiddleware := middleware.CORSMiddleware(
		cfg.CORSEnabled,
		cfg.CORSOrigins,
		cfg.CORSCredentials,
		cfg.CORSMethods,
		cfg.CORSHeaders,
	)
	r.Use(corsMiddleware)

	// Create API server
	apiServer := NewServer(mcpServer, oauthService, cfg)

	// Public routes
	r.Get("/health", apiServer.HealthCheckHandler)
	r.Get("/metrics", MetricsHandler) // Metrics endpoint
	r.Get("/version", VersionHandler) // Version endpoint
	r.Get("/.well-known/oauth-protected-resource", apiServer.OAuthProtectedResourceHandler)
	r.Get("/.well-known/oauth-authorization-server", apiServer.OAuthAuthorizationServerHandler)

	// OAuth routes
	if cfg.EnableAuth {
		r.Post("/auth/register", apiServer.RegisterClientHandler)
		r.Get("/auth/authorize", apiServer.AuthorizeHandler)
		r.Post("/auth/token", apiServer.TokenHandler)
	}

	// MCP routes with optional auth middleware
	r.Group(func(r chi.Router) {
		if cfg.EnableAuth && oauthService != nil {
			r.Use(middleware.AuthMiddleware(oauthService, cfg.EnableAuth))
		}
		// Legacy custom protocol endpoint
		r.Post("/mcp", apiServer.MCPHandler)
		r.Post("/mcp/", apiServer.MCPHandler)

		// Official SDK SSE endpoint (for MCP 2024-11-05 spec compliance)
		// SSE requires both GET (for connection) and POST (for messages)
		sseHandler := apiServer.GetSSEHandler()
		r.Handle("/mcp/sse", sseHandler)
		r.Handle("/mcp/sse/", sseHandler)
	})

	// Development/debugging routes (can be disabled in production)
	if cfg.LogLevel == "DEBUG" {
		r.Mount("/debug", chimiddleware.Profiler())
	}

	return r
}
