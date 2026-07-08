package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server configuration

	// MCPHost is the hostname or IP address the server binds to.
	MCPHost string `mapstructure:"MCP_HOST"`

	// MCPPort is the port number the server listens on (1024-65535).
	MCPPort int `mapstructure:"MCP_PORT"`

	// MCPSSLKeyfile is the path to the SSL private key file (optional).
	MCPSSLKeyfile string `mapstructure:"MCP_SSL_KEYFILE"`

	// MCPSSLCertfile is the path to the SSL certificate file (optional).
	MCPSSLCertfile string `mapstructure:"MCP_SSL_CERTFILE"`

	// MCPTransportProtocol specifies the transport protocol ("stdio", "streamable-http", "sse", "http").
	MCPTransportProtocol string `mapstructure:"MCP_TRANSPORT_PROTOCOL"`

	// MCPHostEndpoint is the full URL endpoint of the server.
	MCPHostEndpoint string `mapstructure:"MCP_HOST_ENDPOINT"`

	// Environment specifies the deployment environment ("development", "staging", "production").
	Environment string `mapstructure:"ENVIRONMENT"`

	// Logging configuration

	// LogLevel sets the logging verbosity ("DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL").
	LogLevel string `mapstructure:"LOG_LEVEL"`

	// CORS configuration

	// CORSEnabled controls whether CORS middleware is enabled.
	CORSEnabled bool `mapstructure:"CORS_ENABLED"`

	// CORSOrigins lists allowed origins for CORS requests.
	CORSOrigins []string `mapstructure:"CORS_ORIGINS"`

	// CORSCredentials controls whether credentials are allowed in CORS requests.
	CORSCredentials bool `mapstructure:"CORS_CREDENTIALS"`

	// CORSMethods lists allowed HTTP methods for CORS requests.
	CORSMethods []string `mapstructure:"CORS_METHODS"`

	// CORSHeaders lists allowed headers for CORS requests.
	CORSHeaders []string `mapstructure:"CORS_HEADERS"`

	// SSO/OAuth configuration

	// SSOClientID is the OAuth client ID for SSO authentication.
	SSOClientID string `mapstructure:"SSO_CLIENT_ID"`

	// SSOClientSecret is the OAuth client secret for SSO authentication.
	SSOClientSecret string `mapstructure:"SSO_CLIENT_SECRET"`

	// SSOCallbackURL is the OAuth callback/redirect URL.
	SSOCallbackURL string `mapstructure:"SSO_CALLBACK_URL"`

	// SSOAuthorizationURL is the OAuth authorization endpoint.
	SSOAuthorizationURL string `mapstructure:"SSO_AUTHORIZATION_URL"`

	// SSOTokenURL is the OAuth token exchange endpoint.
	SSOTokenURL string `mapstructure:"SSO_TOKEN_URL"`

	// SSOIntrospectionURL is the OAuth token introspection endpoint.
	SSOIntrospectionURL string `mapstructure:"SSO_INTROSPECTION_URL"`

	// SessionSecret is the secret key used for session encryption.
	SessionSecret string `mapstructure:"SESSION_SECRET"`

	// UseExternalBrowserAuth enables external browser-based authentication.
	UseExternalBrowserAuth bool `mapstructure:"USE_EXTERNAL_BROWSER_AUTH"`

	// CompatibleWithCursor enables Cursor IDE compatibility mode.
	CompatibleWithCursor bool `mapstructure:"COMPATIBLE_WITH_CURSOR"`

	// CursorCompatibleSSE enables SSE streaming compatible with Cursor IDE.
	CursorCompatibleSSE bool `mapstructure:"CURSOR_COMPATIBLE_SSE"`

	// EnableAuth controls whether authentication is required.
	EnableAuth bool `mapstructure:"ENABLE_AUTH"`

	// PostgreSQL configuration

	// PostgresHost is the PostgreSQL server hostname or IP address.
	PostgresHost string `mapstructure:"POSTGRES_HOST"`

	// PostgresPort is the PostgreSQL server port (default: 5432).
	PostgresPort int `mapstructure:"POSTGRES_PORT"`

	// PostgresDB is the name of the PostgreSQL database.
	PostgresDB string `mapstructure:"POSTGRES_DB"`

	// PostgresUser is the PostgreSQL authentication username.
	PostgresUser string `mapstructure:"POSTGRES_USER"`

	// PostgresPassword is the PostgreSQL authentication password.
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`

	// PostgresPoolSize is the initial number of connections in the pool.
	PostgresPoolSize int `mapstructure:"POSTGRES_POOL_SIZE"`

	// PostgresMaxConnections is the maximum number of concurrent connections.
	PostgresMaxConnections int `mapstructure:"POSTGRES_MAX_CONNECTIONS"`
}

// Load loads configuration from environment variables and optional .env file.
// It applies default values, reads configuration, and performs validation.
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Automatically read environment variables
	v.AutomaticEnv()

	// Allow viper to read from .env file if present
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("..")

	// Read config file if it exists (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is acceptable
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values.
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("MCP_HOST", "localhost")
	v.SetDefault("MCP_PORT", 8080)
	v.SetDefault("MCP_TRANSPORT_PROTOCOL", "http")
	v.SetDefault("MCP_HOST_ENDPOINT", "http://localhost:8080")
	v.SetDefault("ENVIRONMENT", "development")

	// Logging defaults
	v.SetDefault("LOG_LEVEL", "INFO")

	// CORS defaults
	v.SetDefault("CORS_ENABLED", false)
	v.SetDefault("CORS_ORIGINS", []string{"*"})
	v.SetDefault("CORS_CREDENTIALS", true)
	v.SetDefault("CORS_METHODS", []string{"*"})
	v.SetDefault("CORS_HEADERS", []string{"*"})

	// OAuth defaults
	v.SetDefault("USE_EXTERNAL_BROWSER_AUTH", false)
	v.SetDefault("COMPATIBLE_WITH_CURSOR", true)
	v.SetDefault("CURSOR_COMPATIBLE_SSE", true)
	v.SetDefault("ENABLE_AUTH", true)

	// PostgreSQL defaults
	v.SetDefault("POSTGRES_PORT", 5432)
	v.SetDefault("POSTGRES_POOL_SIZE", 10)
	v.SetDefault("POSTGRES_MAX_CONNECTIONS", 20)
}

// Validate performs validation on configuration values.
// It returns an error if any configuration value is invalid.
func (c *Config) Validate() error {
	// Validate port range
	if c.MCPPort < 1024 || c.MCPPort > 65535 {
		return fmt.Errorf("MCP_PORT must be between 1024 and 65535, got %d", c.MCPPort)
	}

	// Validate log level
	validLogLevels := []string{"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"}
	logLevel := strings.ToUpper(c.LogLevel)
	valid := false
	for _, level := range validLogLevels {
		if logLevel == level {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("LOG_LEVEL must be one of %v, got %s", validLogLevels, c.LogLevel)
	}

	// Validate transport protocol
	validProtocols := []string{"stdio", "streamable-http", "sse", "http"}
	protocolValid := false
	for _, protocol := range validProtocols {
		if c.MCPTransportProtocol == protocol {
			protocolValid = true
			break
		}
	}
	if !protocolValid {
		return fmt.Errorf("MCP_TRANSPORT_PROTOCOL must be one of %v, got %s", validProtocols, c.MCPTransportProtocol)
	}

	// Validate SSL configuration
	if (c.MCPSSLKeyfile != "" && c.MCPSSLCertfile == "") || (c.MCPSSLKeyfile == "" && c.MCPSSLCertfile != "") {
		return fmt.Errorf("both MCP_SSL_KEYFILE and MCP_SSL_CERTFILE must be set together")
	}

	// Validate PostgreSQL port if specified
	if c.PostgresPort != 0 && (c.PostgresPort < 1024 || c.PostgresPort > 65535) {
		return fmt.Errorf("POSTGRES_PORT must be between 1024 and 65535, got %d", c.PostgresPort)
	}

	// Validate session secret in production
	if strings.ToLower(c.Environment) == "production" && c.SessionSecret == "" {
		return fmt.Errorf("SESSION_SECRET must be explicitly set in production environment")
	}

	// Note: PostgreSQL configuration is optional - if not provided, in-memory storage will be used
	// Only validate PostgreSQL config if auth is enabled and at least one Postgres field is set
	if c.EnableAuth {
		hasAnyPostgresConfig := c.PostgresHost != "" || c.PostgresDB != "" || c.PostgresUser != ""
		if hasAnyPostgresConfig {
			// If any Postgres config is provided, validate that required fields are present
			if c.PostgresHost == "" || c.PostgresDB == "" || c.PostgresUser == "" {
				return fmt.Errorf("if PostgreSQL is configured, POSTGRES_HOST, POSTGRES_DB, and POSTGRES_USER are required")
			}
		}
	}

	return nil
}

// GetServerAddress returns the server address string in "host:port" format.
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.MCPHost, c.MCPPort)
}

// GetPostgresConnectionString returns the PostgreSQL connection string.
func (c *Config) GetPostgresConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
	)
}

// HasSSL returns true if SSL is configured (both key and cert files are set).
func (c *Config) HasSSL() bool {
	return c.MCPSSLKeyfile != "" && c.MCPSSLCertfile != ""
}

// GetSessionSecret returns the session secret, generating one for development if needed.
// In production, a secret must be explicitly configured.
func (c *Config) GetSessionSecret() string {
	if c.SessionSecret != "" {
		return c.SessionSecret
	}

	// Generate ephemeral key for development
	if strings.ToLower(c.Environment) != "production" {
		return generateEphemeralKey()
	}

	return ""
}

// generateEphemeralKey creates a temporary session secret for development environments.
// This should NEVER be used in production.
func generateEphemeralKey() string {
	// This is a simple implementation; in production use crypto/rand
	return fmt.Sprintf("dev-ephemeral-key-%d", time.Now().UnixNano())
}
