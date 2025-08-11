package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// OIDCMiddleware provides OIDC-aware authentication middleware
type OIDCMiddleware struct {
	oidcProvider *OIDCProvider
	jwtSecret    []byte
}

// NewOIDCMiddleware creates a new OIDC middleware instance
func NewOIDCMiddleware(oidcProvider *OIDCProvider, jwtSecret []byte) *OIDCMiddleware {
	return &OIDCMiddleware{
		oidcProvider: oidcProvider,
		jwtSecret:    jwtSecret,
	}
}

// Authenticate handles both OIDC and JWT authentication
func (m *OIDCMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to extract token from Authorization header
		tokenString, err := ExtractTokenFromHeader(r)
		if err != nil {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Try to validate as JWT first
		claims, err := m.validateJWT(tokenString)
		if err == nil {
			// JWT is valid, add user info to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "email", claims.Email)
			ctx = context.WithValue(ctx, "auth_type", "jwt")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// If JWT validation fails, try to validate as OIDC token
		oidcClaims, err := m.oidcProvider.VerifyIDToken(r.Context(), tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// OIDC token is valid, add user info to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", oidcClaims.UserID)
		ctx = context.WithValue(ctx, "email", oidcClaims.Email)
		ctx = context.WithValue(ctx, "auth_type", "oidc")
		ctx = context.WithValue(ctx, "provider", oidcClaims.Provider)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateJWT validates a JWT token and returns claims
func (m *OIDCMiddleware) validateJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RequireAuth is a simple middleware that requires authentication
func (m *OIDCMiddleware) RequireAuth(next http.Handler) http.Handler {
	return m.Authenticate(next)
}

// RequireRole is a middleware that requires a specific role
func (m *OIDCMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First authenticate the user
			m.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if user has the required role
				// This is a placeholder - you would implement role checking based on your user model
				userID := r.Context().Value("user_id")
				if userID == nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// TODO: Implement role checking logic
				// For now, we'll just allow authenticated users
				next.ServeHTTP(w, r)
			})).ServeHTTP(w, r)
		})
	}
}

// RefreshTokenMiddleware handles token refresh
func (m *OIDCMiddleware) RefreshTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if token is about to expire
		userID := r.Context().Value("user_id")
		email := r.Context().Value("email")
		authType := r.Context().Value("auth_type")

		if userID != nil && email != nil && authType == "jwt" {
			// For JWT tokens, we could implement refresh logic here
			// For now, we'll just pass through
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext extracts user information from request context
func GetUserFromContext(ctx context.Context) (uuid.UUID, string, string) {
	userID := ctx.Value("user_id")
	email := ctx.Value("email")
	authType := ctx.Value("auth_type")

	if userID == nil || email == nil {
		return uuid.Nil, "", ""
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, "", ""
	}

	emailStr, ok := email.(string)
	if !ok {
		return uuid.Nil, "", ""
	}

	authTypeStr, ok := authType.(string)
	if !ok {
		authTypeStr = "unknown"
	}

	return uid, emailStr, authTypeStr
}
