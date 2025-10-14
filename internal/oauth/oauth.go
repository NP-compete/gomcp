package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/NP-compete/gomcp/internal/logger"
	"github.com/NP-compete/gomcp/internal/storage"
)

// Service manages OAuth 2.0 operations with dependency injection
type Service struct {
	storage storage.Store
}

// NewService creates a new OAuth service
func NewService(store storage.Store) *Service {
	return &Service{
		storage: store,
	}
}

// generateRandomString generates a cryptographically secure random string
func generateRandomString(length int) (string, error) {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}
	return string(bytes), nil
}

// base64URLEncode encodes data to base64 URL-safe format without padding
func base64URLEncode(data []byte) string {
	encoded := base64.URLEncoding.EncodeToString(data)
	// Remove padding
	for len(encoded) > 0 && encoded[len(encoded)-1] == '=' {
		encoded = encoded[:len(encoded)-1]
	}
	return encoded
}

// VerifyCodeChallenge verifies PKCE code challenge using S256 method
func VerifyCodeChallenge(codeVerifier, codeChallenge string) bool {
	hash := sha256.Sum256([]byte(codeVerifier))
	computedChallenge := base64URLEncode(hash[:])
	return computedChallenge == codeChallenge
}

// ValidateClient validates client credentials
func (s *Service) ValidateClient(ctx context.Context, clientID string, clientSecret string) (*storage.Client, error) {
	client, err := s.storage.GetClient(ctx, clientID)
	if err != nil || client == nil {
		return nil, err
	}

	if clientSecret != "" && client.Secret != "" && client.Secret != clientSecret {
		return nil, fmt.Errorf("invalid client secret")
	}

	return client, nil
}

// CreateAuthorizationCode creates an authorization code
func (s *Service) CreateAuthorizationCode(ctx context.Context, clientID, redirectURI, scope, codeChallenge, codeChallengeMethod, state string) (string, error) {
	code, err := generateRandomString(32)
	if err != nil {
		return "", err
	}

	authCode := &storage.AuthorizationCode{
		Code:                code,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		Scope:               scope,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		ExpiresAt:           time.Now().Add(10 * time.Minute),
		State:               state,
	}

	if err := s.storage.StoreAuthorizationCode(ctx, authCode); err != nil {
		return "", err
	}

	return code, nil
}

// AddTokenToCode updates authorization code with external token
func (s *Service) AddTokenToCode(ctx context.Context, code string, tokenSet map[string]interface{}) error {
	return s.storage.UpdateAuthorizationCodeToken(ctx, code, tokenSet)
}

// ValidateAuthorizationCode validates an authorization code
func (s *Service) ValidateAuthorizationCode(ctx context.Context, code string) (*storage.AuthorizationCode, error) {
	authCode, err := s.storage.GetAuthorizationCode(ctx, code)
	if err != nil || authCode == nil {
		return nil, err
	}

	if time.Now().After(authCode.ExpiresAt) {
		return nil, fmt.Errorf("authorization code expired")
	}

	return authCode, nil
}

// MarkCodeAsUsed marks authorization code as used by deleting it
func (s *Service) MarkCodeAsUsed(ctx context.Context, code string) error {
	if err := s.storage.DeleteAuthorizationCode(ctx, code); err != nil {
		logger.Warn(fmt.Sprintf("Failed to delete authorization code: %s", code[:8]))
		return err
	}
	logger.Info(fmt.Sprintf("Authorization code deleted after use: %s...", code[:8]))
	return nil
}

// ValidateRefreshToken validates a refresh token
func (s *Service) ValidateRefreshToken(ctx context.Context, refreshToken string) (*storage.RefreshToken, error) {
	token, err := s.storage.GetRefreshToken(ctx, refreshToken)
	if err != nil || token == nil {
		return nil, err
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("refresh token expired")
	}

	return token, nil
}

// RegisterClientRequest represents a client registration request
type RegisterClientRequest struct {
	ClientName    string   `json:"client_name"`
	RedirectURIs  []string `json:"redirect_uris"`
	GrantTypes    []string `json:"grant_types,omitempty"`
	ResponseTypes []string `json:"response_types,omitempty"`
	Scope         string   `json:"scope,omitempty"`
}

// RegisterClientResponse represents a client registration response
type RegisterClientResponse struct {
	ClientID         string   `json:"client_id"`
	ClientSecret     string   `json:"client_secret"`
	ClientName       string   `json:"client_name"`
	RedirectURIs     []string `json:"redirect_uris"`
	GrantTypes       []string `json:"grant_types"`
	ResponseTypes    []string `json:"response_types"`
	Scope            string   `json:"scope"`
	ClientIDIssuedAt int64    `json:"client_id_issued_at"`
}

