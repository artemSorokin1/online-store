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
	grpcauth "github.com/artemSorokin1/Auth-proto/protos/gen/protos/proto"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	GRPCClient  *auth.GRPCAuthClient
	DB          *storage.DB
	redisClient *inmem.RedisClient
}

func New(db *storage.DB) *Handler {
	redisCfg := config.NewRedisConfig()
	redisClient := inmem.NewRedisClient(redisCfg)

	return &Handler{
		GRPCClient:  auth.New(context.Background(), time.Second*1, 3),
		DB:          db,
		redisClient: redisClient,
	}
}

func (h *Handler) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
		}

		userId, err := jwt.GetUserIdFromJWTToken(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		response, err := h.GRPCClient.Api.RefreshTokens(context.Background(), &grpcauth.RefreshTokensRequest{
			UserId: userId,
		})
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		c.Response().Header().Set("Authorization", "Bearer "+response.AccessToken)

		return next(c)
	}
}

func (h *Handler) LoginUserHandler(c echo.Context) error {
	response, err := h.GRPCClient.Api.Login(context.Background(), &grpcauth.LoginRequest{
		Username: c.FormValue("username"),
		Password: c.FormValue("password"),
	})
	if err != nil {
		fmt.Println("error logging in: ", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

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
	slog.Info(fmt.Sprintf("user registered with id: %d", response.UserId))

	err = h.DB.CreateCart(response.UserId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	fmt.Println(response.UserId)

	return c.JSON(http.StatusOK, map[string]string{"message": "success"})
}

func (h *Handler) GetProductsHandler(c echo.Context) error {
	products, err := h.DB.GetProducts()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, products)
}

func (h *Handler) GetProductByIdHandler(c echo.Context) error {
	i := c.Param("id")
	id, _ := strconv.ParseInt(i, 10, 64)
	product, err := h.DB.GetProductById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, product)
}

func (h *Handler) GetCartHandler(c echo.Context) error {
	userId, err := jwt.GetUserIdFromJWTToken(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	fmt.Println("user id: ", userId)
	cart, err := h.DB.GetCart(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, cart)
}

func (h *Handler) AddProductInCartHandler(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	fmt.Println("body ", string(body))
	var req struct {
		ProductID string `json:"productId"`
		Size      int    `json:"size"`
		Price     int    `json:"price"`
		Quantity  int    `json:"quantity"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
	}

	if req.ProductID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid product ID or quantity"})
	}

	userID, err := jwt.GetUserIdFromJWTToken(c)
	if err != nil {
		fmt.Println("error getting user id from jwt token: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	productId, err := strconv.ParseInt(req.ProductID, 10, 64)
	if err != nil {
		fmt.Println("error parsing product id: ", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	err = h.DB.AddProductInCart(userID, productId, int64(req.Size), int64(req.Price))
	if err != nil {
		fmt.Println("error adding product in cart: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

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
	userId, err := jwt.GetUserIdFromJWTToken(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	body, err := io.ReadAll(c.Request().Body)
	fmt.Println("body ", string(body))

	var req struct {
		ProductID int `json:"productId"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
	}

	fmt.Println("req.ProductID: ", req.ProductID)

	err = h.DB.DeleteCartItem(userId, int64(req.ProductID))

	return nil
}

func (h *Handler) CheckoutHandler(c echo.Context) error {
	name := c.FormValue("name")
	address := c.FormValue("address")
	phone := c.FormValue("phone")
	cvv := c.FormValue("cvv")
	cardNumber := c.FormValue("cardNumber")
	expDate := c.FormValue("expDate")

	// Здесь будет ваша логика обработки оплаты
	_ = cvv
	_ = cardNumber
	_ = expDate

	userId, err := jwt.GetUserIdFromJWTToken(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	fmt.Println("user id: ", userId)

	// Получаем информацию о корзине
	cartInfo, err := h.DB.GetCartInfo(userId)
	if err != nil {
		fmt.Println("Error fetching cart info:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if len(cartInfo) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cart is empty"})
	}

	fmt.Println("cart info: ", cartInfo)

	// Подготовка массивов productIds и sizes
	var productIds pq.Int64Array
	var sizes pq.Int64Array

	for _, item := range cartInfo {
		productIds = append(productIds, item.ProductId)
		sizes = append(sizes, item.Size)
	}

	// Получение общей суммы заказа
	total, err := h.DB.GetTotalPrice(userId)
	if err != nil {
		fmt.Println("Error calculating total price:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	fmt.Println("total: ", total)

	// Создание заказа
	err = h.DB.CreateOrder(userId, name, address, phone, productIds, sizes, total)
	if err != nil {
		fmt.Println("Error creating order:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Удаление корзины после оформления заказа
	err = h.DB.DeleteCart(userId)
	if err != nil {
		fmt.Println("Error clearing cart:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	orderId, err := h.DB.GetOrderId(userId)
	if err != nil {
		fmt.Println("Error fetching order ID:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	err = h.DB.AddOrderInUserActions(userId, orderId)
	if err != nil {
		fmt.Println("Error adding order in user actions:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	user, err := h.DB.GetUserById(userId)
	if err != nil {
		slog.Error("error getting user by id", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	go h.redisClient.Publish(user.Email)

	return c.JSON(http.StatusOK, map[string]string{"message": "Order created successfully", "order_id": fmt.Sprintf("%d", orderId)})
}

func (h *Handler) LogoutUserHandler(c echo.Context) error {
	userId, err := jwt.GetUserIdFromJWTToken(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	err = h.DB.DeleteCart(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

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
	// Извлечение данных из запроса
	var product models.Product

	body, err := io.ReadAll(c.Request().Body)
	fmt.Println("body ", string(body))

	if err := json.Unmarshal(body, &product); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
	}

	err = h.DB.AdminAddProduct(product)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	p, err := h.DB.GetLastInsertedProduct()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	fmt.Println("product added successfully")
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
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		fmt.Println("user id: ", userId)

		isAdminResponse, err := h.GRPCClient.Api.IsAdmin(context.Background(), &grpcauth.IsAdminRequest{
			UserId: userId,
		})
		fmt.Println("isAdminResponse: ", isAdminResponse)
		if !isAdminResponse.IsAdmin {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
		}
		return next(c)
	}
}
