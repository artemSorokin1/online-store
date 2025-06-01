package handlers

import (
	"context"
	"dlivery_service/delivery_service/internal/config"
	"dlivery_service/delivery_service/internal/jwt"
	"dlivery_service/delivery_service/internal/models"
	"dlivery_service/delivery_service/internal/repository/storage"
	"dlivery_service/delivery_service/pkg/auth"
	"dlivery_service/delivery_service/pkg/inmem"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	grpcauth "github.com/artemSorokin1/Auth-proto/protos/gen/protos/proto"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

type Handler struct {
	GRPCClient           *auth.GRPCAuthClient
	DB                   *storage.DB
	redisClientForCart   *inmem.RedisClientForCart
	redisClientForNotify *inmem.RedisClientForNotify
	logger               *zap.Logger
}

func New(logger *zap.Logger, db *storage.DB) *Handler {
	redisCfg := config.NewRedisConfig()
	redisClientForNotify := inmem.NewRedisClientForNotify(redisCfg)
	redisClientForCart := inmem.NewRedisClientForCart(redisCfg)

	return &Handler{
		GRPCClient:           auth.New(context.Background(), logger, time.Second*1, 3),
		DB:                   db,
		redisClientForNotify: redisClientForNotify,
		redisClientForCart:   redisClientForCart,
		logger:               logger,
	}
}

func (h *Handler) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			h.logger.Debug("missing token")
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
		}

		userId, err := jwt.GetUserIdFromJWTToken(c)
		if err != nil {
			h.logger.Error("error getting user id from token", zap.Error(err))
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		response, err := h.GRPCClient.Api.RefreshTokens(context.Background(), &grpcauth.RefreshTokensRequest{
			UserId: userId,
		})
		if err != nil {
			h.logger.Error("error refreshing tokens", zap.Error(err))
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		c.Response().Header().Set("Authorization", "Bearer "+response.AccessToken)

		h.logger.Debug("user id from token", zap.Int64("userId", userId))

		return next(c)
	}
}

