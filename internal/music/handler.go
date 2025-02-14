package music

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Bossnicks/music-streaming-service-kurs/pkg/storage"

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
	filename := c.Param("id") // Получаем имя файла

	// Если запрашивают ts-файл, сразу отдаем его из MinIO
	if strings.HasSuffix(filename, ".ts") {
		return h.streamFromMinIO(c, filename)
	}

	// Проверяем трек в БД
	id, err := strconv.Atoi(filename)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID"})
	}

	track, err := h.service.GetTrack(id)
	if err != nil || track == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Трек не найден в БД"})
	}

	// Получаем m3u8 из MinIO
	return h.streamFromMinIO(c, filename+".m3u8")
}

// Отдаёт файл из MinIO
func (h *Handler) streamFromMinIO(c echo.Context, filename string) error {
	obj, err := h.storage.GetFile(filename)
	if err != nil {
		return c.String(http.StatusNotFound, "Файл не найден")
	}
	defer obj.Close()

	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, obj); err != nil {
		return c.String(http.StatusInternalServerError, "Ошибка чтения файла")
	}

	return c.Blob(http.StatusOK, http.DetectContentType(buf.Bytes()), buf.Bytes())
}
