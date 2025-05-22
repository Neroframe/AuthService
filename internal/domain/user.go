package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

//

// Purposes for code verification
const (
	PurposeEmailVerification = "email_verification"
	PurposeResetPassword     = "reset_password"
)

// Domain errors
var (
	// User errors
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrPasswordUnchanged     = errors.New("password unchanged")
	ErrInvalidCredentials    = errors.New("invalid credentials")

	// Token errors
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")

	// Code errors
	ErrCodeInvalid              = errors.New("verification code invalid")
	ErrCodeExpired              = errors.New("verification code expired")
	ErrPasswordResetCodeExpired = errors.New("password reset code expired")
	ErrPasswordResetCodeInvalid = errors.New("invalid password reset code")
	ErrInvalidPurpose           = errors.New("invalid code purpose")
)

type User struct {
	ID        string    `bson:"_id"`
	Email     string    `bson:"email"`
	Username  string    `bson:"username"`
	Password  string    `bson:"password"`
	Role      Role      `bson:"role"`
	Phone     string    `bson:"phone"`
	Verified  bool      `bson:"verified"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

type Role string

const (
	UNSPECIFIED Role = ""
	ADMIN       Role = "admin"
	TEACHER     Role = "teacher"
	STUDENT     Role = "student"
)

func NewUser(email, hashedPwd string, role Role) *User {
	return &User{
		ID:        uuid.NewString(),
		Email:     email,
		Password:  hashedPwd,
		Role:      role,
		Verified:  false,
		CreatedAt: time.Now().UTC(),
	}
}

func (u *User) Verify() {
	if !u.Verified {
		u.Verified = true
		u.UpdatedAt = time.Now().UTC()
	}
}

func (r Role) IsValid() bool {
	switch r {
	case ADMIN, TEACHER, STUDENT:
		return true
	default:
		return false
	}
}

type UserCache interface {
	Set(ctx context.Context, code *VerificationCode) error
	Get(ctx context.Context, userID string) (*VerificationCode, error)
	Delete(ctx context.Context, userID string) error
}

type PasswordHasher interface {
	Hash(ctx context.Context, plain string) (string, error)
	Verify(ctx context.Context, hashed, plain string) bool
}

type EmailSender interface {
	Send(to, subject, htmlBody string) error
}
