package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Neroframe/AuthService/internal/domain"
	"github.com/Neroframe/AuthService/internal/repository"
	"github.com/Neroframe/AuthService/pkg/logger"
)

type userUsecase struct {
	log       *logger.Logger
	repo      repository.UserRepository
	hasher    domain.PasswordHasher
	publisher domain.UserEventPublisher
	cache     domain.UserCache
	jwt       domain.JWTService
}

func NewUserUsecase(
	r repository.UserRepository,
	h domain.PasswordHasher,
	p domain.UserEventPublisher,
	c domain.UserCache,
	log *logger.Logger,
	jwtSvc domain.JWTService,
) UserUsecase {
	return &userUsecase{repo: r, hasher: h, publisher: p, cache: c, log: log, jwt: jwtSvc}
}

func (u *userUsecase) Register(ctx context.Context, email, password string, role domain.Role) (*domain.User, error) {
	// hash password
	hashed, err := u.hasher.Hash(ctx, password)
	if err != nil {
		return nil, fmt.Errorf("Register Hash: %w", err)
	}

	usr := &domain.User{Email: email, Password: hashed, Role: role}

	// Save to DB
	if err := u.repo.Create(ctx, usr); err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyUsed) {
			return nil, domain.ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("Register Create: %w", err)
	}

	// NATS publish
	event := &domain.UserRegisteredEvent{
		UserID:    usr.ID,
		Email:     usr.Email,
		Role:      usr.Role,
		CreatedAt: time.Now().UTC(),
	}
	if err := u.publisher.PublishUserRegistered(ctx, event); err != nil {
		return nil, fmt.Errorf("Register PublishEvent (registered): %w", err) // TODO: handle or fire-and-forget
	}

	return usr, nil
}

func (u *userUsecase) Login(ctx context.Context, email, password string) (accessToken string, payload *domain.TokenPayload, err error) {
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", nil, domain.ErrUserNotFound
		}
		return "", nil, fmt.Errorf("Login FindByEmail: %w", err)
	}

	ok := u.hasher.Verify(ctx, user.Password, password)
	if !ok {
		return "", nil, domain.ErrInvalidCredentials
	}

	token, iat, exp, err := u.jwt.Generate(user.ID, user.Role)
	if err != nil {
		return "", nil, fmt.Errorf("Login jwt.Generate: %w", err)
	}

	return token, &domain.TokenPayload{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		IssuedAt:  iat,
		ExpiresAt: exp,
	}, nil
}

func (u *userUsecase) ValidateToken(ctx context.Context, jwt string) (*domain.TokenPayload, error) {
	payload, err := u.jwt.Validate(ctx, jwt)
	if err != nil {
		return nil, fmt.Errorf("ValidateToken jwt.Validate: %w", err)
	}

	return payload, nil
}

func (u *userUsecase) SendVerificationCode(ctx context.Context, email, purpose string) error {
	// Find user
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("SendVerificationCode FindByEmail: %w", err)
	}

	// Generate struct with code
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	verificationCode := domain.NewVerificationCode(user.ID, code, purpose, 5*time.Minute)

	// Store in redis
	if err := u.cache.Set(ctx, verificationCode); err != nil {
		return fmt.Errorf("SendVerificationCode cache.Set: %w", err)
	}

	// TODO: send email
	fmt.Printf("[DEBUG] Sent verification code to %s: %s\n", user.Email, verificationCode.Code)

	return nil
}

func (u *userUsecase) VerifyCode(ctx context.Context, email string, code string, purpose string) error {
	// Find user
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("VerifyCode FindByEmail: %w", err)
	}

	// Find code
	cachedCode, err := u.cache.Get(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("VerifyCode cache.Get: %w", err)
	}

	// Validate
	if cachedCode.Code != code {
		return domain.ErrCodeInvalid
	}

	if time.Now().After(cachedCode.ExpiresAt) {
		return domain.ErrCodeExpired
	}

	if cachedCode.Purpose != purpose {
		return domain.ErrInvalidPurpose
	}

	// Purpose specific actions
	switch purpose {
	case domain.PurposeEmailVerification:
		// Update user as verified
		user.Verified = true
		if _, err := u.repo.Update(ctx, user, "verified"); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return domain.ErrUserNotFound
			}
			return fmt.Errorf("VerifyCode email Update: %w", err)
		}
	case domain.PurposeResetPassword:
		// wait for ConfirmResetPassword to set new password
	default:
		return domain.ErrInvalidPurpose
	}

	// Remove from cache
	err = u.cache.Delete(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("VerifyCode cache.Delete: %w", err) // TODO: handle or fire-and-forget
	}

	return nil
}
