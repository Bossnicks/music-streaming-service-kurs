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
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	minioStorage, err := storage.NewMinioStorage()
	if err != nil {
		log.Fatalf("Ошибка подключения к MinIO: %v", err)
	}

	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://172.20.10.2:5173"}, // Разрешенные источники
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},                                  // Разрешенные HTTP-методы
	}))

	repo := music.NewRepository(db)
	service := music.NewService(repo)
	handler := music.NewHandler(service, minioStorage)

	e.GET("/songs/:id/info", handler.GetTrackInfo)
	e.GET("/songs/:id", handler.GetTrackPlaylist) // Эндпоинт для m3u8
	e.GET("/songs/:id/download", handler.GetMP3)  // Эндпоинт для mp3
	e.GET("/images/:id/:bucket", handler.GetImage)
	e.POST("/beatstreet/api/users/addnewplaylist", handler.AddPlaylist)
	e.PUT("/beatstreet/api/users/updateplaylist/:id", handler.UpdatePlaylist)
	e.DELETE("/beatstreet/api/users/deleteplaylist/:id", handler.DeletePlaylist)
	e.GET("/beatstreet/api/users/allplaylist", handler.GetUserPlaylists)
	e.POST("/songs/:playlistId/playlist/addsong/:trackID", handler.AddSongToPlaylist)
	e.POST("/songs/upload", handler.UploadTrack)
	e.PUT("/songs/:id", handler.UpdateTrack)
	e.DELETE("/songs/:id", handler.DeleteTrack)
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
	e.PUT("/songs/:id/commenthide", handler.HideComment)
	e.PUT("/songs/:id/commentunhide", handler.UnhideComment)
	e.PUT("/songs/:id/trackhide", handler.HideTrack)
	e.PUT("/songs/:id/trackunhide", handler.UnhideTrack)
	e.POST("/songs/:id/listens", handler.AddTrackListen)
	e.GET("/songs/:id/listens", handler.GetTrackListens)
	e.GET("/songs/:id/trackparts", handler.GetTrackPartsByTrackID)
	e.GET("/artists/top", handler.GetTopUsersByPopularity)
	e.GET("/artists/:id", handler.GetUserByID)
	e.GET("/artists/:id/tracks", handler.GetArtistTracks)
	e.GET("/playlists/:id", handler.GetPlaylist)
	e.GET("/songs/:id/statistics", handler.GetSongStatistics)
	e.GET("/globalstatistics", handler.GetTrackStatisticsGlobal)
	e.GET("/beatstreet/api/users/favoritesongs", handler.GetFavorites)
	e.GET("/playlists/recommendedByAI", handler.GetTopListenedTracks)
	e.GET("/playlists/recommendationByAI/:id", handler.GetRecommendationByAI)
	e.GET("/songs/getRecent", handler.GetRecentTracks)
	e.GET("/artist/topListenedUsers", handler.GetTopListenedUsers)
	e.GET("/neuromusic/:mood", handler.GetNeuroData)
	e.GET("/getMyWave", handler.GetMyWave)
	e.POST("/albums", handler.CreateAlbum)
	e.GET("/albums", handler.GetAlbums)
	e.GET("/albums/:id", handler.GetAlbum)
	e.DELETE("/albums/:id", handler.DeleteAlbum)
	// e.PATCH("/albums/:id/hide", handler.ToggleAlbumVisibility)
	e.GET("/tracks/available-for-album", handler.GetAvailableTracks)
	//e.GET("/playlists/addsong", )

	log.Println("Запуск music-service на порту 11000")
	if err := e.Start(":11000"); err != nil {
		log.Fatal(err)
	}
}
