package domain

import (
	"context"
	"time"
)

// ACCESS_TOKEN_EXPIRY_HOUR = 2
// REFRESH_TOKEN_EXPIRY_HOUR = 168
// ACCESS_TOKEN_SECRET=access_token_secret
// REFRESH_TOKEN_SECRET=refresh_token_secret

// TODO: handle refresh token
type JWTService interface {
	// Create a new signed access token
	Generate(userID string, role Role) (accessToken string, issuedAt time.Time, expiresAt int64, err error)

	// Parse and validate an access token
	Validate(ctx context.Context, token string) (*TokenPayload, error)
}

type TokenPayload struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
	IssuedAt  time.Time `json:"issued_at"`  // UTC
	ExpiresAt int64     `json:"expires_at"` // Unix
}

// Email/SMS verification code
type VerificationCode struct {
	UserID    string
	Code      string
	ExpiresAt time.Time
	Purpose   string // "email_verification" or "reset_password"
}

// Password reset code ?
type PasswordResetCode struct {
	Code      string    `bson:"code"`
	UserID    string    `bson:"user_id"`
	ExpiresAt time.Time `bson:"expires_at"`
}

// Returns true if the code has expired
func (v *VerificationCode) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}

// Creates a verification code with TTL
func NewVerificationCode(userID, code, purpose string, ttl time.Duration) *VerificationCode {
	return &VerificationCode{
		UserID:    userID,
		Code:      code,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Creates a reset code with TTL
func NewPasswordResetCode(userID, code string, ttl time.Duration) *PasswordResetCode {
	return &PasswordResetCode{
		Code:      code,
		UserID:    userID,
		ExpiresAt: time.Now().Add(ttl),
	}
}
