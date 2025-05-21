package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Neroframe/AuthService/internal/domain"
	"github.com/Neroframe/AuthService/internal/repository"
)

func (u *userUsecase) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	user, err := u.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("GetUserByID: %w", err)
	}

	return user, nil
}

func (u *userUsecase) UpdateProfile(ctx context.Context, user *domain.User) (*domain.User, error) {
	updated, err := u.repo.Update(ctx, user, "email", "username", "phone", "updated_at")
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("UpdateProfile: %w", err)
	}

	return updated, nil
}

func (u *userUsecase) ChangePassword(ctx context.Context, userID, oldPw, newPw string) error {
	if oldPw == newPw {
		return domain.ErrPasswordUnchanged
	}

	user, err := u.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("ChangePassword FindByID: %w", err)
	}

	_, err = u.repo.Update(ctx, user, "password")
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("ChangePassword Update: %w", err)
	}

	return nil
}
func (u *userUsecase) StartResetPassword(ctx context.Context, email string) error {
	// Fetch user
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("StartResetPassword FindByEmail: %w", err)
	}

	// Generate code
	c := fmt.Sprintf("%06d", rand.Intn(1000000))
	code := domain.NewVerificationCode(user.ID, c, "reset_password", 10*time.Minute)

	// Cache it
	if err := u.cache.Set(ctx, code); err != nil {
		return fmt.Errorf("StartResetPassword cache.Set: %w", err)
	}

	// TODO: email
	fmt.Printf("[RESET] Code sent to %s: %s\n", email, code.Code)

	return nil
}
func (u *userUsecase) ConfirmResetPassword(ctx context.Context, email, code, newPw string) error {
	// Verify the code
	if err := u.VerifyCode(ctx, email, code, "reset_password"); err != nil {
		return fmt.Errorf("ConfirmResetPassword VerifyCode: %w", err)
	}

	// Fetch user
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("ConfirmResetPassword FindByEmail: %w", err)
	}

	// Hash new password
	hash, err := u.hasher.Hash(ctx, newPw)
	if err != nil {
		return fmt.Errorf("ConfirmResetPassword Hash: %w", err)
	}

	// Update password
	user.Password = hash
	_, err = u.repo.Update(ctx, user, "password")
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("ConfirmResetPassword Update: %w", err)
	}

	// Remove the code from Redis
	_ = u.cache.Delete(ctx, user.ID)

	return nil
}
