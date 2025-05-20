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
	ErrVerificationCodeExpired  = errors.New("verification code expired")
	ErrVerificationCodeInvalid  = errors.New("invalid verification code")
	ErrPasswordResetCodeExpired = errors.New("password reset code expired")
	ErrPasswordResetCodeInvalid = errors.New("invalid password reset code")
)

type User struct {
	ID        string    `bson:"_id"`
	Email     string    `bson:"email"`
	Username  string    `bson:"username"`
	Password  string    `bson:"password"`
	Role      Role      `bson:"role"`
	Phone     string    `bson:"phone"`
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

// NewUser constructs a new User, ensuring the role is valid
func NewUser(id, email, username, passwordHash string, role Role, phone string) (*User, error) {
	if !role.IsValid() {
		return nil, ErrInvalidCredentials
	}
	return &User{
		ID:        id,
		Email:     email,
		Username:  username,
		Password:  passwordHash,
		Role:      role,
		Phone:     phone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// UpdateProfile updates mutable user fields
func (u *User) UpdateProfile(email, username, phone string) {
	if email != "" {
		u.Email = email
	}
	if username != "" {
		u.Username = username
	}
	if phone != "" {
		u.Phone = phone
	}
	u.UpdatedAt = time.Now()
}

// ChangePassword sets a new password hash
func (u *User) ChangePassword(newHash string) {
	u.Password = newHash
	u.UpdatedAt = time.Now()
}

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, u *User, fields ...string) (*User, error)
	Delete(ctx context.Context, id string) error
}

type UserCache interface {
	GetByID(ctx context.Context, userID string) (*User, error)
	// Set(ctx context.Context, user *User) error
	// SetMany(ctx context.Context, users []*User) error
	// Delete(ctx context.Context, userID string) error
}

type PasswordHasher interface {
	Hash(ctx context.Context, plain string) (string, error)
	Verify(ctx context.Context, hashed, plain string) bool
}
