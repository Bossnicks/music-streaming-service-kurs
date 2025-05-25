package main

import (
	"log"

	"github.com/Bossnicks/music-streaming-service-kurs/internal/user"
	"github.com/Bossnicks/music-streaming-service-kurs/pkg/database"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Подключение к БД
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://172.20.10.2:5173"}, // Разрешенные источники
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},                                  // Разрешенные HTTP-методы
	}))

	repo := user.NewRepository(db)
	service := user.NewService(repo)
	handler := user.NewHandler(service)

	e.POST("/beatstreet/api/users/signup", handler.Register)
	e.POST("/beatstreet/api/users/login", handler.Login)
	e.POST("/beatstreet/api/users/forgot-password", handler.RecoverPassword)
	e.POST("/beatstreet/api/users/verficationtoken", handler.RecoverPassword)
	e.GET("/beatstreet/api/users/isloggedin", handler.GetUser)
	e.POST("/beatstreet/api/users/follow/:id", handler.FollowUser)
	e.DELETE("/beatstreet/api/users/unfollow/:id", handler.UnfollowUser)
	e.GET("/beatstreet/api/users/followers/:id", handler.GetFollowers)
	e.GET("/beatstreet/api/users/following/:id", handler.GetFollowing)
	e.GET("/users/:id/isSubscribed", handler.IsUserSubscribed)
	e.PUT("/users/:id/blockcomments", handler.BlockComments)
	e.PUT("/users/:id/unblockcomments", handler.UnblockComments)
	e.GET("/users/:id/isCommentBlocked", handler.IsCommentAbilityBlocked)
	e.GET("/beatstreet/api/music/homepage", handler.GetUserFeed)
	e.GET("/search", handler.SearchHandler)
	e.POST("/beatstreet/api/users/reset-password", handler.ResetPassword)

	log.Println("Запуск user-service на порту 12000")
	if err := e.Start(":12000"); err != nil {
		log.Fatal(err)
	}
}
