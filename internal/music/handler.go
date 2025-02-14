package music

import (
	"io"
	"net/http"
	"strconv"

	"your_project/storage"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
	storage *storage.MinioStorage
}

func NewHandler(service *Service, storage *storage.MinioStorage) *Handler {
	return &Handler{service: service, storage: storage}
}

func (h *Handler) GetTrackInfo(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID"})
	}

	track, err := h.service.GetTrack(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка сервера"})
	}
	if track == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Трек не найден"})
	}

	return c.JSON(http.StatusOK, track)
}

// GetPlaylist отдает m3u8 файл
func (h *Handler) GetPlaylist(c echo.Context) error {
	id := c.Param("id")

	// Проверяем, существует ли трек в БД
	track, err := h.service.GetTrack(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка сервера"})
	}
	if track == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Трек не найден"})
	}

	// Получаем m3u8 из MinIO
	fileName := id + ".m3u8"
	obj, err := h.storage.GetFile(fileName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения файла"})
	}
	defer obj.Close()

	c.Response().Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	_, err = io.Copy(c.Response().Writer, obj)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при передаче файла"})
	}

	return nil
}
