package usecase

import (
	"context"

	"github.com/Neroframe/AuthService/internal/domain"
)

type UserUsecase interface {
	// Auth
	Register(ctx context.Context, email, password string, role domain.Role) (*domain.User, error)
	Login(ctx context.Context, email, password string) (accessToken string, payload *domain.TokenPayload, err error)

	// Token validation
	ValidateToken(ctx context.Context, jwt string) (*domain.TokenPayload, error)

	// Account verification
	SendVerificationCode(ctx context.Context, email, purpose string) error
	VerifyCode(ctx context.Context, email string, code string, purpose string) error

	// User management
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateProfile(ctx context.Context, p domain.UpdateUserProfileParams) (*domain.User, error)
	ChangePassword(ctx context.Context, userID, oldPw, newPw string) error
	VerifyAccount(ctx context.Context, userID string) error
}
