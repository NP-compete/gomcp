package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	MCPHost              string `mapstructure:"MCP_HOST"`
	MCPPort              int    `mapstructure:"MCP_PORT"`
	MCPSSLKeyfile        string `mapstructure:"MCP_SSL_KEYFILE"`
	MCPSSLCertfile       string `mapstructure:"MCP_SSL_CERTFILE"`
	MCPTransportProtocol string `mapstructure:"MCP_TRANSPORT_PROTOCOL"`
	MCPHostEndpoint      string `mapstructure:"MCP_HOST_ENDPOINT"`
	Environment          string `mapstructure:"ENVIRONMENT"`

	// Logging configuration
	LogLevel string `mapstructure:"LOG_LEVEL"`

	// CORS configuration
	CORSEnabled     bool     `mapstructure:"CORS_ENABLED"`
	CORSOrigins     []string `mapstructure:"CORS_ORIGINS"`
	CORSCredentials bool     `mapstructure:"CORS_CREDENTIALS"`
	CORSMethods     []string `mapstructure:"CORS_METHODS"`
	CORSHeaders     []string `mapstructure:"CORS_HEADERS"`

	// SSO/OAuth configuration
	SSOClientID            string `mapstructure:"SSO_CLIENT_ID"`
	SSOClientSecret        string `mapstructure:"SSO_CLIENT_SECRET"`
	SSOCallbackURL         string `mapstructure:"SSO_CALLBACK_URL"`
	SSOAuthorizationURL    string `mapstructure:"SSO_AUTHORIZATION_URL"`
	SSOTokenURL            string `mapstructure:"SSO_TOKEN_URL"`
	SSOIntrospectionURL    string `mapstructure:"SSO_INTROSPECTION_URL"`
	SessionSecret          string `mapstructure:"SESSION_SECRET"`
	UseExternalBrowserAuth bool   `mapstructure:"USE_EXTERNAL_BROWSER_AUTH"`
	CompatibleWithCursor   bool   `mapstructure:"COMPATIBLE_WITH_CURSOR"`
	CursorCompatibleSSE    bool   `mapstructure:"CURSOR_COMPATIBLE_SSE"`
	EnableAuth             bool   `mapstructure:"ENABLE_AUTH"`

	// PostgreSQL configuration
	PostgresHost           string `mapstructure:"POSTGRES_HOST"`
	PostgresPort           int    `mapstructure:"POSTGRES_PORT"`
	PostgresDB             string `mapstructure:"POSTGRES_DB"`
	PostgresUser           string `mapstructure:"POSTGRES_USER"`
	PostgresPassword       string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresPoolSize       int    `mapstructure:"POSTGRES_POOL_SIZE"`
	PostgresMaxConnections int    `mapstructure:"POSTGRES_MAX_CONNECTIONS"`
}

// Load loads configuration from environment variables
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

// Validate performs validation on configuration values
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

// GetServerAddress returns the server address string
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.MCPHost, c.MCPPort)
}

// GetPostgresConnectionString returns the PostgreSQL connection string
func (c *Config) GetPostgresConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
	)
}

// HasSSL returns true if SSL is configured
func (c *Config) HasSSL() bool {
	return c.MCPSSLKeyfile != "" && c.MCPSSLCertfile != ""
}

// GetSessionSecret returns the session secret, generating one for development if needed
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

// generateEphemeralKey creates a temporary session secret for development
func generateEphemeralKey() string {
	// This is a simple implementation; in production use crypto/rand
	return fmt.Sprintf("dev-ephemeral-key-%d", time.Now().UnixNano())
}
