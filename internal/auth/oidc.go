package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// AudienceClaimType handles both single string and array audiences in JWT tokens
type AudienceClaimType []string

// UnmarshalJSON implements custom JSON unmarshaling for audience claims
func (a *AudienceClaimType) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as array first
	var audiences []string
	if err := json.Unmarshal(data, &audiences); err == nil {
		*a = AudienceClaimType(audiences)
		return nil
	}

	// If that fails, try to unmarshal as single string
	var audience string
	if err := json.Unmarshal(data, &audience); err != nil {
		return fmt.Errorf("audience claim must be either string or array of strings: %w", err)
	}

	*a = AudienceClaimType([]string{audience})
	return nil
}

// MarshalJSON implements custom JSON marshaling for audience claims
func (a AudienceClaimType) MarshalJSON() ([]byte, error) {
	if len(a) == 1 {
		return json.Marshal(a[0])
	}
	return json.Marshal([]string(a))
}

// String returns the first audience or empty string if none
func (a AudienceClaimType) String() string {
	if len(a) > 0 {
		return a[0]
	}
	return ""
}

// Contains checks if the audience list contains a specific audience
func (a AudienceClaimType) Contains(audience string) bool {

	if slices.Contains(a, audience) {
		return true
	}
	return false

}

// ToSlice returns the audience as a string slice
func (a AudienceClaimType) ToSlice() []string {
	return []string(a)
}

// OIDCConfig holds OpenID Connect configuration
type OIDCConfig struct {
	ProviderURL     string
	ClientID        string
	ClientSecret    string
	RedirectURL     string
	Scopes          []string
	StateCookieName string
}

// OIDCProvider wraps the OIDC provider and OAuth2 config
type OIDCProvider struct {
	provider        *oidc.Provider
	config          *oauth2.Config
	verifier        *oidc.IDTokenVerifier
	stateCookieName string
}

// UserInfo represents user information from OIDC
type UserInfo struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Picture  string    `json:"picture"`
	Provider string    `json:"provider"`
}

// OIDCClaims represents OIDC token claims
type OIDCClaims struct {
	UserID   uuid.UUID         `json:"user_id"`
	Email    string            `json:"email"`
	Name     string            `json:"name"`
	Provider string            `json:"provider"`
	Issuer   string            `json:"issuer"`
	Subject  string            `json:"sub"`
	Audience AudienceClaimType `json:"aud"`
	Expires  int64             `json:"exp"`
	IssuedAt int64             `json:"iat"`
}

// NewOIDCProvider creates a new OIDC provider instance
func NewOIDCProvider(config *OIDCConfig) (*OIDCProvider, error) {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, config.ProviderURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)

	}

	oidcConfig := &oidc.Config{
		ClientID: config.ClientID,
	}

	verifier := provider.Verifier(oidcConfig)

	log.Println("[+] scopes:", config.Scopes)

	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       config.Scopes,
	}

	return &OIDCProvider{
		provider:        provider,
		config:          oauth2Config,
		verifier:        verifier,
		stateCookieName: config.StateCookieName,
	}, nil
}

