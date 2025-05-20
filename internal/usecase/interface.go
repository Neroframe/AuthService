package usecase

import (
	"context"

	"github.com/Neroframe/AuthService/internal/domain"
)

type UserUsecase interface {
	Register(ctx context.Context, email, password string, role domain.Role) (*domain.User, error)
	// Login(ctx context.Context, email, password string) (accessToken)
	
}
