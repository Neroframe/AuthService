package grpc

import (
	"context"
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
	jwt domain.JWTService
}

func NewHandler(uc usecase.UserUsecase, log *logger.Logger, jwt domain.JWTService) *AuthHandler {
	return &AuthHandler{uc: uc, log: log, jwt: jwt}
}

func (h *AuthHandler) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	h.log.Info("Login called", "email", req.Email, "password", req.Password)

	token, payload, err := h.uc.Login(ctx, req.Email, req.Password)
	if err != nil {
		h.log.Warn("Login failed", "err", err)
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	h.log.Info("Login successful", "user_id", payload.UserID)

	return &authpb.LoginResponse{
		AccessToken:  token,
		RefreshToken: "", // TODO
		ExpiresAt:    payload.ExpiresAt,
		TokenType:    "Bearer",
	}, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	h.log.Info("Register called", "email", req.Email)

	// Convert into domain.Role
	role, err := convertProtoRole(req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	usr, err := h.uc.Register(ctx, req.Email, req.Password, role)
	if err != nil {
		h.log.Error("Register usecase failed", "err", err)
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	h.log.Info("User Registered", "id", usr.ID)

	return &authpb.RegisterResponse{
		Success:     true,
		Message:     "User Registered",
		AccessToken: "", // TODO ?
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	h.log.Info("ValidateToken called", "token", req.Jwt)

	payload, err := h.uc.ValidateToken(ctx, req.Jwt)
	if err != nil {
		h.log.Error("ValidateToken usecase failed", "err", err)
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	h.log.Info("token valid", "user_id", payload.UserID)

	return &authpb.ValidateTokenResponse{
		Valid:     true,
		UserId:    payload.UserID,
		Role:      convertRole(payload.Role), // map domain.Role to authpb.Role
		ExpiresAt: payload.ExpiresAt,
	}, nil
}

func (h *AuthHandler) GetUserByID(ctx context.Context, req *authpb.GetUserByIDRequest) (*authpb.GetUserByIDResponse, error) {
	h.log.Info("GetUserByID called", "id", req.UserId)

	usr, err := h.uc.GetUserByID(ctx, req.UserId)
	if err != nil {
		h.log.Error("failed to find by ID", "err", err)
		return nil, status.Error(codes.NotFound, "failed to find by ID")
	}

	return &authpb.GetUserByIDResponse{
		UserID: usr.ID,
		Email:  usr.Email,
		Role:   string(usr.Role),
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
