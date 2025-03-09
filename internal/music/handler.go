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

	// Получение данных из формы
	title := c.FormValue("title")
	description := c.FormValue("description")

	if title == "" {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Название плейлиста обязательно"})
	}

	// Загрузка обложки (если есть)

	playlistID, err := h.service.AddPlaylist(title, description, userID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании плейлиста"})
	}

	file, err := c.FormFile("cover")
	if err == nil {
		src, err := file.Open()
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при загрузке обложки"})
		}
		defer src.Close()
		fileExtension := filepath.Ext(file.Filename)

		err = h.storage.UploadImage("playlist", fmt.Sprintf("%d%s", playlistID, fileExtension), src)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка загрузки m3u8"})
		}
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

func (h *Handler) GetImage(c echo.Context) error {
	id := c.Param("id") // Получаем id
	bucketType := c.Param("bucket")

	// Попытка получить изображение с расширением .jpg
	filename := fmt.Sprintf("%s.jpg", id)
	obj, err := h.storage.GetImage(bucketType, filename)
	fmt.Println(err)
	if err == nil {
		defer obj.Close()
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, obj); err != nil {
			return c.String(http.StatusInternalServerError, "Ошибка чтения файла")
		}
		return c.Blob(http.StatusOK, http.DetectContentType(buf.Bytes()), buf.Bytes())
	}

	// Если файл с расширением .jpg не найден, пробуем .png
	fmt.Println(err)

	filename = fmt.Sprintf("%s.png", id)
	obj, err = h.storage.GetImage(bucketType, filename)
	if err != nil {
		return c.String(http.StatusNotFound, "Изображение не найдено")
	}

	defer obj.Close()
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, obj); err != nil {
		return c.String(http.StatusInternalServerError, "Ошибка чтения файла")
	}

	return c.Blob(http.StatusOK, http.DetectContentType(buf.Bytes()), buf.Bytes())
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
	genre := c.FormValue("genre")

	cover, _ := c.FormFile("cover")

	//fmt.Println(file)

	// Открываем файл
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка открытия файла"})
	}
	defer src.Close()

	pic, err := cover.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка открытия файла"})
	}
	defer pic.Close()

	// Сохраняем в БД и получаем ID
	trackID, err := h.service.CreateTrack(title, description, genre, claims.UserID)
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
	fileExtension := filepath.Ext(cover.Filename)

	err = h.storage.UploadImage("track", fmt.Sprintf("%d%s", trackID, fileExtension), pic)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка загрузки m3u8"})
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

func (h *Handler) AddSongToPlaylist(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен отсутствует"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	_, err := auth.ParseJWT(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
	}

	trackID, err := strconv.Atoi(c.Param("trackID"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID трека"})
	}

	playlistId, err := strconv.Atoi(c.Param("playlistId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID плейлиста"})
	}

	_, err = h.service.AddSongToPlaylist(playlistId, trackID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка добавления в плейлист"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Успешно добавлено"})
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

func (h *Handler) GetFavorites(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")

	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Нет токена"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)

	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Токен неверный"})
	}

	favorites, err := h.service.GetFavorites(claims.UserID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch comments"})
	}

	return c.JSON(http.StatusOK, favorites)
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

	var req struct {
		SongId  int
		Country string
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат запроса"})
	}

	fmt.Println(req.Country)
	fmt.Println(req.SongId)

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

	// ip, err := network.GetPublicIP()

	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to determine ip"})
	// }
	// country, err := network.GetCountryByIP(ip)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to determine country"})
	// }

	fmt.Println("Country detected:", req.Country)

	listenerIDValue := 0
	if listenerID != nil {
		listenerIDValue = *listenerID
	}

	if listenerID == nil && authHeader == "" {
		listenerIDValue = 0
	}

	id, err := h.service.AddTrackListen(listenerIDValue, trackID, req.Country)
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

func (h *Handler) HideComment(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")

	if authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ParseJWT(tokenString)
		if err != nil || claims.Role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Недостаточно прав"})
		}
	}
	commentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid comment ID"})
	}

	err = h.service.HideComment(commentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hide comment"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Comment hidden"})
}

func (h *Handler) UnhideComment(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")

	if authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ParseJWT(tokenString)
		if err != nil || claims.Role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Недостаточно прав"})
		}
	}
	commentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid comment ID"})
	}

	err = h.service.UnhideComment(commentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unhide comment"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Comment unhidden"})
}

func (h *Handler) GetPlaylist(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	isAdmin := false

	if authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ParseJWT(tokenString)
		if err == nil && claims.Role == "admin" {
			isAdmin = true
		}
	}
	playlistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid playlist ID"})
	}

	playlist, err := h.service.GetPlaylistByID(playlistID, isAdmin)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch playlist"})
	}

	return c.JSON(http.StatusOK, playlist)
}

func (h *Handler) HideTrack(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")

	if authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ParseJWT(tokenString)
		if err != nil || claims.Role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Недостаточно прав"})
		}
	}
	commentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid track ID"})
	}

	err = h.service.HideTrack(commentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hide track"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Track hidden"})
}

func (h *Handler) UnhideTrack(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")

	if authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ParseJWT(tokenString)
		if err != nil || claims.Role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Недостаточно прав"})
		}
	}
	commentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid track ID"})
	}

	err = h.service.UnhideTrack(commentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unhide track"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Track unhidden"})
}

func (h *Handler) GetSongStatistics(c echo.Context) error {
	// Получаем заголовок авторизации
	// authHeader := c.Request().Header.Get("Authorization")
	// isAdmin := false

	// // Проверка на роль администратора
	// if authHeader != "" {
	// 	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	// 	claims, err := auth.ParseJWT(tokenString)
	// 	if err == nil && claims.Role == "admin" {
	// 		isAdmin = true
	// 	}
	// }

	// Получаем trackID из параметров
	trackID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid track ID"})
	}

	// Получаем статистику для трека
	stats, err := h.service.GetSongStatistics(trackID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch song statistics"})
	}

	// Отправляем статистику
	return c.JSON(http.StatusOK, stats)
}

// handler.go

func (h *Handler) GetTrackStatisticsGlobal(c echo.Context) error {
	// Проверка авторизации
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Недостаточно прав"})
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ParseJWT(tokenString)
	if err != nil || claims.Role != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Недостаточно прав"})

	}

	// Извлекаем параметр days
	daysParam := c.QueryParam("days")
	if daysParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Days parameter is required"})
	}

	// Преобразуем days в int
	days, err := strconv.Atoi(daysParam)
	if err != nil || (days != 1 && days != 2 && days != 3) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid days parameter. Allowed values: 1, 2, 3"})
	}

	// Запрашиваем статистику
	listens, likes, listeners, engagement, err := h.service.GetGlobalStatistics(days)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve statistics"})
	}

	// Формируем ответ
	stats := map[string]int{
		"listens":    listens,
		"likes":      likes,
		"listeners":  listeners,
		"engagement": engagement,
	}

	return c.JSON(http.StatusOK, stats)
}