func (h *Handler) LoginUserHandler(c echo.Context) error {
	h.logger.Info("handling login request",
		zap.String("path", c.Path()),
		zap.String("method", c.Request().Method))
	h.logger.Debug("login attempt",
		zap.String("username", c.FormValue("username")))

	response, err := h.GRPCClient.Api.Login(context.Background(), &grpcauth.LoginRequest{
		Username: c.FormValue("username"),
		Password: c.FormValue("password"),
	})
	if err != nil {
		h.logger.Error("login failed",
			zap.String("username", c.FormValue("username")),
			zap.Error(err))
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	h.logger.Debug("user logged in successfully",
		zap.String("username", c.FormValue("username")))

	c.Response().Header().Set("Authorization", "Bearer "+response.AccessToken)
	return c.JSON(http.StatusOK, map[string]string{"message": "Login successful"})
}

func (h *Handler) RegisterUserHandler(c echo.Context) error {
	response, err := h.GRPCClient.Api.Register(context.Background(), &grpcauth.RegisterRequest{
		Email:    c.FormValue("email"),
		Username: c.FormValue("username"),
		Password: c.FormValue("password"),
	})
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	h.logger.Debug("user registered",
		zap.Int64("user_id", response.UserId))

	err = h.DB.CreateCart(response.UserId)
	if err != nil {
		h.logger.Error("failed to create cart", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	h.logger.Debug("cart created for user",
		zap.Int64("user_id", response.UserId))

	return c.JSON(http.StatusOK, map[string]string{"message": "success"})
}

func (h *Handler) GetProductsHandler(c echo.Context) error {
	h.logger.Info("handling get products request",
		zap.String("path", c.Path()),
		zap.String("method", c.Request().Method))

	products, err := h.DB.GetProducts()
	if err != nil {
		h.logger.Error("failed to get products", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	h.logger.Debug("got products", zap.Int("count", len(products)))
	return c.JSON(http.StatusOK, products)
}

func (h *Handler) GetProductByIdHandler(c echo.Context) error {
	i := c.Param("id")
	h.logger.Info("handling get product by id request",
		zap.String("product_id", i),
		zap.String("path", c.Path()),
		zap.String("method", c.Request().Method))

	id, _ := strconv.ParseInt(i, 10, 64)
	product, err := h.DB.GetProductById(id)
	if err != nil {
		h.logger.Error("failed to get product by id",
			zap.Int64("product_id", id),
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	h.logger.Debug("got product",
		zap.Int64("product_id", id),
		zap.Any("product", product))
	return c.JSON(http.StatusOK, product)
}

func (h *Handler) GetCartHandler(c echo.Context) error {
	h.logger.Info("handling get cart request",
		zap.String("path", c.Path()),
		zap.String("method", c.Request().Method))

	userId, err := jwt.GetUserIdFromJWTToken(c)
	if err != nil {
		h.logger.Error("failed to get user ID from token", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	cart, err := h.redisClientForCart.GetCart(userId)
	if err != nil {
		h.logger.Debug("cart not found in redis, trying database",
			zap.Int64("user_id", userId),
			zap.Error(err))

		cart, err = h.DB.GetCart(userId)
		if err != nil {
			h.logger.Error("failed to get cart from database",
				zap.Int64("user_id", userId),
				zap.Error(err))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}

	h.logger.Debug("got cart",
		zap.Int64("user_id", userId),
		zap.Any("cart", cart))
	return c.JSON(http.StatusOK, cart)
}

func (h *Handler) AddProductInCartHandler(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	h.logger.Debug("received add to cart request", zap.String("body", string(body)))

	var req struct {
		ProductID string `json:"productId"`
		Size      int    `json:"size"`
		Price     int    `json:"price"`
		Quantity  int    `json:"quantity"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("failed to unmarshal request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
	}

	if req.ProductID == "" {
		h.logger.Warn("invalid product ID in request")
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid product ID or quantity"})
	}

	userID, err := jwt.GetUserIdFromJWTToken(c)
	if err != nil {
		h.logger.Error("failed to get user ID from token", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	productId, err := strconv.ParseInt(req.ProductID, 10, 64)
	if err != nil {
		h.logger.Error("failed to parse product ID",
			zap.String("product_id", req.ProductID),
			zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	err = h.DB.AddProductInCart(userID, productId, int64(req.Size), int64(req.Price))
	if err != nil {
		h.logger.Error("failed to add product to cart",
			zap.Int64("user_id", userID),
			zap.Int64("product_id", productId),
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	h.logger.Debug("product added to cart successfully",
		zap.Int64("user_id", userID),
		zap.Int64("product_id", productId),
		zap.Int("size", req.Size))

	return c.JSON(http.StatusOK, map[string]string{"message": "success"})
}

func (h *Handler) DeleteCartHandler(c echo.Context) error {
	userId, err := jwt.GetUserIdFromJWTToken(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	err = h.DB.DeleteCart(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "success"})
}

func (h *Handler) DeleteCartItemHandler(c echo.Context) error {
	h.logger.Info("handling delete cart item request",
		zap.String("path", c.Path()),
		zap.String("method", c.Request().Method))

	userId, err := jwt.GetUserIdFromJWTToken(c)
	if err != nil {
		h.logger.Error("failed to get user ID from token", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		h.logger.Error("failed to read request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	h.logger.Debug("received delete cart item request",
		zap.String("body", string(body)))

	var req struct {
		ProductID int `json:"productId"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("failed to unmarshal request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
	}

	h.logger.Debug("deleting cart item",
		zap.Int64("user_id", userId),
		zap.Int("product_id", req.ProductID))

	if err = h.DB.DeleteCartItem(userId, int64(req.ProductID)); err != nil {
		h.logger.Error("failed to delete cart item",
			zap.Int64("user_id", userId),
			zap.Int("product_id", req.ProductID),
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	h.logger.Debug("cart item deleted successfully",
		zap.Int64("user_id", userId),
		zap.Int("product_id", req.ProductID))

	return c.JSON(http.StatusOK, map[string]string{"message": "success"})
}

func (h *Handler) CheckoutHandler(c echo.Context) error {
	h.logger.Info("handling checkout request",
		zap.String("path", c.Path()),
		zap.String("method", c.Request().Method))

	name := c.FormValue("name")
	address := c.FormValue("address")
	phone := c.FormValue("phone")

	h.logger.Debug("processing payment info",
		zap.String("name", name),
		zap.String("phone", phone))

	userId, err := jwt.GetUserIdFromJWTToken(c)
	if err != nil {
		h.logger.Error("failed to get user ID from token", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	h.logger.Debug("processing checkout for user", zap.Int64("user_id", userId))

	cartInfo, err := h.DB.GetCartInfo(userId)
	if err != nil {
		h.logger.Error("failed to fetch cart info",
			zap.Int64("user_id", userId),
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if len(cartInfo) == 0 {
		h.logger.Warn("attempting to checkout empty cart", zap.Int64("user_id", userId))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cart is empty"})
	}

	h.logger.Debug("got cart info",
		zap.Int64("user_id", userId),
		zap.Any("cart_items", cartInfo))

	var productIds pq.Int64Array
	var sizes pq.Int64Array

	for _, item := range cartInfo {
		productIds = append(productIds, item.ProductId)
		sizes = append(sizes, item.Size)
	}

	total, err := h.DB.GetTotalPrice(userId)
	if err != nil {
		h.logger.Error("failed to calculate total price",
			zap.Int64("user_id", userId),
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	h.logger.Debug("calculated total price",
		zap.Int64("user_id", userId),
		zap.Int("total", total))

	err = h.DB.CreateOrder(userId, name, address, phone, productIds, sizes, total)
	if err != nil {
		h.logger.Error("failed to create order",
			zap.Int64("user_id", userId),
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Удаление корзины после оформления заказа
	err = h.DB.DeleteCart(userId)
	if err != nil {
		h.logger.Error("failed to clear cart",
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	orderId, err := h.DB.GetOrderId(userId)
	if err != nil {
		h.logger.Error("failed to fetch order ID",
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	err = h.DB.AddOrderInUserActions(userId, orderId)
	if err != nil {
		h.logger.Error("failed to add order in user actions",
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	h.logger.Debug("order processed successfully",
		zap.Int64("user_id", userId),
		zap.Int("order_id", orderId))
	user, err := h.DB.GetUserById(userId)
	if err != nil {
		h.logger.Error("failed to get user by id",
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	h.logger.Debug("sending notification",
		zap.String("email", user.Email))

	go h.redisClientForNotify.Publish(user.Email)

	return c.JSON(http.StatusOK, map[string]string{
		"message":  "Order created successfully",
		"order_id": fmt.Sprintf("%d", orderId),
	})
}

func (h *Handler) LogoutUserHandler(c echo.Context) error {
	h.logger.Info("handling logout request",
		zap.String("path", c.Path()),
		zap.String("method", c.Request().Method))

	userId, err := jwt.GetUserIdFromJWTToken(c)
	if err != nil {
		h.logger.Error("failed to get user ID from token",
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	err = h.DB.DeleteCart(userId)
	if err != nil {
		h.logger.Error("failed to delete cart",
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	h.logger.Debug("cart deleted for user",
		zap.Int64("user_id", userId))

	req, err := h.GRPCClient.Api.Logout(context.Background(), &grpcauth.LogoutRequest{
		UserId: userId,
	})
	if err != nil {
		slog.Error("error logging out", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	if !req.IsSuccess {
		slog.Error("error logging out")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "logout failed"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Logout successful"})
}

func (h *Handler) AdminAddProductHandler(c echo.Context) error {
	h.logger.Info("handling admin add product request",
		zap.String("path", c.Path()),
		zap.String("method", c.Request().Method))

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		h.logger.Error("failed to read request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	h.logger.Debug("received product data", zap.String("body", string(body)))

	var product models.Product
	if err := json.Unmarshal(body, &product); err != nil {
		h.logger.Error("failed to unmarshal product data", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
	}

	if err = h.DB.AdminAddProduct(product); err != nil {
		h.logger.Error("failed to add product",
			zap.Any("product", product),
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	p, err := h.DB.GetLastInsertedProduct()
	if err != nil {
		h.logger.Error("failed to get inserted product", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	h.logger.Debug("product added successfully", zap.Any("product", p))
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) AdminDeleteProductHandler(c echo.Context) error {
	product_id := c.Param("id")

	id, err := strconv.ParseInt(product_id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid product id"})
	}

	err = h.DB.AdminDeleteProduct(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "product deleted successfully"})
}

func (h *Handler) AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId, err := jwt.GetUserIdFromJWTToken(c)
		if err != nil {
			h.logger.Error("failed to get user ID from token", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		h.logger.Debug("checking admin rights", zap.Int64("user_id", userId))

		isAdminResponse, err := h.GRPCClient.Api.IsAdmin(context.Background(), &grpcauth.IsAdminRequest{
			UserId: userId,
		})
		if err != nil {
			h.logger.Error("failed to check admin rights",
				zap.Int64("user_id", userId),
				zap.Error(err))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		h.logger.Debug("admin check response",
			zap.Int64("user_id", userId),
			zap.Bool("is_admin", isAdminResponse.IsAdmin))

		if !isAdminResponse.IsAdmin {
			h.logger.Warn("access denied: user is not admin", zap.Int64("user_id", userId))
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
		}

		return next(c)
	}
}
