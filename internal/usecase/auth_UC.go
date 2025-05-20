package usecase

import (
	"context"
	"fmt"
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

	// Set redis cacher
	// if err := u.cache.Set(ctx, u); err != nil {
	// 	u.log.Warn("cache set failed", "err", err)
	// }

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

	token, exp, err := u.jwt.Generate(user.ID, user.Role)
	if err != nil {
		u.log.Warn("jwt.generate failed", "err", err)
		return "", nil, fmt.Errorf("failed to generate token")
	}

	u.log.Info("Login success", "token", token)

	return token, &domain.TokenPayload{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		ExpiresAt: exp,
	}, nil
}

func (u *userUsecase) ValidateToken(ctx context.Context, jwt string) (*domain.TokenPayload, error) {
	payload, err := u.jwt.Validate(ctx, jwt)
	if err != nil {
		u.log.Warn("invalid token", "token", jwt)
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// MongoDB or redis
	// ok, err := u.repo.Exists(ctx, payload.UserID, jwt)
	// if err != nil {
	// 	return nil, fmt.Errorf("session check failed: %w", err)
	// }
	// if !ok {
	// 	return nil, fmt.Errorf("token revoked or session expired")
	// }

	return payload, nil
}

func (u *userUsecase) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	user, err := u.repo.FindByID(ctx, userID)
	if err != nil {
		u.log.Warn("user not found", "id", userID)
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
