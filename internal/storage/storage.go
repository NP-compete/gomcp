package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/NP-compete/gomcp/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service provides PostgreSQL storage functionality
type Service struct {
	pool *pgxpool.Pool
	host string
	port int
	db   string
}

// NewService creates a new PostgreSQL storage service
func NewService(cfg Config) *Service {
	return &Service{
		host: cfg.Host,
		port: cfg.Port,
		db:   cfg.Database,
	}
}

// Connect establishes connection pool to PostgreSQL
func (s *Service) Connect(ctx context.Context, cfg Config) error {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse pool config: %w", err)
	}

	poolConfig.MinConns = int32(cfg.PoolSize)
	poolConfig.MaxConns = int32(cfg.MaxConnections)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	s.pool = pool

	// Create tables
	if err := s.createTables(ctx); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	logger.Info("Storage service connected to PostgreSQL")
	return nil
}

// Disconnect closes PostgreSQL connection pool
func (s *Service) Disconnect(ctx context.Context) error {
	if s.pool != nil {
		s.pool.Close()
		s.pool = nil
		logger.Info("Storage service disconnected from PostgreSQL")
	}
	return nil
}

// IsHealthy checks if PostgreSQL is healthy
func (s *Service) IsHealthy(ctx context.Context) bool {
	if s.pool == nil {
		return false
	}

	if err := s.pool.Ping(ctx); err != nil {
		logger.Warn(fmt.Sprintf("PostgreSQL health check failed: %v", err))
		return false
	}

	return true
}

