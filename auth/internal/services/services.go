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

func (s *Service) SearchSeller(username string) (*models.User, error) {
	sellers, err := s.repository.SearchSellerByUsername(username)
	if err != nil {
		log.Println("Error searching sellers:", err)
		return nil, fmt.Errorf("error searching sellers: %w", err)
	}

	return sellers, nil
}

func (s *Service) LoginSeller(credentials dto.UserLoginCredentials) (string, string, error) {
	user, err := s.repository.VerifySellerWithCredentials(credentials.Username, credentials.Password)
	if err != nil {
		log.Println("Error verifying user with credentials:", err)
		return "", "", err
	}

	accessToken, err := JWT.CreateAccessTokenSeller(user, s.cfg)
	if err != nil {
		log.Println("Error creating access token:", err)
		return "", "", err
	}

	refreshToken, err := JWT.CreateRefreshTokenSeller(user, s.cfg)
	if err != nil {
		log.Println("Error creating refresh token:", err)
		return "", "", err

	}

	return accessToken, refreshToken, nil
}

func (s *Service) LoginCustomer(credentials dto.UserLoginCredentials) (string, string, error) {
	user, err := s.repository.VerifyCustomerWithCredentials(credentials.Username, credentials.Password)
	if err != nil {
		log.Println("Error verifying user with credentials:", err)
		return "", "", err
	}

	accessToken, err := JWT.CreateAccessTokenCustomer(user, s.cfg)
	if err != nil {
		log.Println("Error creating access token:", err)
		return "", "", err
	}

	refreshToken, err := JWT.CreateRefreshTokenCustomer(user, s.cfg)
	if err != nil {
		log.Println("Error creating refresh token:", err)
		return "", "", err

	}

	return accessToken, refreshToken, nil
}

func (s *Service) RegisterSeller(credentials dto.UserRegistrationCredentials) error {
	if isExists, err := s.repository.SellerExists(credentials.Email, credentials.Username); isExists || err != nil {
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

	_, err = s.repository.CreateSeller(user)
	if err != nil {
		log.Println("Error creating user:", err)
		return err
	}

	return nil
}

func (s *Service) RegisterCustomer(credentials dto.UserRegistrationCredentials) error {
	if isExists, err := s.repository.CustomerExists(credentials.Email, credentials.Username); isExists || err != nil {
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

	_, err = s.repository.CreateCustomer(user)
	if err != nil {
		log.Println("Error creating user:", err)
		return err
	}

	return nil
}

func (s *Service) RefreshAccessTokenSeller(refreshToken string, cfg *config.Config) (string, error) {
	if refreshToken == "" {
		return "", fmt.Errorf("empty refresh token")
	}

	accessToken, err := JWT.RefreshTokenSeller(refreshToken, cfg)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (s *Service) RefreshAccessTokenCustomer(refreshToken string, cfg *config.Config) (string, error) {
	if refreshToken == "" {
		return "", fmt.Errorf("empty refresh token")
	}

	accessToken, err := JWT.RefreshTokenCustomer(refreshToken, cfg)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
