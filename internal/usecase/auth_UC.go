package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/Neroframe/AuthService/internal/domain"
	"github.com/Neroframe/AuthService/pkg/logger"
)

type userUsecase struct {
	log       *logger.Logger
	repo      domain.UserRepository
	hasher    domain.PasswordHasher
	publisher domain.UserEventPublisher
	cache     domain.UserCache
	jwt       domain.JWTService
}

func NewUserUsecase(r domain.UserRepository,
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
		u.log.Error("failed to hash password", "err", err)
		return nil, err
	}

	usr := &domain.User{Email: email, Password: hashed, Role: role}

	// Save to DB
	if err := u.repo.Create(ctx, usr); err != nil {
		u.log.Error("failed to create user", "email", email, "err", err)
		return nil, err
	}

	// NATS publish
	event := &domain.UserRegisteredEvent{
		UserID:    usr.ID,
		Email:     usr.Email,
		Role:      usr.Role,
		CreatedAt: time.Now().UTC(),
	}
	if err := u.publisher.PublishUserRegistered(ctx, event); err != nil {
		u.log.Warn("failed to publish user_registered event", "user_id", usr.ID, "err", err)
	}
	u.log.Info("user registered", "user_id", usr.ID, "email", usr.Email)

	return usr, nil
}

func (u *userUsecase) Login(ctx context.Context, email, password string) (accessToken string, payload *domain.TokenPayload, err error) {
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		u.log.Warn("user not found", "err", err)
		return "", nil, fmt.Errorf("invalid credentials")
	}

	u.log.Info("user found", "email", user.Email, "hash", user.Password, "plain", password)

	ok := u.hasher.Verify(ctx, user.Password, password)
	if !ok {
		u.log.Warn("invalid password", "email", email)
		return "", nil, fmt.Errorf("invalid credentials")
	}

	token, iat, exp, err := u.jwt.Generate(user.ID, user.Role)
	if err != nil {
		u.log.Warn("jwt.generate failed", "err", err)
		return "", nil, fmt.Errorf("failed to generate token")
	}

	u.log.Info("Login success", "token", token)

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
		u.log.Warn("invalid token", "token", jwt)
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return payload, nil
}

func (u *userUsecase) SendVerificationCode(ctx context.Context, email, purpose string) error {
	// Find user
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	// Generate struct with code
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	verificationCode := domain.NewVerificationCode(user.ID, code, purpose, 5*time.Minute)

	// Store in redis
	if err := u.cache.Set(ctx, verificationCode); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	// TODO: send email
	fmt.Printf("[DEBUG] Sent verification code to %s: %s\n", user.Email, verificationCode.Code)

	return nil
}

func (u *userUsecase) VerifyCode(ctx context.Context, email string, code string, purpose string) error {
	// Find user
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	// Find code
	cachedCode, err := u.cache.Get(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("redis get: %w", err)
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
	case "email_verification":
		// Update user as verified
		user.Verified = true
		if _, err := u.repo.Update(ctx, user, "verified"); err != nil {
			return fmt.Errorf("update user: %w", err)
		}
	case "reset_password":
		// wait for ConfirmResetPassword to set new password
	default:
		return domain.ErrInvalidPurpose
	}

	// Remove from cache
	err = u.cache.Delete(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("redis del: %w", err)
	}

	return nil
}
