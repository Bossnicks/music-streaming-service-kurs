package music

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Bossnicks/music-streaming-service-kurs/pkg/auth"
	"github.com/Bossnicks/music-streaming-service-kurs/pkg/network"

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

func (h *Handler) AddPlaylist(c echo.Context) error {

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

	var req struct {
		Title  string `json:"title"`
		Avatar string `json:"avatar"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректные данные"})
	}

	playlistID, err := h.service.AddPlaylist(req.Title, req.Avatar, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании плейлиста"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message":     "Плейлист создан",
		"playlist_id": playlistID,
	})
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
func (h *Handler) GetTrackPlaylist(c echo.Context) error {
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

func (h *Handler) GetUserPlaylists(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	// if claims.Role != "user" {
	// 	return c.JSON(http.StatusForbidden, map[string]string{"error": "Недостаточно прав"})
	// }

	playlists, err := h.service.GetUserPlaylists(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения плейлистов"})
	}

	return c.JSON(http.StatusOK, playlists)
}

func (h *Handler) UploadTrack(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}
	//fmt.Println(title)

	// Получаем данные из формы
	file, err := c.FormFile("song")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Файл не найден"})
	}
	title := c.FormValue("title")
	description := c.FormValue("description")

	fmt.Println(file)

	// Открываем файл
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка открытия файла"})
	}
	defer src.Close()

	// Сохраняем в БД и получаем ID
	trackID, err := h.service.CreateTrack(title, description, claims.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка сохранения в БД"})
	}

	if _, err := os.Stat("/tmp"); os.IsNotExist(err) {
		if err := os.Mkdir("/tmp", 0755); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка создания /tmp"})
		}
	}

	// Сохраняем файл
	tmpPath := fmt.Sprintf("/tmp/%d.mp3", trackID)
	dst, err := os.Create(tmpPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка создания временного файла"})
	}
	defer dst.Close()
	io.Copy(dst, src)

	// Обработка HLS
	m3u8Path := fmt.Sprintf("/tmp/%d.m3u8", trackID)
	segmentPath := fmt.Sprintf("/tmp/%d_%%d.ts", trackID)
	fmt.Println(tmpPath, m3u8Path, segmentPath)
	cmd := exec.Command("ffmpeg", "-i", tmpPath, "-vn", "-c:a", "aac", "-b:a", "128k", "-f", "hls",
		"-hls_time", "10", "-hls_list_size", "0", "-hls_segment_filename", segmentPath, m3u8Path)

	// cmd := exec.Command("ffmpeg", "-i", tmpPath, "-c:a", "aac", "-b:a", "128k", "-f", "hls",
	// 	"-hls_time", "10", "-hls_list_size", "0", "-hls_segment_filename", segmentPath,
	// 	"-hls_flags", "split_by_time", m3u8Path)
	// cmd := exec.Command("ffmpeg", "-i", tmpPath, "-c:a", "aac", "-b:a", "128k", "-f", "hls",
	// 	"-hls_time", "10", "-hls_list_size", "0", "-hls_segment_filename", segmentPath,
	// 	"-hls_flags", "split_by_time+temp_file", m3u8Path)

	// cmd := exec.Command("ffmpeg", "-i", tmpPath, "-c:a", "aac", "-b:a", "128k", "-f", "hls",
	// 	"-hls_time", "10", "-hls_list_size", "0", "-hls_segment_filename", segmentPath,
	// 	"-hls_flags", "split_by_time+temp_file", "-force_key_frames", "expr:gte(t,n_forced*10)", m3u8Path)

	cmd.Dir = "/tmp"
	fmt.Println(os.Getwd())
	if err := cmd.Run(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка обработки FFmpeg"})
	}

	// Загружаем в MinIO
	err = h.storage.UploadFile(fmt.Sprintf("%d.m3u8", trackID), m3u8Path)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка загрузки m3u8"})
	}

	segmentFiles, _ := filepath.Glob(fmt.Sprintf("/tmp/%d_*.ts", trackID))
	for _, segment := range segmentFiles {
		if err := h.storage.UploadFile(filepath.Base(segment), segment); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка загрузки сегментов"})
		}
	}

	// Удаляем временные файлы
	os.Remove(tmpPath)
	os.Remove(m3u8Path)
	for _, segment := range segmentFiles {
		os.Remove(segment)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Трек загружен", "id": fmt.Sprintf("%d", trackID)})
}

func (h *Handler) AddLike(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	trackID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID трека"})
	}

	liked, err := h.service.AddLike(claims.UserID, trackID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка добавления лайка"})
	}

	return c.JSON(http.StatusOK, map[string]bool{"liked": liked})
}

func (h *Handler) RemoveLike(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	trackID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID трека"})
	}

	removed, err := h.service.RemoveLike(claims.UserID, trackID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка удаления лайка"})
	}

	return c.JSON(http.StatusOK, map[string]bool{"removed": removed})
}

func (h *Handler) GetLikeCount(c echo.Context) error {
	trackID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID трека"})
	}

	count, err := h.service.GetLikeCount(trackID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения количества лайков"})
	}

	return c.JSON(http.StatusOK, map[string]int{"likes": count})
}

func (h *Handler) IsTrackLiked(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	trackID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID трека"})
	}

	liked, err := h.service.IsTrackLiked(claims.UserID, trackID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка проверки лайка"})
	}

	return c.JSON(http.StatusOK, map[string]bool{"liked": liked})
}

func (h *Handler) AddRepost(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	trackID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID трека"})
	}

	reposted, err := h.service.AddRepost(claims.UserID, trackID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка добавления репоста"})
	}

	return c.JSON(http.StatusOK, map[string]bool{"reposted": reposted})
}

func (h *Handler) RemoveRepost(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	trackID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID трека"})
	}

	removed, err := h.service.RemoveRepost(claims.UserID, trackID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка удаления репоста"})
	}

	return c.JSON(http.StatusOK, map[string]bool{"removed": removed})
}

func (h *Handler) GetRepostCount(c echo.Context) error {
	trackID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID трека"})
	}

	count, err := h.service.GetRepostCount(trackID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения количества репостов"})
	}

	return c.JSON(http.StatusOK, map[string]int{"reposts": count})
}

func (h *Handler) IsTrackReposted(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	trackID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID трека"})
	}

	reposted, err := h.service.IsTrackReposted(claims.UserID, trackID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка проверки репоста"})
	}

	return c.JSON(http.StatusOK, map[string]bool{"reposted": reposted})
}

func (h *Handler) GetComments(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	isAdmin := false

	if authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ParseJWT(tokenString)
		if err == nil && claims.Role == "admin" {
			isAdmin = true
		}
	}

	trackID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid track ID"})
	}

	comments, err := h.service.GetCommentsByTrackID(trackID, isAdmin)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch comments"})
	}

	return c.JSON(http.StatusOK, comments)
}

func (h *Handler) AddComment(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	trackID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid track ID"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	var req struct {
		Text   string `json:"text"`
		Moment int    `json:"moment"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат запроса"})
	}

	commentID, err := h.service.AddComment(trackID, claims.UserID, req.Text, req.Moment)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка сохранения комментария"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "Комментарий добавлен",
		"comment_id": commentID,
	})
}

