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
	UpdateProfile(ctx context.Context, updated *domain.User) (*domain.User, error)

	SendVerificationCode(ctx context.Context, email string) error
	VerifyAccount(ctx context.Context, email string, code string) error

	ChangePassword(ctx context.Context, userID, oldPw, newPw string) error
	StartResetPassword(ctx context.Context, email string) error
	ConfirmResetPassword(ctx context.Context, email, code, newPw string) error
}
