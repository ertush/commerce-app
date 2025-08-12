package tests

import (
	"os"
	"testing"

	"commerce-app/internal/auth"

	"github.com/google/uuid"
)

func TestOIDCConfig(t *testing.T) {
	config := &auth.OIDCConfig{
		ProviderURL:     "https://accounts.google.com",
		ClientID:        os.Getenv("OIDC_CLIENT_ID"),
		ClientSecret:    os.Getenv("OIDC_CLIENT_SECRET"),
		RedirectURL:     "http://localhost:8080/callback",
		Scopes:          []string{"openid", "profile", "email"},
		StateCookieName: "test_state",
	}

	if config.ProviderURL != "https://accounts.google.com" {
		t.Errorf("Expected ProviderURL to be https://accounts.google.com, got %s", config.ProviderURL)
	}

	if config.ClientID != os.Getenv("OIDC_CLIENT_ID") {
		t.Errorf("Expected ClientID to be test-client-id, got %s", config.ClientID)
	}

	if len(config.Scopes) != 3 {
		t.Errorf("Expected 3 scopes, got %d", len(config.Scopes))
	}
}

func TestRandomStringGeneration(t *testing.T) {
	length := 32
	randomString, err := auth.GenerateRandomString(length)
	if err != nil {
		t.Fatalf("Failed to generate random string: %v", err)
	}

	if len(randomString) == 0 {
		t.Error("Expected non-empty random string")
	}

	// Test that we get different strings
	randomString2, err := auth.GenerateRandomString(length)
	if err != nil {
		t.Fatalf("Failed to generate second random string: %v", err)
	}

	if randomString == randomString2 {
		t.Error("Expected different random strings")
	}
}

func TestOIDCClaims(t *testing.T) {
	testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	claims := &auth.OIDCClaims{
		UserID:   testUUID,
		Email:    "test@example.com",
		Name:     "Test User",
		Provider: "https://accounts.google.com",
		Issuer:   "https://accounts.google.com",
		Subject:  "550e8400-e29b-41d4-a716-446655440000",
		Audience: auth.AudienceClaimType([]string{"test-client-id"}),
		Expires:  1234567890,
		IssuedAt: 1234567800,
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Expected email to be test@example.com, got %s", claims.Email)
	}

	if claims.Name != "Test User" {
		t.Errorf("Expected name to be Test User, got %s", claims.Name)
	}

	if claims.Provider != "https://accounts.google.com" {
		t.Errorf("Expected provider to be https://accounts.google.com, got %s", claims.Provider)
	}
}

func TestUserInfo(t *testing.T) {
	testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	userInfo := &auth.UserInfo{
		ID:       testUUID,
		Email:    "test@example.com",
		Name:     "Test User",
		Picture:  "https://example.com/avatar.jpg",
		Provider: "https://accounts.google.com",
	}

	if userInfo.Email != "test@example.com" {
		t.Errorf("Expected email to be test@example.com, got %s", userInfo.Email)
	}

	if userInfo.Name != "Test User" {
		t.Errorf("Expected name to be Test User, got %s", userInfo.Name)
	}

	if userInfo.Picture != "https://example.com/avatar.jpg" {
		t.Errorf("Expected picture to be https://example.com/avatar.jpg, got %s", userInfo.Picture)
	}
}
