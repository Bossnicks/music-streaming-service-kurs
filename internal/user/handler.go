package user

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Bossnicks/music-streaming-service-kurs/pkg/auth"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Регистрация пользователя
func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректные данные"})
	}

	user := &User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Avatar:   req.Avatar,
	}

	err := h.service.RegisterUser(user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Пользователь уже существует!"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Успешная регистрация"})
}

// Авторизация пользователя
func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат запроса"})
	}

	token, user, err := h.service.Authenticate(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверные учетные данные"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

// Получение информации о пользователе
func (h *Handler) GetUser(c echo.Context) error {

	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	userID := claims.UserID

	// userID, err := strconv.Atoi(c.Param("id"))
	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID"})
	// }

	user, err := h.service.GetUser(userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Пользователь не найден"})
	}
	if tokenString != user.Token {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен устарел"})
	}
	return c.JSON(http.StatusOK, user)
}

// Обновление информации о пользователе
func (h *Handler) UpdateUser(c echo.Context) error {
	userID := c.Get("id").(int)

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректные данные"})
	}

	err := h.service.UpdateUser(userID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка обновления данных"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Данные пользователя обновлены"})
}

// Получение аватара
func (h *Handler) GetAvatar(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID"})
	}

	avatar, err := h.service.GetAvatar(userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Аватар не найден"})
	}

	return c.Blob(http.StatusOK, "image/png", avatar)
}

func (h *Handler) FollowUser(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	followingUserID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	if err := h.service.FollowUser(claims.UserID, followingUserID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to follow user"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Followed successfully"})
}

func (h *Handler) UnfollowUser(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	followingUserID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	if err := h.service.UnfollowUser(claims.UserID, followingUserID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unfollow user"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Unfollowed successfully"})
}

func (h *Handler) GetFollowers(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	followers, err := h.service.GetFollowers(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get followers"})
	}

	return c.JSON(http.StatusOK, followers)
}

func (h *Handler) GetFollowing(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	following, err := h.service.GetFollowing(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get following"})
	}

	return c.JSON(http.StatusOK, following)
}

func (h *Handler) IsUserSubscribed(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	subscribed, err := h.service.IsUserSubscribed(claims.UserID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]bool{"subscribed": subscribed})
}
