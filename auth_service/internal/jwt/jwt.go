package jwt

import (
	"auth_service/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"os"
	"time"
)

func CreateAccessToken(user *models.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	accessTTLString := os.Getenv("ACCESS_TOKEN_TTL")
	accessTTL, err := time.ParseDuration(accessTTLString)
	if err != nil {
		slog.Warn("error parsing access ttl")
		return "", err
	}

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["exp"] = time.Now().Add(accessTTL).Unix()
	claims["email"] = user.Email
	claims["username"] = user.Username

	secretKey := os.Getenv("ACCESS_TOKEN_SECRET")
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		slog.Warn("error creating token")
		return "", err
	}

	return tokenString, nil

}

func CreateRefreshToken(user *models.User) (string, error) {
	refreshTTLString := os.Getenv("REFRESH_TOKEN_TTL")
	refreshTTL, err := time.ParseDuration(refreshTTLString)
	if err != nil {
		slog.Warn("error parsing refresh ttl")
		return "", err
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"exp":      time.Now().Add(refreshTTL).Unix(),
		"email":    user.Email,
		"username": user.Username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey := os.Getenv("REFRESH_TOKEN_SECRET")
	refreshToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func IsValidRefreshToken(tokenString string) (bool, error) {
	claims := jwt.MapClaims{}
	secretKey := os.Getenv("REFRESH_TOKEN_SECRET")
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		slog.Warn("error parsing token")
		return false, err
	}

	if !token.Valid {
		slog.Warn("invalid token")
		return false, nil
	}

	return true, nil
}
