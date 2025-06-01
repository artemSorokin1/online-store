package middleware

import (
	"auth/internal/config"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := c.GetHeader("Authorization")
		if accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "empty token"})
			return
		}

		c.Next()
	}
}
