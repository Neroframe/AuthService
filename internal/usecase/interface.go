package usecase

import (
	"context"

	"github.com/Neroframe/AuthService/internal/domain"
)

type UserUsecase interface {
	Register(ctx context.Context, email, password string, role domain.Role) (*domain.User, error)
	Login(ctx context.Context, email, password string) (accessToken string, payload *domain.TokenPayload, err error)

	ValidateToken(ctx context.Context, jwt string) (*domain.TokenPayload, error)
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
}
