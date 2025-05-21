package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/Neroframe/AuthService/internal/domain"
	"github.com/Neroframe/AuthService/internal/usecase"
	"github.com/Neroframe/AuthService/pkg/logger"
	authpb "github.com/Neroframe/AuthService/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	uc  usecase.UserUsecase
	log *logger.Logger
}

func NewHandler(uc usecase.UserUsecase, log *logger.Logger) *AuthHandler {
	return &AuthHandler{uc: uc, log: log}
}

func (h *AuthHandler) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "Email and password required")
	}

	// Validate and convert into domain.Role
	role, err := convertProtoRole(req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}
	h.log.Info("role converted", "email", req.Email, "role", role)

	usr, err := h.uc.Register(ctx, req.Email, req.Password, role)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "email already in use")
		}
		h.log.Error("Register failed", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	h.log.Info("Register successful", "user_id", usr.ID, "email", usr.Email)

	return &authpb.RegisterResponse{
		Success:     true,
		Message:     "User Registered",
		AccessToken: "", // TODO ?
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	token, payload, err := h.uc.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		h.log.Error("Login failed", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	h.log.Info("Login successful", "user_id", payload.UserID)

	return &authpb.LoginResponse{
		AccessToken:  token,
		RefreshToken: "", // TODO
		ExpiresAt:    payload.ExpiresAt,
		TokenType:    "Bearer",
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	payload, err := h.uc.ValidateToken(ctx, req.Jwt)
	if err != nil {
		h.log.Error("ValidateToken usecase failed", "err", err)
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	h.log.Info("Token valid", "user_id", payload.UserID)

	return &authpb.ValidateTokenResponse{
		Valid:     true,
		UserId:    payload.UserID,
		Role:      convertRole(payload.Role), // map domain.Role to authpb.Role
		ExpiresAt: payload.ExpiresAt,
	}, nil
}

func convertRole(r domain.Role) authpb.Role {
	switch r {
	case domain.ADMIN:
		return authpb.Role_ADMIN
	case domain.TEACHER:
		return authpb.Role_TEACHER
	case domain.STUDENT:
		return authpb.Role_STUDENT
	default:
		return authpb.Role_UNSPECIFIED
	}
}

func convertProtoRole(r authpb.Role) (domain.Role, error) {
	switch r {
	case authpb.Role_ADMIN:
		return domain.ADMIN, nil
	case authpb.Role_TEACHER:
		return domain.TEACHER, nil
	case authpb.Role_STUDENT:
		return domain.STUDENT, nil
	default:
		return domain.UNSPECIFIED, fmt.Errorf("unknown proto Role: %v", r)
	}
}