// GetAuthURL generates the OAuth2 authorization URL
func (p *OIDCProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// GetAuthURLWithOptions generates the OAuth2 authorization URL with extra options
func (p *OIDCProvider) GetAuthURLWithOptions(state string, opts ...oauth2.AuthCodeOption) string {
	all := append([]oauth2.AuthCodeOption{oauth2.AccessTypeOffline}, opts...)
	return p.config.AuthCodeURL(state, all...)
}

// ExchangeCode exchanges authorization code for tokens
func (p *OIDCProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

// ExchangeCodeWithOptions exchanges the code with additional options (e.g., PKCE code_verifier)
func (p *OIDCProvider) ExchangeCodeWithOptions(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code, opts...)
}

// VerifyIDToken verifies the ID token and extracts user information
func (p *OIDCProvider) VerifyIDToken(ctx context.Context, idToken string) (*OIDCClaims, error) {
	token, err := p.verifier.Verify(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	var claims OIDCClaims
	if err := token.Claims(&claims); err != nil {
		// log.Printf("\n\n[+] Claims: %v \n\n", claims)
		return nil, fmt.Errorf("failed to parse token claims: %w", err)
	}

	// Extract user info from token
	claims.UserID = uuid.MustParse(token.Subject)
	claims.Email = token.Subject   // You might want to extract this from custom claims
	claims.Provider = token.Issuer // Use issuer from token instead of provider
	claims.Issuer = token.Issuer
	claims.Subject = token.Subject
	claims.Audience = AudienceClaimType(token.Audience) //token.Audience[0]
	claims.Expires = token.Expiry.Unix()
	claims.IssuedAt = token.IssuedAt.Unix()

	return &claims, nil
}

// GetUserInfo retrieves additional user information from the OIDC provider
func (p *OIDCProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	userInfo, err := p.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	var user UserInfo
	if err := userInfo.Claims(&user); err != nil {
		return nil, fmt.Errorf("failed to parse user info claims: %w", err)
	}

	user.ID = uuid.MustParse(userInfo.Subject)
	user.Provider = "oidc" // Use default provider value

	return &user, nil
}

// GetConfig returns the OAuth2 configuration
func (p *OIDCProvider) GetConfig() *oauth2.Config {
	return p.config
}

// GetScopes returns the configured scopes
func (p *OIDCProvider) GetScopes() []string {
	return p.config.Scopes
}

// GetStateCookieName returns the state cookie name
func (p *OIDCProvider) GetStateCookieName() string {
	if p.stateCookieName != "" {
		return p.stateCookieName
	}
	return "oidc_state"
}

// SetStateCookie sets the state cookie for CSRF protection
func SetStateCookie(w http.ResponseWriter, state string, cookieName string) {
	// log.Printf("[+]state: %s\n[+]cookieName: %s\n[+]isCookieSecure: %t\n", state, cookieName, isCookieSecure())
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		Secure:   isCookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})
}

// GetStateCookie retrieves the state cookie
func GetStateCookie(r *http.Request, cookieName string) (string, error) {
	cookie, err := r.Cookie(cookieName)
	log.Println("[+]Cookie", cookie)
	if err != nil {
		return "", fmt.Errorf("state cookie not found: %w", err)
	}
	return cookie.Value, nil
}

// ClearStateCookie clears the state cookie
func ClearStateCookie(w http.ResponseWriter, cookieName string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isCookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})
}

// IsCookieSecure is a public version of isCookieSecure for use in handlers
func IsCookieSecure() bool {
	return isCookieSecure()
}

// isCookieSecure determines whether to set the Secure flag on cookies.
// Controlled via env var OIDC_COOKIE_SECURE. Defaults to true.
func isCookieSecure() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("OIDC_COOKIE_SECURE")))
	if value == "" {
		return true
	}
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return true
	}
}

// UsePKCE returns whether PKCE should be used based on env OIDC_USE_PKCE
func UsePKCE() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("OIDC_USE_PKCE")))
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

// GeneratePKCEVerifier generates a high-entropy code_verifier for PKCE
func GeneratePKCEVerifier() (string, error) {
	// Reuse existing random string generator but ensure length in [43, 128]
	v, err := GenerateRandomString(64)
	if err != nil {
		return "", err
	}
	// Base64 URL encoding may add padding; strip it
	v = strings.TrimRight(v, "=")
	if len(v) < 43 {
		// Pad with random characters if needed
		pad, err := GenerateRandomString(43 - len(v))
		if err != nil {
			return "", err
		}
		v += strings.TrimRight(pad, "=")
	}
	if len(v) > 128 {
		v = v[:128]
	}
	return v, nil
}

// PKCEChallenge computes the S256 code_challenge from a verifier
func PKCEChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	// Base64 URL without padding
	s := base64.RawURLEncoding.EncodeToString(sum[:])
	return s
}

// PKCE cookie helpers
const pkceCookieName = "oidc_pkce_verifier"

func SetPKCECookie(w http.ResponseWriter, verifier string) {
	http.SetCookie(w, &http.Cookie{
		Name:     pkceCookieName,
		Value:    verifier,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		Secure:   isCookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})
}

func GetPKCECookie(r *http.Request) (string, error) {
	c, err := r.Cookie(pkceCookieName)
	if err != nil {
		return "", fmt.Errorf("pkce cookie not found: %w", err)
	}
	return c.Value, nil
}

func ClearPKCECookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     pkceCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isCookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})
}