func (h *Handler) AddTrackListen(c echo.Context) error {
	trackID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid track ID"})
	}

	var listenerID *int

	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		listenerID = nil
	} else {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ParseJWT(tokenString)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
		}

		listenerID = &claims.UserID

	}

	ip, err := network.GetPublicIP()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to determine ip"})
	}
	country, err := network.GetCountryByIP(ip)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to determine country"})
	}

	fmt.Println("Country detected:", country)

	listenerIDValue := 0
	if listenerID != nil {
		listenerIDValue = *listenerID
	}

	if listenerID == nil && authHeader == "" {
		listenerIDValue = 0
	}

	id, err := h.service.AddTrackListen(listenerIDValue, trackID, country)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add track listen"})
	}

	return c.JSON(http.StatusOK, map[string]int{"listen_id": id})
}

func (h *Handler) GetTrackListens(c echo.Context) error {
	trackID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid track ID"})
	}

	count, err := h.service.GetTrackListens(trackID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get track listens"})
	}
	return c.JSON(http.StatusOK, map[string]int{"track_listens": count})
}

func (h *Handler) GetTopUsersByPopularity(c echo.Context) error {
	// trackID, err := strconv.Atoi(c.Param("id"))
	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid track ID"})
	// }
	fmt.Println("DSC")

	authors, err := h.service.GetTopUsersByPopularity()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get track listens"})
	}
	return c.JSON(http.StatusOK, authors)
}

func (h *Handler) GetUserByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	user, err := h.service.GetUser(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user"})
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return c.JSON(http.StatusOK, user)
}

func (h *Handler) GetArtistTracks(c echo.Context) error {
	artistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid artist ID"})
	}

	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}

	songs, err := h.service.GetArtistTracks(artistID, page)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch songs"})
	}

	if len(songs) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No songs found"})
	}

	return c.JSON(http.StatusOK, songs)
}
