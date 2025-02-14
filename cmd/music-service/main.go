package main

import (
	"log"

	"github.com/Bossnicks/music-streaming-service-kurs/internal/music"
	"github.com/Bossnicks/music-streaming-service-kurs/pkg/database"
	"github.com/Bossnicks/music-streaming-service-kurs/pkg/storage"

	"github.com/labstack/echo/v4"
)

func main() {
	// Подключение к БД
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	// Подключение к MinIO
	minioStorage, err := storage.NewMinioStorage("localhost:9000", "minioadmin", "minioadmin", "music", false)
	if err != nil {
		log.Fatalf("Ошибка подключения к MinIO: %v", err)
	}

	e := echo.New()
	repo := music.NewRepository(db)
	service := music.NewService(repo)
	handler := music.NewHandler(service, minioStorage)

	e.GET("/music/:id", handler.GetTrackInfo)
	e.GET("/music/:id/playlist", handler.GetPlaylist) // Эндпоинт для m3u8

	log.Println("Запуск music-service на порту 11000")
	if err := e.Start(":11000"); err != nil {
		log.Fatal(err)
	}
}
