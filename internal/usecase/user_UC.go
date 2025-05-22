package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/Neroframe/AuthService/internal/adapters/grpc/middleware"
	"github.com/Neroframe/AuthService/internal/domain"
	"github.com/Neroframe/AuthService/internal/repository"
)

func (u *userUsecase) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	claims := ctx.Value(middleware.UserCtxKey).(*domain.TokenPayload)

	target, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("UpdateProfile fetch: %w", err)
	}

	// Check for role permission
	switch claims.Role {
	case domain.ADMIN:
		// allowed
	case domain.TEACHER: // only itself and students
		if target.Role == domain.ADMIN || (target.Role == domain.TEACHER && target.ID != userID) {
			return nil, domain.ErrPermissionDenied
		}
	case domain.STUDENT: // only itself
		if target.Role == domain.ADMIN || target.Role == domain.TEACHER || (target.Role == domain.STUDENT && target.ID != userID) {
			return nil, domain.ErrPermissionDenied
		}
	default:
		return nil, domain.ErrPermissionDenied
	}

	usr, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("GetUserByID: %w", err)
	}

	return usr, nil
}

func (u *userUsecase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	usr, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("GetUserByEmail: %w", err)
	}

	return usr, nil
}

func (u *userUsecase) UpdateProfile(ctx context.Context, p domain.UpdateUserProfileParams) (*domain.User, error) {
	claims := ctx.Value(middleware.UserCtxKey).(*domain.TokenPayload)

	target, err := u.repo.GetByID(ctx, p.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("UpdateProfile fetch: %w", err)
	}

	// Check for role permission
	switch claims.Role {
	case domain.ADMIN:
		// allowed
	case domain.TEACHER: // only itself and students
		if target.Role == domain.ADMIN || (target.Role == domain.TEACHER && target.ID != p.ID) {
			return nil, domain.ErrPermissionDenied
		}
	case domain.STUDENT: // only itself
		if target.Role == domain.ADMIN || target.Role == domain.TEACHER || (target.Role == domain.STUDENT && target.ID != p.ID) {
			return nil, domain.ErrPermissionDenied
		}
	default:
		return nil, domain.ErrPermissionDenied
	}

	// Set new values
	target.Email = p.Email
	target.Username = p.Username
	target.Phone = p.Phone

	// Update user profile
	updated, err := u.repo.Update(ctx, target, "email", "username", "phone", "updated_at")
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("UpdateProfile: %w", err)
	}

	return updated, nil
}

func (u *userUsecase) VerifyAccount(ctx context.Context, userID string) error {
	// Get user
	usr, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("VerifyAccount: %w", err)
	}

	// Update user as verified
	usr.Verified = true
	if _, err := u.repo.Update(ctx, usr, "verified", "updated_at"); err != nil {
		return fmt.Errorf("VerifyAccount Update: %w", err)
	}

	return nil
}

func (u *userUsecase) ChangePassword(ctx context.Context, userID, oldPw, newPw string) error {
	hashed, err := u.hasher.Hash(ctx, newPw)
	if err != nil {
		return fmt.Errorf("ChangePassword Hash: %w", err)
	}

	if oldPw == hashed {
		return domain.ErrPasswordUnchanged
	}

	// Get user
	usr, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("ChangePassword FindByID: %w", err)
	}

	// Update password
	usr.Password = hashed
	_, err = u.repo.Update(ctx, usr, "password", "updated_at")
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("ChangePassword Update: %w", err)
	}

	return nil
}


