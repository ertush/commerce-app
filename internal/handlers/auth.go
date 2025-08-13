package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"commerce-app/internal/auth"

	"golang.org/x/oauth2"
)

// AuthHandler handles OIDC authentication
type AuthHandler struct {
	oidcProvider *auth.OIDCProvider
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler() (*AuthHandler, error) {
	config := &auth.OIDCConfig{
		ProviderURL:     getEnv("OIDC_PROVIDER_URL", "https://accounts.google.com"),
		ClientID:        getEnv("OIDC_CLIENT_ID", ""),
		ClientSecret:    getEnv("OIDC_CLIENT_SECRET", ""),
		RedirectURL:     getEnv("OIDC_REDIRECT_URL", "http://localhost:8181/api/auth/callback"),
		Scopes:          []string{"openid", "offline_access", "email"},
		StateCookieName: "oidc_state",
	}

	provider, err := auth.NewOIDCProvider(config)
	if err != nil {
		return nil, fmt.Errorf("Error creating OIDC provider: %w", err)
	}

	return &AuthHandler{
		oidcProvider: provider,
	}, nil
}

// Login initiates the OIDC login flow
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Generate state parameter for CSRF protection
	//
	state, err := auth.GenerateRandomString(32)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set state cookie
	auth.SetStateCookie(w, state, h.oidcProvider.GetStateCookieName())

	// Generate authorization URL (optionally with PKCE)
	authURL := ""
	if auth.UsePKCE() {
		verifier, err := auth.GeneratePKCEVerifier()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		// Persist verifier in cookie for the callback
		auth.SetPKCECookie(w, verifier)
		challenge := auth.PKCEChallenge(verifier)
		authURL = h.oidcProvider.GetAuthURLWithOptions(state,
			oauth2.SetAuthURLParam("code_challenge", challenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
	} else {
		authURL = h.oidcProvider.GetAuthURL(state)
	}

	// Redirect to OIDC provider
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// Callback handles the OIDC callback from the provider
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	error := r.URL.Query().Get("error")

	// Check for errors
	if error != "" {
		http.Error(w, fmt.Sprintf("OIDC error: %s", error), http.StatusBadRequest)
		return
	}

	// Validate required params before proceeding
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}
	if state == "" {
		http.Error(w, "Missing state parameter", http.StatusBadRequest)
		return
	}

	// Validate state parameter
	stateCookie, err := auth.GetStateCookie(r, h.oidcProvider.GetStateCookieName())
	if err != nil || state != stateCookie {
		http.Error(w, fmt.Sprintf("Invalid state parameter: %v\n state: %s", err, state), http.StatusBadRequest)
		return
	}

	// Clear state cookie
	auth.ClearStateCookie(w, h.oidcProvider.GetStateCookieName())

	// Exchange authorization code for tokens (use PKCE if enabled)
	ctx := r.Context()
	var token *oauth2.Token

	if auth.UsePKCE() {
		log.Println("[+] Using PKCE")
		verifier, err := auth.GetPKCECookie(r)
		if err != nil {
			http.Error(w, "Missing PKCE verifier", http.StatusBadRequest)
			return
		}
		// Clear PKCE cookie regardless of outcome
		defer auth.ClearPKCECookie(w)
		token, err = h.oidcProvider.ExchangeCodeWithOptions(ctx, code,
			oauth2.SetAuthURLParam("code_verifier", verifier),
		)
	} else {
		log.Println("[+] Exchanging code without PKCE")
		token, err = h.oidcProvider.ExchangeCode(ctx, code)
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange code: %v", err), http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		http.Error(w, "Missing id_token in token response", http.StatusInternalServerError)
		return
	}

	// log.Printf("token: %v\ncode: %v\nrawIDToken: %v", token, code, rawIDToken)

	// Verify ID token
	claims, err := h.oidcProvider.VerifyIDToken(ctx, rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to verify token: %v", err), http.StatusInternalServerError)
		return
	}

	// Get additional user info
	userInfo, err := h.oidcProvider.GetUserInfo(ctx, token)
	if err != nil {
		// Use claims if user info fails
		userInfo = &auth.UserInfo{
			ID:       claims.UserID,
			Email:    claims.Email,
			Name:     claims.Name,
			Provider: claims.Provider,
		}
	}

	// Generate JWT token for the user
	jwtToken, err := auth.GenerateToken(userInfo.ID, userInfo.Email)
	if err != nil {
		http.Error(w, "Failed to generate JWT token", http.StatusInternalServerError)
		return
	}

	// Return user info and JWT token
	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":       userInfo.ID,
			"email":    userInfo.Email,
			"name":     userInfo.Name,
			"picture":  userInfo.Picture,
			"provider": userInfo.Provider,
		},
		"access_token": jwtToken,
		"token_type":   "Bearer",
		"expires_in":   86400, // 24 hours
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract user info from context (set by middleware)
	_, email, authType := auth.GetUserFromContext(r.Context())

	// log.Printf("[+] Logout request - userID: %s, email: %s, authType: %s", userID, email, authType)

	// Clear OIDC-related cookies (if they exist)
	auth.ClearStateCookie(w, h.oidcProvider.GetStateCookieName())
	auth.ClearPKCECookie(w)

	// Clear any other authentication cookies that might exist
	cookiesToClear := []string{"auth_token", "session_id", "oidc_session"}
	for _, cookieName := range cookiesToClear {
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   auth.IsCookieSecure(), // Use the same function as other cookies
			SameSite: http.SameSiteLaxMode,
		})
	}

	// For JWT tokens, we can't invalidate them server-side without a blacklist
	// But we can provide information to the client about what to do
	response := map[string]interface{}{
		"message": "Successfully logged out",
		"instructions": map[string]interface{}{
			"client_action": "clear_token",
			"description":   "Remove the JWT token from client storage (localStorage, sessionStorage, etc.)",
		},
	}

	// If this was an OIDC session, we might want to redirect to OIDC logout endpoint
	if authType == "oidc" {
		// Optional: Get the OIDC logout URL for complete logout
		// Some providers support logout URLs that clear their session too
		response["oidc_logout_url"] = h.getOIDCLogoutURL()
		response["instructions"].(map[string]interface{})["oidc_logout"] = "Consider redirecting to oidc_logout_url to clear provider session"
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode logout response: %v", err)
		http.Error(w, "Failed to process logout", http.StatusInternalServerError)
		return
	}

	log.Printf("Logout successful for user %s", email)
}

