package repository

import (
	"context"
	"errors"

	"github.com/Neroframe/AuthService/internal/domain"
)

var (
	ErrNotFound         = errors.New("user not found")
	ErrEmailAlreadyUsed = errors.New("email already used")
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, u *domain.User, fields ...string) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}
