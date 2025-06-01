package JWT

import (
	"auth/internal/config"
	"auth/internal/models"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var (
	TimeExpired  = errors.New("token has expired")
	InvalidToken = errors.New("invalid token")
)

func CreateAccessToken(user models.User, cfg *config.Config) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"exp":      time.Now().Add(cfg.TokenCfg.AccessTokenTTL).Unix(),
		"email":    user.Email,
		"username": user.Username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err := token.SignedString([]byte(cfg.TokenCfg.AccessSecret))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func CreateRefreshToken(user models.User, cfg *config.Config) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"exp":      time.Now().Add(cfg.TokenCfg.RefreshTokenTTL).Unix(),
		"email":    user.Email,
		"username": user.Username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	refreshToken, err := token.SignedString([]byte(cfg.TokenCfg.RefreshSecret))
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func RefreshToken(tokenString string, cfg *config.Config) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.TokenCfg.RefreshSecret), nil
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

	accessToken, err := CreateAccessToken(models.User{
		ID:       int(claims["user_id"].(float64)),
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