// getOIDCLogoutURL constructs the OIDC provider logout URL (if supported)
func (h *AuthHandler) getOIDCLogoutURL() string {
	// This depends on your OIDC provider
	// Google: https://accounts.google.com/logout
	// Microsoft: https://login.microsoftonline.com/common/oauth2/logout
	// Auth0: https://YOUR_DOMAIN.auth0.com/v2/logout

	providerURL := getEnv("OIDC_PROVIDER_URL", "")
	if strings.Contains(providerURL, "accounts.google.com") {
		return "https://accounts.google.com/logout"
	} else if strings.Contains(providerURL, "login.microsoftonline.com") {
		return providerURL + "/oauth2/logout"
	} else if strings.Contains(providerURL, "auth0.com") {
		return providerURL + "/v2/logout"
	} else if strings.Contains(providerURL, "okta.com") {
		return providerURL + "/oauth2/v1/logout"
	} else if strings.Contains(providerURL, "keycloak.org") {
		return providerURL + "/realms/your-realm/protocol/openid-connect/logout"
	} else if strings.Contains(providerURL, "oryapis.com") {
		return providerURL + "/oauth2/sessions/logout"
	}

	return "" // No known logout URL
}

// UserInfo returns the current user's information
func (h *AuthHandler) UserInfo(w http.ResponseWriter, r *http.Request) {
	// Extract user info from context (set by middleware)
	userID := r.Context().Value("user_id")
	email := r.Context().Value("email")

	if userID == nil || email == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"user_id": userID,
		"email":   email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOIDCProvider returns the OIDC provider instance
func (h *AuthHandler) GetOIDCProvider() *auth.OIDCProvider {
	return h.oidcProvider
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
