package grpc

import (
	"context"

	"github.com/Neroframe/AuthService/internal/usecase"
	"github.com/Neroframe/AuthService/pkg/logger"
	authpb "github.com/Neroframe/AuthService/proto"
)

type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	uc  usecase.UserUsecase
	log *logger.Logger
}

func NewHandler(uc usecase.UserUsecase, log *logger.Logger) *AuthHandler {
	return &AuthHandler{uc: uc, log: log}
}

func (h *AuthHandler) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	h.log.Info("Login called", "email", req.Email, "password", req.Password)
	// TODO: call h.uc.Login(...)
	return &authpb.LoginResponse{
		AccessToken:  "",
		RefreshToken: "",
		ExpiresAt:    0,
		TokenType:    "",
	}, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	h.log.Info("Register called", "email", req.Email)
	// TODO: call h.uc.Register(...)
	return &authpb.RegisterResponse{
		Success:     false,
		Message:     "",
		AccessToken: "",
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	h.log.Info("ValidateToken called", "token", req.Jwt)
	// TODO: call h.uc.ValidateToken(...)
	return &authpb.ValidateTokenResponse{
		Valid:     false,
		UserId:    "",
		Role:      authpb.Role_UNSPECIFIED,
		ExpiresAt: 0,
	}, nil
}

func (h *AuthHandler) GetUserByID(ctx context.Context, req *authpb.GetUserByIDRequest) (*authpb.GetUserByIDResponse, error) {
	h.log.Info("GetUserByID called", "id", req.UserId)
	// TODO: call h.uc.GetUserByID(...)
	return &authpb.GetUserByIDResponse{
		UserID: "",
		Email:  "",
		Role:   "",
	}, nil
}
