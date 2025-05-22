package domain

import (
	"context"
	"time"
)

type TokenPayload struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
	IssuedAt  time.Time `json:"issued_at"`  // UTC
	ExpiresAt int64     `json:"expires_at"` // Unix
}

type VerificationCode struct {
	UserID    string
	Code      string
	ExpiresAt time.Time
	Purpose   string // "email_verification" or "reset_password"
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

// TODO: handle refresh token
type JWTService interface {
	Generate(userID string, role Role) (accessToken string, issuedAt time.Time, expiresAt int64, err error) // generate access_token
	Validate(ctx context.Context, token string) (*TokenPayload, error)                                      // validate access_token
}

type CodeCache interface {
	Set(ctx context.Context, code *VerificationCode) error
	Get(ctx context.Context, userID string) (*VerificationCode, error)
	Delete(ctx context.Context, userID string) error
}
