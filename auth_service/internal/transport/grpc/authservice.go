package grpc

import (
	"auth_service/internal/jwt"
	"auth_service/internal/models"
	"auth_service/internal/repositiry/storage"
	"auth_service/pkg/storage/inmem"
	"fmt"

	//"auth_service/pkg/api"
	"context"
	api "github.com/artemSorokin1/Auth-proto/protos/gen/protos/proto"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

type AuthService struct {
	api.UnimplementedAuthServiceServer
	stor                *storage.Storage
	refreshInMemStorage inmem.RefreshTokenStorage
}

func (s *AuthService) Register(ctx context.Context, req *api.RegisterRequest) (*api.RegisterResponse, error) {
	slog.Info("Register method called")

	passHash, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		slog.Warn("error hashing password")
		return nil, err
	}

	user := models.User{
		Email:    req.GetEmail(),
		Username: req.GetUsername(),
		PassHash: string(passHash),
		Role:     "user",
	}

	id, err := s.stor.AddNewUser(user)
	if err != nil || id == -1 {
		slog.Warn("error adding new user", err)
		return nil, err
	}

	return &api.RegisterResponse{
		UserId: id,
	}, nil

}

func (s *AuthService) Login(ctx context.Context, req *api.LoginRequest) (*api.LoginResponse, error) {
	slog.Info("Login method called")

	user, err := s.stor.GetUser(req.GetUsername())
	if err != nil {
		slog.Warn("user not found in login")
		return nil, storage.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(req.GetPassword())); err != nil {
		slog.Warn("error comparing hashes")
		return nil, storage.ErrUserNotFound
	}

	accessToken, err := jwt.CreateAccessToken(&user)
	if err != nil {
		slog.Warn("error creating access token")
		return nil, err
	}
	refreshToken, err := jwt.CreateRefreshToken(&user)
	if err != nil {
		slog.Warn("error creating refresh token")
		return nil, err
	}

	err = s.refreshInMemStorage.SaveToken(context.Background(), user.ID, refreshToken)
	if err != nil {
		slog.Warn("error saving refresh token to redis")
		return nil, fmt.Errorf("error saving refresh token to redis: %w", err)
	}

	return &api.LoginResponse{
		AccessToken: accessToken,
	}, nil
}

func (s *AuthService) IsAdmin(ctx context.Context, req *api.IsAdminRequest) (*api.IsAdminResponse, error) {
	slog.Info("IsAdmin method called")

	isAdmin, err := s.stor.IsAdmin(req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &api.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil

}

func (s *AuthService) RefreshTokens(ctx context.Context, req *api.RefreshTokensRequest) (*api.RefreshTokensResponse, error) {
	slog.Info("RefreshTokens method called")

	userId := req.GetUserId()
	refreshToken, err := s.refreshInMemStorage.GetToken(context.Background(), userId)
	if err != nil {
		slog.Warn("error getting refresh token from redis")
		return nil, err
	}

	isValid, err := jwt.IsValidRefreshToken(refreshToken)
	if err != nil {
		slog.Warn("error validating refresh token")
		return nil, err
	}

	if !isValid {
		slog.Warn("refresh token is not valid")
		return nil, fmt.Errorf("refresh token is not valid")
	}

	user, err := s.stor.GetUserById(userId)
	if err != nil {
		slog.Warn("user not found")
		return nil, err
	}

	refreshToken, err = jwt.CreateRefreshToken(&user)
	if err != nil {
		slog.Warn("error creating refresh token")
		return nil, err
	}

	// так плохо, но пока так
	go s.refreshInMemStorage.SaveToken(context.Background(), userId, refreshToken)

	accessToken, err := jwt.CreateAccessToken(&user)
	if err != nil {
		slog.Warn("error creating access token")
		return nil, err
	}

	return &api.RefreshTokensResponse{
		AccessToken: accessToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *api.LogoutRequest) (*api.LogoutResponse, error) {
	slog.Info("Logout method called")

	err := s.refreshInMemStorage.RemoveToken(context.Background(), req.GetUserId())
	if err != nil {
		slog.Warn("error removing refresh token from redis")
		return nil, err
	}

	return &api.LogoutResponse{
		IsSuccess: true,
	}, nil
}
