package usecase

import (
	"context"

	"github.com/Neroframe/AuthService/internal/domain"
)

type userUsecase struct {
	repo      domain.UserRepository
	hasher    domain.PasswordHasher
	publisher domain.UserEventPublisher
	cache     domain.UserCache
}

func NewUserUsecase(r domain.UserRepository,
	hash domain.PasswordHasher,
	p domain.UserEventPublisher,
	cache domain.UserCache,
) UserUsecase {
	return &userUsecase{repo: r, hasher: hash, publisher: p, cache: cache}
}

func (u *userUsecase) Register(ctx context.Context, email, password string, role domain.Role) (*domain.User, error) {
	hashed, err := u.hasher.Hash(ctx, password)
	if err != nil {
		return nil, err
	}
	usr := &domain.User{Email: email, Password: hashed, Role: role}

	return usr, nil
}
