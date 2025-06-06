package user

import (
	"fmt"
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
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Некорректные данные"})
	}

	user := &User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Avatar:   req.Avatar,
	}

	err := h.service.RegisterUser(user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Пользователь уже существует!"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Успешная регистрация"})
}

// Авторизация пользователя
func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Неверный формат запроса"})
	}

	token, user, err := h.service.Authenticate(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Неверные учетные данные"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (h *Handler) RecoverPassword(c echo.Context) error {
	var req RecoverRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Некорректные данные"})
	}

	// Генерация токена для сброса пароля
	token, err := auth.GenerateResetToken(req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Ошибка генерации токена"})
	}

	// Формируем ссылку для сброса пароля
	resetLink := fmt.Sprintf("http://localhost:5173/resetpassword?token=%s", token)

	// Отправка письма
	err = auth.SendResetEmail(req.Email, resetLink)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Ошибка отправки письма"})
	}

	fmt.Println(req.Email, token)

	err = h.service.UpdateUserResetToken(req.Email, token)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Не удалось установить токен"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Успешно отправлено"})
}

func (h *Handler) SendVerification(c echo.Context) error {
	var req RecoverRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Некорректные данные"})
	}

	fmt.Println(req.Email)

	// Генерация токена для сброса пароля
	// token, err := auth.GenerateResetToken(req.Email)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Ошибка генерации токена"})
	// }

	// // Формируем ссылку для сброса пароля
	// resetLink := fmt.Sprintf("http://localhost:5173/resetpassword?token=%s", token)

	// // Отправка письма
	// err = auth.SendResetEmail(req.Email, resetLink)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Ошибка отправки письма"})
	// }

	return c.JSON(http.StatusCreated, map[string]string{"message": "Успешно отправлено"})
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

func (h *Handler) BlockComments(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	if claims.Role == "user" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Недостаточно прав"})
	}

	fmt.Println(claims.Role)

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
	}

	user, err := h.service.GetUser(userID)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
	}

	fmt.Println(user.Role)

	if user.Role == "admin" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Недостаточно прав"})
	}

	err = h.service.BlockComments(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to block comments"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "comments blocked"})
}

func (h *Handler) UnblockComments(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	if claims.Role == "user" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Недостаточно прав"})
	}

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
	}

	err = h.service.UnblockComments(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to unblock comments"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "comments unblocked"})
}

func (h *Handler) IsCommentAbilityBlocked(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	if claims.Role == "user" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Недостаточно прав"})
	}

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	user, err := h.service.GetUser(userID)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
	}

	fmt.Println(user.Role)

	if user.Role == "admin" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Недостаточно прав"})
	}

	commentBlocked, err := h.service.IsCommentAbilityBlocked(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]bool{"commentBlocked": commentBlocked})
}

func (h *Handler) GetUserFeed(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	feed, err := h.service.GetUserFeed(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Не удалось получить ленту"})
	}

	return c.JSON(http.StatusOK, feed)
}

// SearchHandler обрабатывает запрос на поиск
func (h *Handler) SearchHandler(c echo.Context) error {

	query := c.QueryParam("q")
	entityTypes := strings.Split(c.QueryParam("type"), ",") // Получаем список категорий
	genre := c.QueryParam("genre")                          // Получаем жанр
	sortField := c.QueryParam("sort")
	order := c.QueryParam("order")

	authHeader := c.Request().Header.Get("Authorization")
	isAdmin := false

	if authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		fmt.Println("Token:", tokenString) // Проверяем, что приходит
		claims, err := auth.ParseJWT(tokenString)
		if err != nil {
			fmt.Println("JWT Error:", err)
		}
		if err == nil && claims.Role == "admin" {
			isAdmin = true
		}
	}
	fmt.Println("Token:", isAdmin)

	// Устанавливаем значения по умолчанию
	if sortField == "" {
		sortField = "title"
	}
	if order == "" {
		order = "asc"
	}

	// Выполняем поиск
	result, err := h.service.Search(query, entityTypes, genre, sortField, order, isAdmin)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) ResetPassword(c echo.Context) error {

	fmt.Println("1")
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	fmt.Println("1")

	if req.NewPassword != req.ConfirmPassword {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Passwords don't match"})
	}
	fmt.Println(req.Token, req.NewPassword, req.ConfirmPassword)

	claims, _ := auth.ParseResetToken(req.Token)
	fmt.Println(claims.Email)
	// if err != nil {
	// 	return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token"})
	// }
	fmt.Println("1")

	// if err := h.service.IsValidResetToken(claims.Email, req.Token); err != nil && {
	// 	return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	// }
	fmt.Println("1")

	isValid, err := h.service.IsValidResetToken(req.Token, claims.Email)

	fmt.Println(isValid)

	if err != nil {
		fmt.Println("token good")
		fmt.Println("1")

	} else if !isValid {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	fmt.Println("1")

	err = h.service.ResetPassword(claims.Email, req.NewPassword)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	fmt.Println("dasdasdas")

	return c.JSON(http.StatusOK, map[string]string{"success": "Password updated successfully"})
}

func (h *Handler) HideAlbum(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	// if claims.Role == "user" {
	// 	return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Недостаточно прав"})
	// }

	fmt.Println(claims.Role)

	albumID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
	}

	// user, err := h.service.GetUser()

	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
	// }

	// fmt.Println(user.Role)

	// if user.Role == "admin" {
	// 	return c.JSON(http.StatusBadRequest, map[string]string{"error": "Недостаточно прав"})
	// }

	err = h.service.HideAlbum(claims.UserID, albumID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to hide album"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "album hidden"})
}

func (h *Handler) UnhideAlbum(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	// if claims.Role == "user" {
	// 	return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Недостаточно прав"})
	// }

	fmt.Println(claims.Role)

	albumID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
	}

	// user, err := h.service.GetUser()

	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
	// }

	// fmt.Println(user.Role)

	// if user.Role == "admin" {
	// 	return c.JSON(http.StatusBadRequest, map[string]string{"error": "Недостаточно прав"})
	// }

	err = h.service.UnhideAlbum(claims.UserID, albumID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to unhide album"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "album unhidden"})
}

func (h *Handler) IsAlbumHidden(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	albumID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	albumHidden, err := h.service.IsAlbumHidden(claims.UserID, albumID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]bool{"commentBlocked": albumHidden})
}
