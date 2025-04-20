package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your_secret_key")
var jwtSecretReset = []byte("your_secret_key_reset")

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type ResetClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID int, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Токен действует 24 часа
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		fmt.Println(err, "cdcd")
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("невалидный токен")
	}

	return claims, nil
}

// GenerateResetToken создает JWT для сброса пароля (1 час)
func GenerateResetToken(email string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &ResetClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretReset)
}

// ParseResetToken парсит и валидирует токен сброса пароля
func ParseResetToken(tokenString string) (*ResetClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ResetClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretReset, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*ResetClaims)
	if !ok || !token.Valid {
		return nil, errors.New("невалидный токен")
	}

	return claims, nil
}
