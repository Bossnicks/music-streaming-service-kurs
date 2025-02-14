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
		AllowOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"}, // Разрешенные источники
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},       // Разрешенные HTTP-методы
	}))

	repo := music.NewRepository(db)
	service := music.NewService(repo)
	handler := music.NewHandler(service, minioStorage)

	e.GET("/music/:id/info", handler.GetTrackInfo)
	e.GET("/music/:id", handler.GetPlaylist) // Эндпоинт для m3u8

	log.Println("Запуск music-service на порту 11000")
	if err := e.Start(":11000"); err != nil {
		log.Fatal(err)
	}
}