func (s *Service) createTables(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS oauth_clients (
			client_id VARCHAR(255) PRIMARY KEY,
			client_secret VARCHAR(255) NOT NULL,
			client_name VARCHAR(255) NOT NULL,
			redirect_uris JSONB NOT NULL,
			grant_types JSONB NOT NULL,
			response_types JSONB NOT NULL,
			scope VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			CONSTRAINT unique_client_name_redirect UNIQUE (client_name, redirect_uris)
		)`,
		`CREATE TABLE IF NOT EXISTS oauth_authorization_codes (
			code VARCHAR(255) PRIMARY KEY,
			client_id VARCHAR(255) NOT NULL,
			redirect_uri VARCHAR(500) NOT NULL,
			scope VARCHAR(255),
			code_challenge VARCHAR(255) NOT NULL,
			code_challenge_method VARCHAR(10) NOT NULL,
			snowflake_token JSONB,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			state VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (client_id) REFERENCES oauth_clients(client_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS oauth_access_tokens (
			token VARCHAR(255) PRIMARY KEY,
			client_id VARCHAR(255) NOT NULL,
			scope VARCHAR(255),
			token_type VARCHAR(50) DEFAULT 'Bearer',
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (client_id) REFERENCES oauth_clients(client_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS oauth_refresh_tokens (
			token VARCHAR(255) PRIMARY KEY,
			client_id VARCHAR(255) NOT NULL,
			access_token VARCHAR(255),
			scope VARCHAR(255),
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (client_id) REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
			FOREIGN KEY (access_token) REFERENCES oauth_access_tokens(token) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_auth_codes_expires ON oauth_authorization_codes (expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_access_tokens_expires ON oauth_access_tokens (expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON oauth_refresh_tokens (expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_client_name ON oauth_clients (client_name)`,
	}

	for _, query := range queries {
		if _, err := s.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	logger.Info("OAuth database tables created successfully")
	return nil
}

// GetClientByNameAndRedirectURIs finds an existing client by name and redirect URIs
func (s *Service) GetClientByNameAndRedirectURIs(ctx context.Context, name string, redirectURIs []string) (*Client, error) {
	urisJSON, _ := json.Marshal(redirectURIs)

	var client Client
	var redirectURIsJSON []byte
	var grantTypesJSON []byte
	var responseTypesJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT client_id, client_secret, client_name, redirect_uris, grant_types, response_types, scope, created_at
		FROM oauth_clients
		WHERE client_name = $1 AND redirect_uris = $2
	`, name, urisJSON).Scan(
		&client.ID,
		&client.Secret,
		&client.Name,
		&redirectURIsJSON,
		&grantTypesJSON,
		&responseTypesJSON,
		&client.Scope,
		&client.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	json.Unmarshal(redirectURIsJSON, &client.RedirectURIs)
	json.Unmarshal(grantTypesJSON, &client.GrantTypes)
	json.Unmarshal(responseTypesJSON, &client.ResponseTypes)

	return &client, nil
}

// StoreClient stores a new OAuth client
func (s *Service) StoreClient(ctx context.Context, client *Client) error {
	redirectURIsJSON, _ := json.Marshal(client.RedirectURIs)
	grantTypesJSON, _ := json.Marshal(client.GrantTypes)
	responseTypesJSON, _ := json.Marshal(client.ResponseTypes)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO oauth_clients (client_id, client_secret, client_name, redirect_uris, grant_types, response_types, scope)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, client.ID, client.Secret, client.Name, redirectURIsJSON, grantTypesJSON, responseTypesJSON, client.Scope)

	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Stored client: %s", client.ID))
	return nil
}

// GetClient gets a client by ID
func (s *Service) GetClient(ctx context.Context, clientID string) (*Client, error) {
	var client Client
	var redirectURIsJSON []byte
	var grantTypesJSON []byte
	var responseTypesJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT client_id, client_secret, client_name, redirect_uris, grant_types, response_types, scope, created_at
		FROM oauth_clients
		WHERE client_id = $1
	`, clientID).Scan(
		&client.ID,
		&client.Secret,
		&client.Name,
		&redirectURIsJSON,
		&grantTypesJSON,
		&responseTypesJSON,
		&client.Scope,
		&client.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	json.Unmarshal(redirectURIsJSON, &client.RedirectURIs)
	json.Unmarshal(grantTypesJSON, &client.GrantTypes)
	json.Unmarshal(responseTypesJSON, &client.ResponseTypes)

	return &client, nil
}

// StoreAuthorizationCode stores an authorization code
func (s *Service) StoreAuthorizationCode(ctx context.Context, code *AuthorizationCode) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO oauth_authorization_codes (code, client_id, redirect_uri, scope, code_challenge, code_challenge_method, expires_at, state)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, code.Code, code.ClientID, code.RedirectURI, code.Scope, code.CodeChallenge, code.CodeChallengeMethod, code.ExpiresAt, code.State)

	return err
}

// GetAuthorizationCode gets authorization code data
func (s *Service) GetAuthorizationCode(ctx context.Context, code string) (*AuthorizationCode, error) {
	var authCode AuthorizationCode
	var snowflakeTokenJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT code, client_id, redirect_uri, scope, code_challenge, code_challenge_method, snowflake_token, expires_at, state
		FROM oauth_authorization_codes
		WHERE code = $1
	`, code).Scan(
		&authCode.Code,
		&authCode.ClientID,
		&authCode.RedirectURI,
		&authCode.Scope,
		&authCode.CodeChallenge,
		&authCode.CodeChallengeMethod,
		&snowflakeTokenJSON,
		&authCode.ExpiresAt,
		&authCode.State,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if len(snowflakeTokenJSON) > 0 {
		json.Unmarshal(snowflakeTokenJSON, &authCode.SnowflakeToken)
	}

	return &authCode, nil
}

// UpdateAuthorizationCodeToken updates authorization code with token
func (s *Service) UpdateAuthorizationCodeToken(ctx context.Context, code string, token map[string]interface{}) error {
	tokenJSON, _ := json.Marshal(token)
	_, err := s.pool.Exec(ctx, `
		UPDATE oauth_authorization_codes SET snowflake_token = $2 WHERE code = $1
	`, code, tokenJSON)
	return err
}

// DeleteAuthorizationCode deletes an authorization code
func (s *Service) DeleteAuthorizationCode(ctx context.Context, code string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM oauth_authorization_codes WHERE code = $1`, code)
	return err
}

// StoreAccessToken stores an access token
func (s *Service) StoreAccessToken(ctx context.Context, token *AccessToken) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO oauth_access_tokens (token, client_id, scope, expires_at)
		VALUES ($1, $2, $3, $4)
	`, token.Token, token.ClientID, token.Scope, token.ExpiresAt)
	return err
}

// GetAccessToken gets access token data
func (s *Service) GetAccessToken(ctx context.Context, token string) (*AccessToken, error) {
	var accessToken AccessToken

	err := s.pool.QueryRow(ctx, `
		SELECT token, client_id, scope, token_type, expires_at
		FROM oauth_access_tokens
		WHERE token = $1
	`, token).Scan(&accessToken.Token, &accessToken.ClientID, &accessToken.Scope, &accessToken.TokenType, &accessToken.ExpiresAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &accessToken, nil
}

// DeleteAccessToken deletes an access token
func (s *Service) DeleteAccessToken(ctx context.Context, token string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM oauth_access_tokens WHERE token = $1`, token)
	return err
}

// StoreRefreshToken stores a refresh token
func (s *Service) StoreRefreshToken(ctx context.Context, token *RefreshToken) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO oauth_refresh_tokens (token, client_id, access_token, scope, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, token.Token, token.ClientID, token.AccessToken, token.Scope, token.ExpiresAt)
	return err
}

// GetRefreshToken gets refresh token data
func (s *Service) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	var refreshToken RefreshToken

	err := s.pool.QueryRow(ctx, `
		SELECT token, client_id, access_token, scope, expires_at
		FROM oauth_refresh_tokens
		WHERE token = $1
	`, token).Scan(&refreshToken.Token, &refreshToken.ClientID, &refreshToken.AccessToken, &refreshToken.Scope, &refreshToken.ExpiresAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &refreshToken, nil
}

// DeleteRefreshToken deletes a refresh token
func (s *Service) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM oauth_refresh_tokens WHERE token = $1`, token)
	return err
}

// GetStatus returns storage service status
func (s *Service) GetStatus(ctx context.Context) map[string]interface{} {
	status := map[string]interface{}{
		"type":     "postgresql",
		"healthy":  s.IsHealthy(ctx),
		"host":     s.host,
		"port":     s.port,
		"database": s.db,
	}

	if s.pool != nil {
		stat := s.pool.Stat()
		status["pool_size"] = stat.TotalConns()
		status["pool_max_size"] = stat.MaxConns()
	}

	return status
}
