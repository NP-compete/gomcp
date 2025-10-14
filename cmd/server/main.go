package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NP-compete/gomcp/internal/api"
	"github.com/NP-compete/gomcp/internal/config"
	"github.com/NP-compete/gomcp/internal/logger"
	"github.com/NP-compete/gomcp/internal/mcp"
	"github.com/NP-compete/gomcp/internal/oauth"
	"github.com/NP-compete/gomcp/internal/version"
	mcpSdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Command line flags
	versionFlag := flag.Bool("version", false, "Print version information and exit")
	flag.Parse()

	// Handle version flag
	if *versionFlag {
		fmt.Println(version.String())
		fmt.Println("\nDetailed version information:")
		info := version.Get()
		fmt.Printf("  Version:    %s\n", info.Version)
		fmt.Printf("  Commit:     %s\n", info.GitCommit)
		fmt.Printf("  Build Time: %s\n", info.BuildTime)
		fmt.Printf("  Go Version: %s\n", info.GoVersion)
		fmt.Printf("  OS/Arch:    %s/%s\n", info.OS, info.Arch)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.LogLevel)
	logger.Info(fmt.Sprintf("Template MCP Server starting... (%s)", version.String()))

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Fatal(fmt.Sprintf("Configuration validation failed: %v", err))
	}

	logger.Info(fmt.Sprintf("Server configured to use %s protocol", cfg.MCPTransportProtocol))

	// Check transport protocol and start accordingly
	if cfg.MCPTransportProtocol == "stdio" {
		// stdio transport - for Claude Desktop and other local clients
		runStdioTransport()
	} else {
		// HTTP/SSE transport - for web clients and HTTP-based communication
		runHTTPTransport(cfg)
	}
}

// runStdioTransport starts the server with stdio transport
func runStdioTransport() {
	logger.Info("Starting MCP server with stdio transport")

	// Initialize SDK server for stdio
	sdkServer := mcp.NewServerSDK()

	// Create context
	ctx := context.Background()

	logger.Info("Server is ready to accept stdio connections")

	// Run server with stdio transport (blocks until completion or error)
	if err := sdkServer.GetServer().Run(ctx, &mcpSdk.StdioTransport{}); err != nil {
		logger.Fatal(fmt.Sprintf("stdio server failed: %v", err))
	}

	logger.Info("stdio server shutting down")
}

// runHTTPTransport starts the server with HTTP/SSE transport
func runHTTPTransport(cfg *config.Config) {
	// Initialize MCP server
	mcpServer := mcp.NewServer()

	// Initialize storage and OAuth service if auth is enabled
	var oauthService *oauth.Service
	var err error
	if cfg.EnableAuth {
		oauthService, err = api.InitializeStorage(cfg)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Failed to initialize storage: %v", err))
		}
		logger.Info("Storage service initialized successfully")
	} else {
		logger.Info("Authentication disabled, skipping storage initialization")
	}

	// Create HTTP router
	router := api.NewRouter(cfg, mcpServer, oauthService)

	// Create HTTP server
	addr := cfg.GetServerAddress()
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info(fmt.Sprintf("Starting server on %s", addr))

		if cfg.HasSSL() {
			logger.Info(fmt.Sprintf("Starting server with SSL (keyfile: %s, certfile: %s)",
				cfg.MCPSSLKeyfile, cfg.MCPSSLCertfile))
			if err := srv.ListenAndServeTLS(cfg.MCPSSLCertfile, cfg.MCPSSLKeyfile); err != nil && err != http.ErrServerClosed {
				logger.Fatal(fmt.Sprintf("Server failed to start: %v", err))
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal(fmt.Sprintf("Server failed to start: %v", err))
			}
		}
	}()

	logger.Info("Server is ready to accept connections")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Received shutdown signal, shutting down gracefully...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("Server forced to shutdown: %v", err))
	}

	// Cleanup storage if initialized
	if oauthService != nil {
		logger.Info("Cleaning up storage service...")
		// Storage cleanup would go here
	}

	logger.Info("Template MCP server shutting down")
}
