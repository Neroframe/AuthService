package domain

import (
	"context"
)

type User struct {
	ID       string
	Email    string
	Password string
	Role     Role
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
	// Create(ctx context.Context, u *User) error
	// FindByEmail(ctx context.Context, email string) (*User, error)
	// FindByID(ctx context.Context, id string) (*User, error)
	// Update(ctx context.Context, u *User) error
	// Delete(ctx context.Context, id string) error
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
