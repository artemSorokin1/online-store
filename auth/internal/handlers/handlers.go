package handlers

import (
	"auth/internal/config"
	"auth/internal/models/dto"
	"auth/internal/repository/postgres"
	"auth/internal/services"
	"auth/pkg/JWT"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	cfg     *config.Config
	service *services.Service
}

func NewHandler(cfg *config.Config, storage *postgres.Storage) *Handler {
	return &Handler{
		cfg:     cfg,
		service: services.NewService(cfg, storage),
	}
}

func (h *Handler) SearchSeller(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sellers, err := h.service.SearchSeller(req.Username)
	if err != nil {
		log.Println("Error searching sellers:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sellers)
}

func (h *Handler) SignUpSeller(c *gin.Context) {
	var credentials dto.UserRegistrationCredentials
	if err := c.BindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.RegisterSeller(credentials)
	fmt.Println(err)
	if err != nil {
		log.Println("Error registering user:", err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created"})
}

func (h *Handler) SignUpCustomer(c *gin.Context) {
	var credentials dto.UserRegistrationCredentials
	if err := c.BindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.RegisterCustomer(credentials)
	fmt.Println(err)
	if err != nil {
		log.Println("Error registering user:", err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created"})
}

func (h *Handler) SignInCustomer(c *gin.Context) {
	var credentials dto.UserLoginCredentials
	if err := c.BindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.service.LoginCustomer(credentials)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"refresh_token",
		refreshToken,
		int(h.cfg.CustomerTokenCfg.RefreshTokenTTL.Seconds()),
		"/",
		h.cfg.ServerCfg.Domain,
		false, // if https - true
		true)

	c.Header("Authorization", "Bearer "+accessToken)
}

func (h *Handler) SignInSeller(c *gin.Context) {
	var credentials dto.UserLoginCredentials
	if err := c.BindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.service.LoginSeller(credentials)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"refresh_token",
		refreshToken,
		int(h.cfg.SellerTokenCfg.RefreshTokenTTL.Seconds()),
		"/",
		h.cfg.ServerCfg.Domain,
		false, // if https - true
		true)

	c.Header("Authorization", "Bearer "+accessToken)
}

func (h *Handler) Logout(c *gin.Context) {
	// Удаляем refresh token из cookies
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		h.cfg.ServerCfg.Domain,
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (h *Handler) RefreshSeller(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token is required"})
		return
	}

	accessToken, err := h.service.RefreshAccessTokenSeller(refreshToken, h.cfg)
	if err != nil {
		if errors.Is(err, JWT.InvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		} else if errors.Is(err, JWT.TimeExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.Header("Authorization", "Bearer "+accessToken)
}

func (h *Handler) RefreshCustomer(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token is required"})
		return
	}

	accessToken, err := h.service.RefreshAccessTokenCustomer(refreshToken, h.cfg)
	if err != nil {
		if errors.Is(err, JWT.InvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		} else if errors.Is(err, JWT.TimeExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.Header("Authorization", "Bearer "+accessToken)
}
