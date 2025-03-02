package main

import (
	"log"

	"github.com/Bossnicks/music-streaming-service-kurs/internal/music"
	"github.com/Bossnicks/music-streaming-service-kurs/pkg/database"
	"github.com/Bossnicks/music-streaming-service-kurs/pkg/storage"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Подключение к БД
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	// Подключение к MinIO
	minioStorage, err := storage.NewMinioStorage()
	if err != nil {
		log.Fatalf("Ошибка подключения к MinIO: %v", err)
	}

	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173"}, // Разрешенные источники
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},       // Разрешенные HTTP-методы
	}))

	repo := music.NewRepository(db)
	service := music.NewService(repo)
	handler := music.NewHandler(service, minioStorage)

	e.GET("/songs/:id/info", handler.GetTrackInfo)
	e.GET("/songs/:id", handler.GetTrackPlaylist) // Эндпоинт для m3u8
	e.POST("/beatstreet/api/users/addnewplaylist", handler.AddPlaylist)
	e.GET("/beatstreet/api/users/allplaylist", handler.GetUserPlaylists)
	e.POST("/songs/upload", handler.UploadTrack)
	e.POST("/songs/:id/likes", handler.AddLike)
	e.DELETE("/songs/:id/likes", handler.RemoveLike)
	e.GET("/songs/:id/likes", handler.GetLikeCount)
	e.GET("/songs/:id/isLiked", handler.IsTrackLiked)
	e.POST("/songs/:id/reposts", handler.AddRepost)
	e.DELETE("/songs/:id/reposts", handler.RemoveRepost)
	e.GET("/songs/:id/reposts", handler.GetRepostCount)
	e.GET("/songs/:id/isReposted", handler.IsTrackReposted)
	e.GET("/songs/:id/comments", handler.GetComments)
	e.POST("/songs/:id/comments", handler.AddComment)

	log.Println("Запуск music-service на порту 11000")
	if err := e.Start(":11000"); err != nil {
		log.Fatal(err)
	}
}
