package storage

import (
	"context"
	"time"
)

// Client represents an OAuth 2.0 client registered with the server.
// Clients are identified by their ID and authenticated using their Secret.
type Client struct {
	// ID is the unique identifier for this client.
	ID string `json:"id"`

	// Secret is the client's authentication credential.
	// This should be stored securely and never exposed in logs.
	Secret string `json:"secret"`

	// Name is a human-readable name for the client.
	Name string `json:"name"`

	// RedirectURIs are the allowed callback URLs for this client.
	RedirectURIs []string `json:"redirect_uris"`

	// GrantTypes specifies which OAuth grant types this client can use.
	// Common values: "authorization_code", "refresh_token", "client_credentials"
	GrantTypes []string `json:"grant_types"`

	// ResponseTypes specifies allowed response types.
	// Common values: "code", "token"
	ResponseTypes []string `json:"response_types"`

	// Scope defines the default scope for this client.
	Scope string `json:"scope"`

	// CreatedAt is when this client was registered.
	CreatedAt time.Time `json:"created_at"`
}

// AuthorizationCode represents an OAuth 2.0 authorization code.
// Authorization codes are short-lived credentials exchanged for access tokens.
type AuthorizationCode struct {
	// Code is the authorization code value.
	Code string `json:"code"`

	// ClientID identifies which client this code was issued to.
	ClientID string `json:"client_id"`

	// RedirectURI is the callback URL that must match during token exchange.
	RedirectURI string `json:"redirect_uri"`

	// Scope defines the permissions granted by this authorization.
	Scope string `json:"scope"`

	// CodeChallenge is the PKCE code challenge (if using PKCE).
	CodeChallenge string `json:"code_challenge"`

	// CodeChallengeMethod specifies the PKCE challenge method ("S256" or "plain").
	CodeChallengeMethod string `json:"code_challenge_method"`

	// SnowflakeToken contains Snowflake-specific token data.
	SnowflakeToken map[string]interface{} `json:"snowflake_token,omitempty"`

	// ExpiresAt is when this authorization code expires.
	ExpiresAt time.Time `json:"expires_at"`

	// State is the CSRF protection state parameter.
	State string `json:"state"`
}

// AccessToken represents an OAuth 2.0 access token.
// Access tokens grant access to protected resources.
type AccessToken struct {
	// Token is the access token value.
	Token string `json:"token"`

	// ClientID identifies which client this token was issued to.
	ClientID string `json:"client_id"`

	// Scope defines the permissions granted by this token.
	Scope string `json:"scope"`

	// TokenType specifies the token type (typically "Bearer").
	TokenType string `json:"token_type"`

	// ExpiresAt is when this access token expires.
	ExpiresAt time.Time `json:"expires_at"`
}

// RefreshToken represents an OAuth 2.0 refresh token.
// Refresh tokens are used to obtain new access tokens without re-authentication.
type RefreshToken struct {
	// Token is the refresh token value.
	Token string `json:"token"`

	// ClientID identifies which client this token was issued to.
	ClientID string `json:"client_id"`

	// AccessToken is the associated access token.
	AccessToken string `json:"access_token"`

	// Scope defines the permissions granted by this token.
	Scope string `json:"scope"`

	// ExpiresAt is when this refresh token expires.
	ExpiresAt time.Time `json:"expires_at"`
}

// Store defines the interface for persistent storage operations.
// Implementations must be safe for concurrent use.
type Store interface {
	// Connect establishes a connection to the storage backend.
	// It should be called before any other operations.
	Connect(ctx context.Context, cfg Config) error

	// Disconnect closes the connection to the storage backend.
	// It should be called during graceful shutdown.
	Disconnect(ctx context.Context) error

	// IsHealthy returns true if the storage backend is responsive.
	IsHealthy(ctx context.Context) bool

	// GetClientByNameAndRedirectURIs retrieves a client by name and redirect URIs.
	// Returns nil if no matching client is found.
	GetClientByNameAndRedirectURIs(ctx context.Context, name string, redirectURIs []string) (*Client, error)

	// StoreClient persists a client to storage.
	StoreClient(ctx context.Context, client *Client) error

	// GetClient retrieves a client by ID.
	// Returns nil if the client doesn't exist.
	GetClient(ctx context.Context, clientID string) (*Client, error)

	// StoreAuthorizationCode persists an authorization code to storage.
	StoreAuthorizationCode(ctx context.Context, code *AuthorizationCode) error

	// GetAuthorizationCode retrieves an authorization code by value.
	// Returns nil if the code doesn't exist or has expired.
	GetAuthorizationCode(ctx context.Context, code string) (*AuthorizationCode, error)

	// UpdateAuthorizationCodeToken updates the Snowflake token associated with a code.
	UpdateAuthorizationCodeToken(ctx context.Context, code string, token map[string]interface{}) error

	// DeleteAuthorizationCode removes an authorization code from storage.
	DeleteAuthorizationCode(ctx context.Context, code string) error

	// StoreAccessToken persists an access token to storage.
	StoreAccessToken(ctx context.Context, token *AccessToken) error

	// GetAccessToken retrieves an access token by value.
	// Returns nil if the token doesn't exist or has expired.
	GetAccessToken(ctx context.Context, token string) (*AccessToken, error)

	// DeleteAccessToken removes an access token from storage.
	DeleteAccessToken(ctx context.Context, token string) error

	// StoreRefreshToken persists a refresh token to storage.
	StoreRefreshToken(ctx context.Context, token *RefreshToken) error

	// GetRefreshToken retrieves a refresh token by value.
	// Returns nil if the token doesn't exist or has expired.
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error)

	// DeleteRefreshToken removes a refresh token from storage.
	DeleteRefreshToken(ctx context.Context, token string) error

	// GetStatus returns diagnostic information about the storage backend.
	GetStatus(ctx context.Context) map[string]interface{}
}

// Config holds storage configuration.
type Config struct {
	// Host is the database server hostname or IP address.
	Host string

	// Port is the database server port.
	Port int

	// Database is the name of the database to connect to.
	Database string

	// Username is the database authentication username.
	Username string

	// Password is the database authentication password.
	Password string

	// PoolSize is the initial number of connections in the pool.
	PoolSize int

	// MaxConnections is the maximum number of concurrent connections allowed.
	MaxConnections int
}
