package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/Neroframe/AuthService/internal/domain"
)

func (u *userUsecase) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	u.log.Info("in usecase getuserbyid")

	user, err := u.repo.FindByID(ctx, userID)
	if err != nil {
		u.log.Warn("FindByID", "id", userID)
		return nil, err
	}

	u.log.Info("return:", "user", user.Role)
	return user, nil
}

func (u *userUsecase) UpdateProfile(ctx context.Context, user *domain.User) (*domain.User, error) {
	updated, err := u.repo.Update(ctx, user, "email", "username", "phone", "updated_at")
	if err != nil {
		u.log.Warn("user update failed", "id", user.ID)
		return nil, err
	}
	if updated == nil {
		return nil, err
	}

	return updated, nil
}

func (u *userUsecase) ChangePassword(ctx context.Context, userID, oldPw, newPw string) error {
	if oldPw == newPw {
		u.log.Info("old and new password are same")
		return nil
	}

	user, err := u.repo.FindByID(ctx, userID)
	if err != nil {
		u.log.Warn("user not found", "id", userID)
		return err
	}

	_, err = u.repo.Update(ctx, user, "password")
	if err != nil {
		u.log.Warn("password update failed", "id", user.ID)
		return err
	}

	return nil
}
func (u *userUsecase) StartResetPassword(ctx context.Context, email string) error {
	// Fetch user
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// Generate code
	c := fmt.Sprintf("%06d", rand.Intn(1000000))
	code := domain.NewVerificationCode(user.ID, c, "reset_password", 10*time.Minute)

	// Cache it
	if err := u.cache.Set(ctx, code); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}

	// TODO email
	fmt.Printf("[RESET] Code sent to %s: %s\n", email, code.Code)

	return nil
}
func (u *userUsecase) ConfirmResetPassword(ctx context.Context, email, code, newPw string) error {
	// Verify the code
	if err := u.VerifyCode(ctx, email, code, "reset_password"); err != nil {
		return fmt.Errorf("verify code: %w", err)
	}

	// Fetch user
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// Hash new password
	hash, err := u.hasher.Hash(ctx, newPw)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Update password
	user.Password = hash
	_, err = u.repo.Update(ctx, user, "password")
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	// Remove the code from Redis
	_ = u.cache.Delete(ctx, user.ID)

	return nil
}
