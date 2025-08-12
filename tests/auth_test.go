package tests

import (
	"net/http"
	"testing"

	"ecommerce-app/internal/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"

	token, err := auth.GenerateToken(userID, email)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"

	// Generate a token
	token, err := auth.GenerateToken(userID, email)
	assert.NoError(t, err)

	// Validate the token
	claims, err := auth.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	// Test with invalid token
	_, err := auth.ValidateToken("invalid-token")
	assert.Error(t, err)
}

func TestExtractTokenFromHeader(t *testing.T) {
	// Test valid header
	header := "Bearer valid-token"
	token, err := auth.ExtractTokenFromHeader(&http.Request{
		Header: map[string][]string{
			"Authorization": {header},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "valid-token", token)
}

func TestExtractTokenFromHeader_NoHeader(t *testing.T) {
	// Test missing header
	_, err := auth.ExtractTokenFromHeader(&http.Request{
		Header: map[string][]string{},
	})
	assert.Error(t, err)
}

func TestExtractTokenFromHeader_InvalidFormat(t *testing.T) {
	// Test invalid format
	_, err := auth.ExtractTokenFromHeader(&http.Request{
		Header: map[string][]string{
			"Authorization": {"InvalidFormat"},
		},
	})
	assert.Error(t, err)
}
