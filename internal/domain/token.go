package domain

import "context"

// ACCESS_TOKEN_EXPIRY_HOUR = 2
// REFRESH_TOKEN_EXPIRY_HOUR = 168
// ACCESS_TOKEN_SECRET=access_token_secret
// REFRESH_TOKEN_SECRET=refresh_token_secret

type JWTService interface {
	// Create a new signed access token
	Generate(userID string, role Role) (accessToken string, expiresAt int64, err error)

	// Parse and validate an access token
	Validate(ctx context.Context, token string) (*TokenPayload, error)
}

type TokenPayload struct {
	UserID    string
	Email     string
	Role      Role
	ExpiresAt int64 // Unix
}

// CreateAccessToken(user *domain.User, secret string, expiry int)
// CreateRefreshToken(user *domain.User, secret string, expiry int)
// IsAuthorized(requestToken string, secret string)
// ExtractIDFromToken(requestToken string, secret string)
