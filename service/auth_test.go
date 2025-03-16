package service_test

import (
	"testing"
	"time"

	"github.com/vultisig/vultisigner/service"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	testCases := []struct {
		name        string
		secret      string
		shouldError bool
	}{
		{
			name:        "Valid secret",
			secret:      "secret-key-for-testing",
			shouldError: false,
		},
		{
			name:        "Empty secret",
			secret:      "",
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService := service.NewAuthService(tc.secret)
			token, err := authService.GenerateToken()

			if tc.shouldError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	secret := "test-secret-key"
	wrongSecret := "wrong-secret-key"

	testCases := []struct {
		name        string
		setupToken  func() string
		secret      string
		shouldError bool
	}{
		{
			name: "Valid token",
			setupToken: func() string {
				auth := service.NewAuthService(secret)
				token, _ := auth.GenerateToken()
				return token
			},
			secret:      secret,
			shouldError: false,
		},
		{
			name: "Expired token",
			setupToken: func() string {
				// Create a token that's already expired
				claims := &service.Claims{
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(secret))
				return tokenString
			},
			secret:      secret,
			shouldError: true,
		},
		{
			name: "Invalid signing method",
			setupToken: func() string {
				claims := &service.Claims{
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
				tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
				return tokenString
			},
			secret:      secret,
			shouldError: true,
		},
		{
			name: "Wrong secret",
			setupToken: func() string {
				auth := service.NewAuthService(secret)
				token, _ := auth.GenerateToken()
				return token
			},
			secret:      wrongSecret,
			shouldError: true,
		},
		{
			name: "Malformed token",
			setupToken: func() string {
				return "not-a-valid-token"
			},
			secret:      secret,
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenString := tc.setupToken()
			authService := service.NewAuthService(tc.secret)

			claims, err := authService.ValidateToken(tokenString)

			if tc.shouldError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.True(t, claims.ExpiresAt > time.Now().Unix())
			}
		})
	}
}

func TestRefreshToken(t *testing.T) {
	secret := "refresh-test-secret"

	testCases := []struct {
		name        string
		setupToken  func() string
		shouldError bool
	}{
		{
			name: "Valid token refresh",
			setupToken: func() string {
				auth := service.NewAuthService(secret)
				token, _ := auth.GenerateToken()
				time.Sleep(1 * time.Second)
				return token
			},
			shouldError: false,
		},
		{
			name: "Expired token refresh",
			setupToken: func() string {
				claims := &service.Claims{
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(secret))
				return tokenString
			},
			shouldError: true,
		},
		{
			name: "Invalid token refresh",
			setupToken: func() string {
				return "invalid-token-string"
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenString := tc.setupToken()
			authService := service.NewAuthService(secret)

			// For valid tokens, we need to guarantee a different ExpiresAt
			if !tc.shouldError {
				time.Sleep(1 * time.Second)
			}

			newToken, err := authService.RefreshToken(tokenString)

			if tc.shouldError {
				assert.Error(t, err)
				assert.Empty(t, newToken)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, newToken)
				assert.NotEqual(t, tokenString, newToken, "Refreshed token should be different from the original")

				claims, validationErr := authService.ValidateToken(newToken)
				assert.NoError(t, validationErr)
				assert.NotNil(t, claims)
			}
		})
	}
}

// TestAuthService tests the complete token lifecycle
func TestAuthService(t *testing.T) {
	secret := "integration-test-secret"
	authService := service.NewAuthService(secret)

	// Generate a token
	token, err := authService.GenerateToken()
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Validate the token
	claims, err := authService.ValidateToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)
	initialExpiry := claims.ExpiresAt

	// Wait to ensure different expiration time
	time.Sleep(1 * time.Second)

	// Refresh the token
	newToken, err := authService.RefreshToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, newToken)
	require.NotEqual(t, token, newToken, "Refreshed token should be different from the original")

	// Validate the new token
	newClaims, err := authService.ValidateToken(newToken)
	require.NoError(t, err)
	require.NotNil(t, newClaims)

	// Check that the new expiration time is later
	assert.True(t, newClaims.ExpiresAt > initialExpiry, "New token should have a later expiration time")
}
