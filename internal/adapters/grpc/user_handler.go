package grpc

import (
	"context"
	"errors"

	"github.com/Neroframe/AuthService/internal/domain"
	authpb "github.com/Neroframe/AuthService/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *AuthHandler) GetUserByID(ctx context.Context, req *authpb.GetUserByIDRequest) (*authpb.GetUserByIDResponse, error) {
	// Get user
	usr, err := h.uc.GetUserByID(ctx, req.UserId)
	if err != nil {
		h.log.Error("failed to find by ID", "err", err)
		return nil, status.Error(codes.NotFound, "failed to find by ID")
	}

	return &authpb.GetUserByIDResponse{
		Success: true,
		Message: "user found",
		User: &authpb.User{
			UserId:   usr.ID,
			Email:    usr.Email,
			Username: usr.Username,
			Password: usr.Password,
			Role:     convertRole(usr.Role), // ? panics if no user found
			Phone:    usr.Phone,
		},
	}, nil
}

func (h *AuthHandler) UpdateUserProfile(ctx context.Context, req *authpb.UpdateUserRequest) (*authpb.UpdateUserResponse, error) {
	params := domain.UpdateUserProfileParams{
		ID:       req.UserId,
		Email:    req.Email,
		Username: req.Username,
		Phone:    req.Phone,
	}

	usr, err := h.uc.UpdateProfile(ctx, params)
	if err != nil {
		h.log.Error("failed to update user profile", "err", err)
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return &authpb.UpdateUserResponse{
		Success: true,
		Message: "user updated",
		User: &authpb.User{
			UserId:   usr.ID,
			Email:    usr.Email,
			Username: usr.Username,
			Password: usr.Password,
			Role:     convertRole(usr.Role),
			Phone:    usr.Phone,
		},
	}, nil
}

func (h *AuthHandler) ChangePassword(ctx context.Context, req *authpb.ChangePasswordRequest) (*authpb.ChangePasswordResponse, error) {
	err := h.uc.ChangePassword(ctx, req.UserId, req.OldPassword, req.NewPassword)
	if err != nil {
		h.log.Error("failed to change password", "err", err)
		return nil, status.Error(codes.Internal, "failed to change password")
	}

	return &authpb.ChangePasswordResponse{
		Success: true,
		Message: "password changed",
	}, nil
}

func (h *AuthHandler) SendVerificationCode(ctx context.Context, req *authpb.VerificationCodeRequest) (*authpb.VerificationCodeResponse, error) {
	// Send verification code to email
	err := h.uc.SendVerificationCode(ctx, req.Email, domain.PurposeEmailVerification)
	if err != nil {
		h.log.Error("failed to send verification code", "err", err)
		return nil, status.Error(codes.Internal, "failed to send verification code")
	}

	return &authpb.VerificationCodeResponse{
		Success: true,
		Message: "Verification code sent to email",
	}, nil
}

func (h *AuthHandler) VerifyAccount(ctx context.Context, req *authpb.VerifyAccountRequest) (*authpb.VerifyAccountResponse, error) {
	// Validate code and purpose
	if err := h.uc.VerifyCode(ctx, req.Email, req.Code, domain.PurposeEmailVerification); err != nil {
		if errors.Is(err, domain.ErrCodeExpired) {
			return nil, status.Error(codes.InvalidArgument, "expired code")
		}
		if errors.Is(err, domain.ErrCodeInvalid) {
			return nil, status.Error(codes.InvalidArgument, "invalid code")
		}
		h.log.Error("Failed to verify code", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to verify code: %v", err)
	}

	// Fetch user
	usr, err := h.uc.GetUserByEmail(ctx, req.Email)
	if err != nil {
		h.log.Error("Failed to get user", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch user detail: %v", err)
	}

	if err := h.uc.VerifyAccount(ctx, usr.ID); err != nil {
		h.log.Error("Failed to verify user", "err", err)
		return nil, status.Error(codes.Internal, "failed to verify user")
	}

	return nil, nil
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest) (*authpb.ResetPasswordResponse, error) {
	// Send reset code to email
	err := h.uc.SendVerificationCode(ctx, req.GetEmail(), domain.PurposeResetPassword)
	if err != nil {
		h.log.Error("failed to send reset code", "err", err)
		return nil, status.Error(codes.Internal, "failed to send reset code")
	}

	return &authpb.ResetPasswordResponse{
		Success: true,
		Message: "Reset code sent to email",
	}, nil
}

func (h *AuthHandler) ConfirmResetPassword(ctx context.Context, req *authpb.ConfirmResetRequest) (*authpb.ConfirmResetResponse, error) {
	// Validate code and purpose
	if err := h.uc.VerifyCode(ctx, req.Email, req.Code, domain.PurposeResetPassword); err != nil {
		if errors.Is(err, domain.ErrCodeExpired) {
			return nil, status.Error(codes.InvalidArgument, "expired code")
		}
		if errors.Is(err, domain.ErrCodeInvalid) {
			return nil, status.Error(codes.InvalidArgument, "invalid code")
		}
		h.log.Error("Failed to verify code", "err", err)
		return nil, status.Error(codes.Internal, "failed to verify code")
	}

	// Fetch user
	usr, err := h.uc.GetUserByEmail(ctx, req.Email)
	if err != nil {
		h.log.Error("Failed to get user", "err", err)
		return nil, status.Error(codes.Internal, "failed to fetch user detail")
	}

	// Change password
	if err := h.uc.ChangePassword(ctx, usr.ID, usr.Password, req.NewPassword); err != nil {
		h.log.Error("Failed to change password", "err", err)
		return nil, status.Error(codes.Internal, "failed to reset password")
	}

	return &authpb.ConfirmResetResponse{
		Success: true,
		Message: "Password has been reset",
	}, nil
}
