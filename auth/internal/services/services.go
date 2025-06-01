package services

import (
	"auth/internal/config"
	"auth/internal/models"
	"auth/internal/models/dto"
	"auth/internal/repository/postgres"
	"auth/pkg/JWT"
	"auth/pkg/hash"
	"fmt"
	"log"
)

type Service struct {
	repository *postgres.Storage
	cfg        *config.Config
}

func NewService(cfg *config.Config, repository *postgres.Storage) *Service {
	return &Service{repository: repository, cfg: cfg}
}

func (s *Service) LoginUser(credentials dto.UserLoginCredentials) (string, string, error) {
	user, err := s.repository.VerifyUserWithCredentials(credentials.Username, credentials.Password)
	if err != nil {
		log.Println("Error verifying user with credentials:", err)
		return "", "", err
	}

	accessToken, err := JWT.CreateAccessToken(user, s.cfg)
	if err != nil {
		log.Println("Error creating access token:", err)
		return "", "", err
	}

	refreshToken, err := JWT.CreateRefreshToken(user, s.cfg)
	if err != nil {
		log.Println("Error creating refresh token:", err)
		return "", "", err

	}

	return accessToken, refreshToken, nil
}

func (s *Service) RegisterUser(credentials dto.UserRegistrationCredentials) error {
	if isExists, err := s.repository.UserExists(credentials.Email, credentials.Username); isExists || err != nil {
		return fmt.Errorf("user with email or username already exists")
	}

	passHash, err := hash.HashPassword(credentials.Password)
	if err != nil {
		log.Println("Error hashing password:", err)
		return err
	}

	user := models.User{
		Email:    credentials.Email,
		Username: credentials.Username,
		PassHash: passHash,
	}

	_, err = s.repository.CreateUser(user)
	if err != nil {
		log.Println("Error creating user:", err)
		return err
	}

	return nil
}

func (s *Service) RefreshAccessToken(refreshToken string, cfg *config.Config) (string, error) {
	if refreshToken == "" {
		return "", fmt.Errorf("empty refresh token")
	}

	accessToken, err := JWT.RefreshToken(refreshToken, cfg)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
