package usecase

import (
	"context"

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

	return nil
}
func (u *userUsecase) ConfirmResetPassword(ctx context.Context, email, code, newPw string) error {

	return nil
}