// RegisterClient registers a new OAuth client
func (s *Service) RegisterClient(ctx context.Context, req RegisterClientRequest) (*RegisterClientResponse, error) {
	// Check if client already exists
	existingClient, err := s.storage.GetClientByNameAndRedirectURIs(ctx, req.ClientName, req.RedirectURIs)
	if err != nil {
		return nil, err
	}

	if existingClient != nil {
		logger.Info(fmt.Sprintf("Returning existing client for name '%s': %s", req.ClientName, existingClient.ID))
		return &RegisterClientResponse{
			ClientID:         existingClient.ID,
			ClientSecret:     existingClient.Secret,
			ClientName:       existingClient.Name,
			RedirectURIs:     existingClient.RedirectURIs,
			GrantTypes:       existingClient.GrantTypes,
			ResponseTypes:    existingClient.ResponseTypes,
			Scope:            existingClient.Scope,
			ClientIDIssuedAt: existingClient.CreatedAt.Unix(),
		}, nil
	}

	// Generate new client credentials
	clientID, err := generateRandomString(16)
	if err != nil {
		return nil, err
	}

	clientSecret, err := generateRandomString(32)
	if err != nil {
		return nil, err
	}

	// Set defaults
	grantTypes := req.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []string{"authorization_code", "refresh_token"}
	}

	responseTypes := req.ResponseTypes
	if len(responseTypes) == 0 {
		responseTypes = []string{"code"}
	}

	scope := req.Scope
	if scope == "" {
		scope = "read write"
	}

	client := &storage.Client{
		ID:            clientID,
		Secret:        clientSecret,
		Name:          req.ClientName,
		RedirectURIs:  req.RedirectURIs,
		GrantTypes:    grantTypes,
		ResponseTypes: responseTypes,
		Scope:         scope,
		CreatedAt:     time.Now(),
	}

	if err := s.storage.StoreClient(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to persist client registration: %w", err)
	}

	logger.Info(fmt.Sprintf("New client registered: %s for '%s'", clientID, req.ClientName))

	return &RegisterClientResponse{
		ClientID:         clientID,
		ClientSecret:     clientSecret,
		ClientName:       req.ClientName,
		RedirectURIs:     req.RedirectURIs,
		GrantTypes:       grantTypes,
		ResponseTypes:    responseTypes,
		Scope:            scope,
		ClientIDIssuedAt: time.Now().Unix(),
	}, nil
}

// StoreAccessToken stores an access token
func (s *Service) StoreAccessToken(ctx context.Context, token string, clientID string, scope string, expiresIn int) error {
	accessToken := &storage.AccessToken{
		Token:     token,
		ClientID:  clientID,
		Scope:     scope,
		TokenType: "Bearer",
		ExpiresAt: time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
	return s.storage.StoreAccessToken(ctx, accessToken)
}

// RetrieveAccessToken retrieves an access token
func (s *Service) RetrieveAccessToken(ctx context.Context, token string) (*storage.AccessToken, error) {
	return s.storage.GetAccessToken(ctx, token)
}

// StoreRefreshToken stores a refresh token
func (s *Service) StoreRefreshToken(ctx context.Context, token string, clientID string, accessToken string, scope string, expiresIn int) error {
	refreshToken := &storage.RefreshToken{
		Token:       token,
		ClientID:    clientID,
		AccessToken: accessToken,
		Scope:       scope,
		ExpiresAt:   time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
	return s.storage.StoreRefreshToken(ctx, refreshToken)
}

// RetrieveRefreshToken retrieves a refresh token
func (s *Service) RetrieveRefreshToken(ctx context.Context, token string) (*storage.RefreshToken, error) {
	return s.storage.GetRefreshToken(ctx, token)
}

// RevokeAccessToken revokes (deletes) an access token
func (s *Service) RevokeAccessToken(ctx context.Context, token string) error {
	return s.storage.DeleteAccessToken(ctx, token)
}

// RevokeRefreshToken revokes (deletes) a refresh token
func (s *Service) RevokeRefreshToken(ctx context.Context, token string) error {
	return s.storage.DeleteRefreshToken(ctx, token)
}

// GetStorageStatus returns storage service status
func (s *Service) GetStorageStatus(ctx context.Context) map[string]interface{} {
	return s.storage.GetStatus(ctx)
}

// GenerateToken generates a secure random token
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
