package grpc

import (
	"context"
	"errors"

	"github.com/Neroframe/AuthService/internal/adapters/grpc/middleware"
	"github.com/Neroframe/AuthService/internal/domain"
	authpb "github.com/Neroframe/AuthService/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *AuthHandler) GetUserByID(ctx context.Context, req *authpb.GetUserByIDRequest) (*authpb.GetUserByIDResponse, error) {
	// Admin only
	val := ctx.Value(middleware.UserCtxKey)
	claims, ok := val.(*domain.TokenPayload)
	if !ok || claims == nil {
		return nil, status.Error(codes.Unauthenticated, "invalid context claims")
	}

	if claims.Role != domain.ADMIN {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	h.log.Info("before getuserbyid", "role", claims.Role)

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
	user := &domain.User{
		ID:       req.GetUserId(),
		Email:    req.GetEmail(),
		Username: req.GetUsername(),
		Phone:    req.GetPhone(),
	}

	usr, err := h.uc.UpdateProfile(ctx, user)
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

func (h *AuthHandler) SendVerificationCode(ctx context.Context, req *authpb.VerificationCodeRequest) (*authpb.VerificationCodeResponse, error) {
	// generate code
	// code := fmt.Sprintf("%06d", rand.Intn(1000000)) // 6-digit code

	// save to redis

	// send code to user email
	return nil, nil
}

func (h *AuthHandler) VerifyAccount(ctx context.Context, req *authpb.VerifyAccountRequest) (*authpb.VerifyAccountResponse, error) {
	// fetch code from redis:

	// does code exist
	// has it expired
	// codes match?

	// verify code or status.Error(codes.InvalidArgument, "invalid or expired code")
	return nil, nil
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

func (h *AuthHandler) ResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest) (*authpb.ResetPasswordResponse, error) {
	err := h.uc.SendVerificationCode(ctx, req.GetEmail(), "reset_password")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send reset code: %v", err)
	}

	return &authpb.ResetPasswordResponse{
		Success: true,
		Message: "Reset code sent to email",
	}, nil
}

func (h *AuthHandler) ConfirmResetPassword(ctx context.Context, req *authpb.ConfirmResetRequest) (*authpb.ConfirmResetResponse, error) {
	// Validate code and purpose
	if err := h.uc.VerifyCode(ctx, req.GetEmail(), req.GetCode(), "reset_password"); err != nil {
		if errors.Is(err, domain.ErrCodeInvalid) || errors.Is(err, domain.ErrCodeExpired) {
			return nil, status.Error(codes.InvalidArgument, "invalid or expired code")
		}
		return nil, status.Errorf(codes.Internal, "failed to verify code: %v", err)
	}

	// Update password
	err := h.uc.ConfirmResetPassword(ctx, req.GetEmail(), req.GetCode(), req.GetNewPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reset password: %v", err)
	}

	return &authpb.ConfirmResetResponse{
		Success: true,
		Message: "Password has been reset",
	}, nil
}
