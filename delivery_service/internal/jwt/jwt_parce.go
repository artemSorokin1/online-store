package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"log/slog"
	"os"
	"strings"
)

func GetUserIdFromJWTToken(c echo.Context) (int64, error) {
	bearerToken := c.Request().Header.Get("Authorization")
	if bearerToken == "" {
		slog.Warn("bearer token not found")
		return 0, fmt.Errorf("bearer token not found")
	}

	parce := strings.Split(bearerToken, " ")

	if parce[0] != "Bearer" || len(parce) != 2 {
		slog.Warn("error parsing token")
		return 0, fmt.Errorf("error parsing token")
	}

	accessSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	token, err := jwt.Parse(parce[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(accessSecret), nil
	})

	if err != nil || !token.Valid {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, err
	}

	userId, ok := claims["user_id"].(float64)
	if !ok {
		return 0, err
	}

	return int64(userId), nil

}
