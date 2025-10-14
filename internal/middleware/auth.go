package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/NP-compete/gomcp/internal/logger"
	"github.com/NP-compete/gomcp/internal/oauth"
	"github.com/NP-compete/gomcp/internal/storage"
)

type contextKey string

const (
	// TokenInfoKey is the context key for token info
	TokenInfoKey contextKey = "token_info"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(oauthService *oauth.Service, enabled bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Public paths that don't require authentication
			publicPaths := map[string]bool{
				"/.well-known/oauth-protected-resource":   true,
				"/.well-known/oauth-authorization-server": true,
				"/docs":                    true,
				"/redoc":                   true,
				"/openapi.json":            true,
				"/auth/authorize":          true,
				"/auth/token":              true,
				"/auth/revoke":             true,
				"/auth/introspect":         true,
				"/auth/register":           true,
				"/auth/callback":           true,
				"/auth/callback/snowflake": true,
				"/auth/callback/oidc":      true,
				"/health":                  true,
			}

			if publicPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// Check Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Warn("Missing Authorization header for protected route: " + r.URL.Path)
				w.Header().Set("WWW-Authenticate", "Bearer")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Verify token
			tokenInfo, err := verifyAuthorizationHeader(r.Context(), oauthService, authHeader)
			if err != nil || tokenInfo == nil {
				logger.Warn("Invalid token for protected route: " + r.URL.Path)
				w.Header().Set("WWW-Authenticate", "Bearer")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add token info to context
			ctx := context.WithValue(r.Context(), TokenInfoKey, tokenInfo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// verifyAuthorizationHeader verifies the Authorization header and returns token info
func verifyAuthorizationHeader(ctx context.Context, oauthService *oauth.Service, authHeader string) (*storage.AccessToken, error) {
	// Extract Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, nil
	}

	token := parts[1]

	// Retrieve and validate token
	tokenInfo, err := oauthService.RetrieveAccessToken(ctx, token)
	if err != nil || tokenInfo == nil {
		return nil, err
	}

	return tokenInfo, nil
}
