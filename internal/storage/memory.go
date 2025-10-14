package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/NP-compete/gomcp/internal/logger"
)

// MemoryStore implements Store interface using in-memory storage
type MemoryStore struct {
	mu                 sync.RWMutex
	clients            map[string]*Client
	clientsByName      map[string]map[string]*Client // name -> redirectURIs (JSON) -> client
	authorizationCodes map[string]*AuthorizationCode
	accessTokens       map[string]*AccessToken
	refreshTokens      map[string]*RefreshToken
}

// NewMemoryStore creates a new in-memory storage service
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		clients:            make(map[string]*Client),
		clientsByName:      make(map[string]map[string]*Client),
		authorizationCodes: make(map[string]*AuthorizationCode),
		accessTokens:       make(map[string]*AccessToken),
		refreshTokens:      make(map[string]*RefreshToken),
	}
}

// Connect is a no-op for in-memory storage
func (m *MemoryStore) Connect(ctx context.Context, cfg Config) error {
	logger.Info("Using in-memory storage (PostgreSQL not configured)")
	return nil
}

// Disconnect is a no-op for in-memory storage
func (m *MemoryStore) Disconnect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear all data
	m.clients = make(map[string]*Client)
	m.clientsByName = make(map[string]map[string]*Client)
	m.authorizationCodes = make(map[string]*AuthorizationCode)
	m.accessTokens = make(map[string]*AccessToken)
	m.refreshTokens = make(map[string]*RefreshToken)

	logger.Info("In-memory storage cleared")
	return nil
}

// IsHealthy always returns true for in-memory storage
func (m *MemoryStore) IsHealthy(ctx context.Context) bool {
	return true
}

// GetClientByNameAndRedirectURIs finds a client by name and redirect URIs
func (m *MemoryStore) GetClientByNameAndRedirectURIs(ctx context.Context, name string, redirectURIs []string) (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	urisJSON, _ := json.Marshal(redirectURIs)
	urisKey := string(urisJSON)

	if clientsMap, ok := m.clientsByName[name]; ok {
		if client, ok := clientsMap[urisKey]; ok {
			return client, nil
		}
	}

	return nil, nil
}

// StoreClient stores a new OAuth client
func (m *MemoryStore) StoreClient(ctx context.Context, client *Client) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[client.ID] = client

	// Index by name and redirect URIs
	if m.clientsByName[client.Name] == nil {
		m.clientsByName[client.Name] = make(map[string]*Client)
	}
	urisJSON, _ := json.Marshal(client.RedirectURIs)
	urisKey := string(urisJSON)
	m.clientsByName[client.Name][urisKey] = client

	logger.Info(fmt.Sprintf("Stored client in memory: %s", client.ID))
	return nil
}

// GetClient gets a client by ID
func (m *MemoryStore) GetClient(ctx context.Context, clientID string) (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, ok := m.clients[clientID]
	if !ok {
		return nil, nil
	}

	return client, nil
}

// StoreAuthorizationCode stores an authorization code
func (m *MemoryStore) StoreAuthorizationCode(ctx context.Context, code *AuthorizationCode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.authorizationCodes[code.Code] = code

	// Start cleanup goroutine for expired codes
	go m.cleanupExpiredCode(code.Code, code.ExpiresAt)

	return nil
}

// GetAuthorizationCode gets authorization code data
func (m *MemoryStore) GetAuthorizationCode(ctx context.Context, code string) (*AuthorizationCode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	authCode, ok := m.authorizationCodes[code]
	if !ok {
		return nil, nil
	}

	// Check if expired
	if time.Now().After(authCode.ExpiresAt) {
		return nil, nil
	}

	return authCode, nil
}

// UpdateAuthorizationCodeToken updates authorization code with token
func (m *MemoryStore) UpdateAuthorizationCodeToken(ctx context.Context, code string, token map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if authCode, ok := m.authorizationCodes[code]; ok {
		authCode.SnowflakeToken = token
		return nil
	}

	return fmt.Errorf("authorization code not found")
}

// DeleteAuthorizationCode deletes an authorization code
func (m *MemoryStore) DeleteAuthorizationCode(ctx context.Context, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.authorizationCodes, code)
	return nil
}

// StoreAccessToken stores an access token
func (m *MemoryStore) StoreAccessToken(ctx context.Context, token *AccessToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.accessTokens[token.Token] = token

	// Start cleanup goroutine for expired tokens
	go m.cleanupExpiredAccessToken(token.Token, token.ExpiresAt)

	return nil
}

// GetAccessToken gets access token data
func (m *MemoryStore) GetAccessToken(ctx context.Context, token string) (*AccessToken, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	accessToken, ok := m.accessTokens[token]
	if !ok {
		return nil, nil
	}

	// Check if expired
	if time.Now().After(accessToken.ExpiresAt) {
		return nil, nil
	}

	return accessToken, nil
}

// DeleteAccessToken deletes an access token
func (m *MemoryStore) DeleteAccessToken(ctx context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.accessTokens, token)
	return nil
}

// StoreRefreshToken stores a refresh token
func (m *MemoryStore) StoreRefreshToken(ctx context.Context, token *RefreshToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.refreshTokens[token.Token] = token

	// Start cleanup goroutine for expired tokens
	go m.cleanupExpiredRefreshToken(token.Token, token.ExpiresAt)

	return nil
}

// GetRefreshToken gets refresh token data
func (m *MemoryStore) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	refreshToken, ok := m.refreshTokens[token]
	if !ok {
		return nil, nil
	}

	// Check if expired
	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, nil
	}

	return refreshToken, nil
}

// DeleteRefreshToken deletes a refresh token
func (m *MemoryStore) DeleteRefreshToken(ctx context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.refreshTokens, token)
	return nil
}

// GetStatus returns storage service status
func (m *MemoryStore) GetStatus(ctx context.Context) map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"type":                      "in-memory",
		"healthy":                   true,
		"clients_count":             len(m.clients),
		"authorization_codes_count": len(m.authorizationCodes),
		"access_tokens_count":       len(m.accessTokens),
		"refresh_tokens_count":      len(m.refreshTokens),
	}
}

// Cleanup goroutines for expired items

func (m *MemoryStore) cleanupExpiredCode(code string, expiresAt time.Time) {
	time.Sleep(time.Until(expiresAt))
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.authorizationCodes, code)
}

func (m *MemoryStore) cleanupExpiredAccessToken(token string, expiresAt time.Time) {
	time.Sleep(time.Until(expiresAt))
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.accessTokens, token)
}

func (m *MemoryStore) cleanupExpiredRefreshToken(token string, expiresAt time.Time) {
	time.Sleep(time.Until(expiresAt))
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.refreshTokens, token)
}
