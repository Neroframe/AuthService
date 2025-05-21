package domain

import (
	"context"
	"errors"
	"time"
)

// Domain errors
var (
	ErrEmailAlreadyExists       = errors.New("email already exists")
	ErrUsernameAlreadyExists    = errors.New("username already exists")
	ErrUserNotFound             = errors.New("user not found")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrInvalidToken             = errors.New("invalid token")
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

func (r Role) IsValid() bool {
	switch r {
	case ADMIN, TEACHER, STUDENT:
		return true
	default:
		return false
	}
}

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, u *User, fields ...string) (*User, error)
	Delete(ctx context.Context, id string) error
}

type UserCache interface {
	Set(ctx context.Context, code *VerificationCode) error
	Get(ctx context.Context, userID string) (*VerificationCode, error)
	Delete(ctx context.Context, userID string) error

	GetByID(ctx context.Context, userID string) (*User, error)
}

type PasswordHasher interface {
	Hash(ctx context.Context, plain string) (string, error)
	Verify(ctx context.Context, hashed, plain string) bool
}
