package JWT

import (
	"auth/internal/config"
	"auth/internal/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	TimeExpired  = errors.New("token has expired")
	InvalidToken = errors.New("invalid token")
)

func CreateAccessTokenSeller(user models.User, cfg *config.Config) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"exp":      time.Now().Add(cfg.SellerTokenCfg.AccessTokenTTL).Unix(),
		"email":    user.Email,
		"username": user.Username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err := token.SignedString([]byte(cfg.SellerTokenCfg.AccessSecret))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func CreateAccessTokenCustomer(user models.User, cfg *config.Config) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"exp":      time.Now().Add(cfg.CustomerTokenCfg.AccessTokenTTL).Unix(),
		"email":    user.Email,
		"username": user.Username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err := token.SignedString([]byte(cfg.CustomerTokenCfg.AccessSecret))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func CreateRefreshTokenSeller(user models.User, cfg *config.Config) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"exp":      time.Now().Add(cfg.SellerTokenCfg.RefreshTokenTTL).Unix(),
		"email":    user.Email,
		"username": user.Username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	refreshToken, err := token.SignedString([]byte(cfg.SellerTokenCfg.RefreshSecret))
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func CreateRefreshTokenCustomer(user models.User, cfg *config.Config) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"exp":      time.Now().Add(cfg.CustomerTokenCfg.RefreshTokenTTL).Unix(),
		"email":    user.Email,
		"username": user.Username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	refreshToken, err := token.SignedString([]byte(cfg.CustomerTokenCfg.RefreshSecret))
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func RefreshTokenSeller(tokenString string, cfg *config.Config) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.SellerTokenCfg.RefreshSecret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", InvalidToken
	}

	if exp, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		if expTime.Before(time.Now()) {
			return "", TimeExpired
		}
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return "", InvalidToken
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", InvalidToken
	}

	accessToken, err := CreateAccessTokenSeller(models.User{
		ID:       userID,
		Email:    claims["email"].(string),
		Username: claims["username"].(string),
	}, cfg)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func RefreshTokenCustomer(tokenString string, cfg *config.Config) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.CustomerTokenCfg.RefreshSecret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", InvalidToken
	}

	if exp, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		if expTime.Before(time.Now()) {
			return "", TimeExpired
		}
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return "", InvalidToken
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", InvalidToken
	}

	accessToken, err := CreateAccessTokenCustomer(models.User{
		ID:       userID,
		Email:    claims["email"].(string),
		Username: claims["username"].(string),
	}, cfg)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func VerifyToken(tokenString string, secret []byte) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

}
