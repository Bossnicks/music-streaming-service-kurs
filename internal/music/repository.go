package music

import (
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/Bossnicks/music-streaming-service-kurs/pkg/errorspkg"

	"github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AddPlaylist(title, description string, userID int) (int, error) {
	var playlistID int
	query := "INSERT INTO playlists (title, description, author_id) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRow(query, title, description, userID).Scan(&playlistID)
	if err != nil {
		return 0, err
	}
	return playlistID, nil
}

func (r *Repository) UpdatePlaylist(playlistID int, title, description string, userID int) error {
	query := `UPDATE playlists SET title = $1, description = $2 WHERE id = $3 AND author_id = $4`
	_, err := r.db.Exec(query, title, description, playlistID, userID)
	return err
}

func (r *Repository) DeletePlaylist(playlistID int, userID int) error {
	query := `DELETE FROM playlists WHERE id = $1 AND author_id = $2`
	_, err := r.db.Exec(query, playlistID, userID)
	return err
}

func (r *Repository) UpdateTrack(id int, title, description, genre string, userID int) error {
	query := `
        UPDATE tracks 
        SET title = $1, 
            description = $2, 
            genre = $3, 
            updated_at = NOW() 
        WHERE id = $4 
        AND author_id = $5
    `
	_, err := r.db.Exec(query, title, description, genre, id, userID)
	return err
}

func (r *Repository) DeleteTrack(id, userID int) error {
	query := "DELETE FROM tracks WHERE id = $1 AND author_id = $2"
	_, err := r.db.Exec(query, id, userID)
	return err
}

// func (r *Repository) GetTrackByID(id int) (*Track, error) {
// 	var track Track
// 	query := "SELECT id, author_id, title, description, duration, created_at FROM tracks WHERE id = $1"
// 	err := r.db.QueryRow(query, id).Scan(&track.ID, &track.Artist, &track.Title, &track.Description, &track.Duration, &track.Created_at)
// 	fmt.Println()
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &track, nil
// }

func (r *Repository) GetTrackByID(id int) (*Track, error) {
	var track Track
	query := `
		SELECT 
			t.id, 
			t.author_id, 
			t.title, 
			t.description, 
			t.duration, 
			t.created_at, 
			u.id AS user_id, 
			u.username, 
			u.avatar 
		FROM tracks t
		JOIN users u ON t.author_id = u.id
		WHERE t.id = $1`

	// Используем QueryRow, чтобы получить данные и присвоить их в структуру
	err := r.db.QueryRow(query, id).Scan(
		&track.ID,
		&track.Artist,
		&track.Title,
		&track.Description,
		&track.Duration,
		&track.Created_at,
		&track.Author.ID,
		&track.Author.Username,
		&track.Author.Avatar,
		//&track.Author.Popularity,
	)

	// Если произошла ошибка
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &track, nil
}

func (r *Repository) GetUserPlaylists(userID int) ([]Playlist, error) {
	query := "SELECT id, title FROM playlists WHERE author_id = $1"
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []Playlist
	for rows.Next() {
		var p Playlist
		if err := rows.Scan(&p.ID, &p.Title); err != nil {
			return nil, err
		}
		playlists = append(playlists, p)
	}

	return playlists, nil
}

func (r *Repository) GetFavorites(userID int) ([]Track, error) {
	query := `
		SELECT t.id, t.title, t.description, t.duration 
		FROM tracks t
		JOIN likes l ON t.id = l.track_id
		WHERE l.user_id = $1
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var track []Track
	for rows.Next() {
		var t Track
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Duration); err != nil {
			return nil, err
		}
		track = append(track, t)
	}

	return track, nil
}

func (r *Repository) CreateTrack(title, description, genre string, authorID int) (int, error) {
	var id int
	query := "INSERT INTO tracks (author_id, title, description, genre) VALUES ($1, $2, $3, $4) RETURNING id"
	err := r.db.QueryRow(query, authorID, title, description, genre).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) AddLike(userID, trackID int) (bool, error) {
	query := "INSERT INTO likes (user_id, track_id) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	res, err := r.db.Exec(query, userID, trackID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (r *Repository) AddSongToPlaylist(playlistId, trackID int) (bool, error) {
	// Получаем максимальную позицию для данного playlist_id и track_id
	query := `SELECT COALESCE(MAX(position), 0) FROM tracks_playlists WHERE playlist_id = $1 AND track_id = $2`
	var maxPosition int
	err := r.db.QueryRow(query, playlistId, trackID).Scan(&maxPosition)
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	// Вставляем новый трек с позицией +1
	insertQuery := `
		INSERT INTO tracks_playlists (playlist_id, track_id, position)
		VALUES ($1, $2, $3)
		ON CONFLICT (playlist_id, track_id) DO NOTHING
	`
	res, err := r.db.Exec(insertQuery, playlistId, trackID, maxPosition+1)
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	// Если строка не была вставлена (из-за конфликта)
	if rowsAffected == 0 {
		return false, fmt.Errorf("конфликт: трек с таким ID уже существует в плейлисте")
	}

	return true, nil
}

// RemoveLike удаляет лайк и возвращает true, если он был удален
func (r *Repository) RemoveLike(userID, trackID int) (bool, error) {
	query := "DELETE FROM likes WHERE user_id = $1 AND track_id = $2"
	res, err := r.db.Exec(query, userID, trackID)
	if err != nil {
		return false, err
	}

	// Проверяем, была ли удалена строка
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

// GetLikeCount получает количество лайков у трека
func (r *Repository) GetLikeCount(trackID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM likes WHERE track_id = $1"
	err := r.db.QueryRow(query, trackID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// IsTrackLiked проверяет, лайкнул ли пользователь трек
func (r *Repository) IsTrackLiked(userID, trackID int) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND track_id = $2)"
	var exists bool
	err := r.db.QueryRow(query, userID, trackID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) AddRepost(userID, trackID int) (bool, error) {
	query := "INSERT INTO reposts (user_id, track_id) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	res, err := r.db.Exec(query, userID, trackID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (r *Repository) RemoveRepost(userID, trackID int) (bool, error) {
	query := "DELETE FROM reposts WHERE user_id = $1 AND track_id = $2"
	res, err := r.db.Exec(query, userID, trackID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (r *Repository) GetRepostCount(trackID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM reposts WHERE track_id = $1"
	err := r.db.QueryRow(query, trackID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) IsTrackReposted(userID, trackID int) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM reposts WHERE user_id = $1 AND track_id = $2)"
	var exists bool
	err := r.db.QueryRow(query, userID, trackID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) GetCommentsByTrackID(trackID int, isAdmin bool) ([]Comment, error) {
	var comments []Comment
	query := `
		SELECT 
			c.id AS comment_id,
			c.text AS comment_text,
			c.moment AS comment_moment,
			c.created_at AS comment_date,
			c.is_hidden,
			u.id AS user_id,
			u.username,
			u.avatar
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.track_id = $1`

	if !isAdmin {
		query += " AND c.is_hidden = false"
	}

	query += " ORDER BY c.created_at ASC"

	rows, err := r.db.Query(query, trackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.ID, &comment.Text, &comment.Moment, &comment.CreatedAt, &comment.IsHidden, &comment.User.ID, &comment.User.Username, &comment.User.Avatar)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *Repository) AddComment(trackID, userID int, text string, moment int) (int, error) {
	// Проверяем, разрешено ли пользователю оставлять комментарии
	var canComment bool
	queryCheck := `SELECT can_comment FROM users WHERE id = $1`
	err := r.db.QueryRow(queryCheck, userID).Scan(&canComment)
	if err != nil {
		return 0, err
	}

	// Если пользователю запрещено оставлять комментарии
	if !canComment {
		return 0, errorspkg.ErrCommentBanned
	}

	var id int
	query := `INSERT INTO comments (track_id, user_id, text, moment) 
	          VALUES ($1, $2, $3, $4) RETURNING id`
	err = r.db.QueryRow(query, trackID, userID, text, moment).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) AddTrackListen(listenerID int, trackID int, country string, device string, duration int, parts []TrackParts) (int, error) {
	var id int
	query := `INSERT INTO track_listens (listener_id, track_id, country, device, total_listen_time) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`

	// Если listenerID == 0, передаём nil
	var listenerIDPtr sql.NullInt32
	if listenerID == 0 {
		listenerIDPtr = sql.NullInt32{Valid: false} // NULL в БД
	} else {
		listenerIDPtr = sql.NullInt32{Int32: int32(listenerID), Valid: true}
	}
	fmt.Println("cdcdcd" + device)

	err := r.db.QueryRow(query, listenerIDPtr, trackID, country, device, duration).Scan(&id)
	if err != nil {
		return 0, err
	}

	if len(parts) > 0 {
		for _, part := range parts {
			_, err := r.db.Exec(
				`INSERT INTO listens_parts (listen_id, start_time, end_time) VALUES ($1, $2, $3)`,
				id, part.StartTime, part.EndTime,
			)
			if err != nil {
				return 0, err
			}
		}
	}
	return id, nil
}

func (r *Repository) GetTrackPartsByTrackID(trackID int) ([]TrackPartsAverage, error) {
	var parts []TrackPartsAverage
	query := `
		SELECT * FROM get_similar_gaps($1);`

	rows, err := r.db.Query(query, trackID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var part TrackPartsAverage
		err := rows.Scan(&part.StartTime, &part.EndTime, &part.Count)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		fmt.Println(part)
		parts = append(parts, part)
	}

	return parts, nil
}

func (r *Repository) GetTrackListens(trackID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM track_listens WHERE track_id = $1`
	err := r.db.QueryRow(query, trackID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetTopUsersByPopularity() ([]User, error) {
	query := `
		SELECT u.id, u.username, u.avatar, COALESCE(SUM(tl.listen_count), 0) AS popularity
		FROM users u
		LEFT JOIN tracks t ON u.id = t.author_id
		LEFT JOIN (
			SELECT track_id, COUNT(*) AS listen_count
			FROM track_listens
			GROUP BY track_id
		) tl ON t.id = tl.track_id
		GROUP BY u.id
		ORDER BY popularity DESC
		LIMIT 30;
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var popularity int
		err := rows.Scan(&user.ID, &user.Username, &user.Avatar, &popularity)
		if err != nil {
			return nil, err
		}
		user.Popularity = popularity
		users = append(users, user)
	}

	return users, nil
}

func (r *Repository) GetUserByID(userID int) (*User, error) {
	query := `
		SELECT id, username, avatar 
		FROM users 
		WHERE id = $1
	`

	var user User
	err := r.db.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.Avatar)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Пользователь не найден
		}
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetArtistTracks(artistID, page int) ([]Track, error) {
	const pageSize = 10
	offset := (page - 1) * pageSize

	query := `
		SELECT id, author_id, title, description, duration, created_at
		FROM tracks
		WHERE author_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3;
	`

	rows, err := r.db.Query(query, artistID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []Track
	for rows.Next() {
		var track Track
		err := rows.Scan(&track.ID, &track.Artist, &track.Title, &track.Description, &track.Duration, &track.Created_at)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (r *Repository) HideComment(commentID int) error {
	_, err := r.db.Exec("UPDATE comments SET is_hidden = TRUE WHERE id = $1", commentID)
	return err
}

func (r *Repository) UnhideComment(commentID int) error {
	_, err := r.db.Exec("UPDATE comments SET is_hidden = FALSE WHERE id = $1", commentID)
	return err
}

//p.avatar AS playlist_avatar,
//			u.avatar AS author_avatar,
//			t.avatar AS track_avatar,
//			t.updated_at AS track_updated_at,

func (r *Repository) GetPlaylistByID(playlistID int, isAdmin bool) (*Playlist, error) {
	query := `
		SELECT 
			p.id AS playlist_id,
			p.title AS playlist_title,
			p.description AS playlist_description,
			p.created_at AS playlist_created_at,
			p.updated_at AS playlist_updated_at,
			u.id AS author_id,
			u.username AS author_username,
			COALESCE(t.id, 0) AS track_id,
			COALESCE(t.title, '') AS track_title,
			COALESCE(t.description, '') AS track_description,
			COALESCE(t.duration, 0) AS track_duration,
			COALESCE(t.created_at, NOW()) AS track_created_at,
			COALESCE(t.is_blocked, false) AS track_is_blocked,
			COALESCE(u2.id, 0) AS track_author_id,
			COALESCE(u2.username, '') AS track_author_username
		FROM playlists p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN tracks_playlists tp ON p.id = tp.playlist_id
		LEFT JOIN tracks t ON tp.track_id = t.id
		LEFT JOIN users u2 ON t.author_id = u2.id
		WHERE p.id = $1`

	if !isAdmin {
		query += " AND (t.id IS NULL OR t.is_blocked = false)"
	}

	rows, err := r.db.Query(query, playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlist Playlist
	var tracks []Track
	firstRow := true // Флаг для первой строки

	for rows.Next() {
		var track Track
		var trackAuthor User

		err := rows.Scan(
			&playlist.ID,
			&playlist.Title,
			&playlist.Description,
			&playlist.CreatedAt,
			&playlist.UpdatedAt,
			&playlist.Author.ID,
			&playlist.Author.Username,
			&track.ID,
			&track.Title,
			&track.Description,
			&track.Duration,
			&track.Created_at,
			&track.Is_blocked,
			&trackAuthor.ID,
			&trackAuthor.Username,
		)

		if err != nil {
			fmt.Println("Scan error:", err)
			return nil, err
		}

		// Если это первая строка, то уже есть данные о плейлисте
		if firstRow {
			firstRow = false
		}

		// Если трек реально существует (id != 0), добавляем его в список
		if track.ID != 0 {
			track.Author = trackAuthor
			tracks = append(tracks, track)
		}
	}

	// Если вообще не было строк, значит плейлиста нет
	if firstRow {
		return nil, fmt.Errorf("playlist not found")
	}

	playlist.Tracks = tracks
	return &playlist, nil
}

func (r *Repository) HideTrack(commentID int) error {
	_, err := r.db.Exec("UPDATE tracks SET is_blocked = TRUE WHERE id = $1", commentID)
	return err
}

func (r *Repository) UnhideTrack(commentID int) error {
	_, err := r.db.Exec("UPDATE tracks SET is_blocked = FALSE WHERE id = $1", commentID)
	return err
}

// -- Среднее время прослушивания (в секундах)
// COALESCE(AVG(l.listen_time), 0) AS average_listen_time,

func (r *Repository) GetSongStatistics(trackID int) (*TrackStatistics, error) {
	var stats TrackStatistics

	query := `
		SELECT
			-- Общее количество прослушиваний
			COUNT(l.id) AS total_listens,



			-- Процент прослушиваний по времени суток (определяется по created_at)
			COALESCE(COUNT(CASE WHEN EXTRACT(HOUR FROM l.created_at) >= 6 AND EXTRACT(HOUR FROM l.created_at) < 12 THEN 1 END) * 100.0 / NULLIF(COUNT(l.id), 0), 0) AS morning_percent,
			COALESCE(COUNT(CASE WHEN EXTRACT(HOUR FROM l.created_at) >= 12 AND EXTRACT(HOUR FROM l.created_at) < 18 THEN 1 END) * 100.0 / NULLIF(COUNT(l.id), 0), 0) AS afternoon_percent,
			COALESCE(COUNT(CASE WHEN EXTRACT(HOUR FROM l.created_at) >= 18 AND EXTRACT(HOUR FROM l.created_at) < 24 THEN 1 END) * 100.0 / NULLIF(COUNT(l.id), 0), 0) AS evening_percent,
			COALESCE(COUNT(CASE WHEN EXTRACT(HOUR FROM l.created_at) >= 0 AND EXTRACT(HOUR FROM l.created_at) < 6 THEN 1 END) * 100.0 / NULLIF(COUNT(l.id), 0), 0) AS night_percent,

			-- Количество лайков
			(SELECT COUNT(*) FROM likes WHERE track_id = $1) AS total_likes,

			-- Количество репостов
			(SELECT COUNT(*) FROM reposts WHERE track_id = $1) AS total_reposts,

			-- Топ 5 стран
			ARRAY(
				SELECT l.country
				FROM track_listens l
				WHERE l.track_id = $1
				GROUP BY l.country
				ORDER BY COUNT(l.country) DESC
				LIMIT 5
			) AS top_countries

			FROM track_listens l
			WHERE l.track_id = $1;
				`

	row := r.db.QueryRow(query, trackID)
	err := row.Scan(
		&stats.TotalListens,
		//&stats.AverageListenTime,
		&stats.MorningPercent,
		&stats.AfternoonPercent,
		&stats.EveningPercent,
		&stats.NightPercent,
		&stats.TotalLikes,
		&stats.TotalReposts,
		pq.Array(&stats.TopCountries), // Используем pq.Array для работы с массивами в PostgreSQL
	)
	fmt.Println(err)
	if err != nil {
		return nil, fmt.Errorf("failed to get song statistics: %w", err)
	}

	return &stats, nil
}

// repository/statistics.go

func (r *Repository) GetGlobalStatistics(days int) (int, int, int, int, error) {
	fmt.Println(days)
	var listens, likes, listeners, engagement int
	fmt.Println(days)

	// Количество всех прослушиваний
	queryListens := fmt.Sprintf(`
        SELECT COUNT(*) FROM track_listens
        WHERE created_at >= NOW() - INTERVAL '%d days'
    `, days)
	if err := r.db.QueryRow(queryListens).Scan(&listens); err != nil {
		fmt.Println(err)
		return 0, 0, 0, 0, err
	}

	// Количество всех лайков
	queryLikes := fmt.Sprintf(`
        SELECT COUNT(*) FROM likes
        WHERE created_at >= NOW() - INTERVAL '%d days'
    `, days)
	if err := r.db.QueryRow(queryLikes).Scan(&likes); err != nil {
		return 0, 0, 0, 0, err
	}

	// Количество уникальных слушателей
	queryListeners := fmt.Sprintf(`
        SELECT COUNT(DISTINCT listener_id) FROM track_listens
        WHERE created_at >= NOW() - INTERVAL '%d days'
    `, days)
	if err := r.db.QueryRow(queryListeners).Scan(&listeners); err != nil {
		return 0, 0, 0, 0, err
	}

	// Подсчет вовлеченности (если есть прослушивания)
	if listens > 0 {
		engagement = (likes * 100) / listens
	} else {
		engagement = 0
	}

	return listens, likes, listeners, engagement, nil
}

func round(f float64) float64 {
	return math.Round(f*1e5) / 1e5
}

func (r *Repository) UpdateTrackFeatures(trackID int, features *AudioFeatures) error {
	query := `
		UPDATE tracks SET 
			duration_sec = $1, tempo_bpm = $2, chroma_mean = $3, rmse_mean = $4, 
			spectral_centroid = $5, spectral_bandwidth = $6, rolloff = $7, 
			zero_crossing_rate = $8
		WHERE id = $9
	`
	_, err := r.db.Exec(query,
		round(features.DurationSec),
		round(features.TempoBPM),
		round(features.ChromaMean),
		round(features.RMSEMean),
		round(features.SpectralCentroid),
		round(features.SpectralBandwidth),
		round(features.Rolloff),
		round(features.ZeroCrossingRate),
		trackID,
	)

	fmt.Println(features.DurationSec,
		features.TempoBPM,
		features.ChromaMean,
		features.RMSEMean,
		features.SpectralCentroid,
		features.SpectralBandwidth,
		features.Rolloff,
		features.ZeroCrossingRate)
	return err
}

func (r *Repository) GetTopListenedTracks(userID int) ([]Track, error) {
	timeframes := []string{
		"7 days",
		"30 days",
		"90 days",
	}

	for _, tf := range timeframes {
		query := fmt.Sprintf(`
			SELECT 
				t.id, t.author_id, t.title, t.description, t.duration, t.is_blocked, t.created_at,
				COUNT(*) AS listen_count
			FROM track_listens tl
			JOIN tracks t ON tl.track_id = t.id
			WHERE tl.listener_id = $1 AND tl.created_at >= NOW() - INTERVAL '%s'
			GROUP BY t.id
			HAVING COUNT(*) > 10
			ORDER BY listen_count DESC
		`, tf)

		rows, err := r.db.Query(query, userID)
		// остальной код без изменений

		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var tracks []Track
		for rows.Next() {
			var t Track
			var listenCount int // можно использовать если хочешь передавать ещё и счётчик
			err := rows.Scan(
				&t.ID,
				&t.Artist,
				&t.Title,
				&t.Description,
				&t.Duration,
				&t.Is_blocked,
				&t.Created_at,
				&listenCount,
			)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			tracks = append(tracks, t)
		}

		if len(tracks) >= 5 {
			// Возвращаем только 5 треков с наибольшим listen_count
			return tracks[:5], nil
		}
	}

	return []Track{}, nil
}

func (r *Repository) GetRecommendationByAI(trackID int) ([]Track, error) {
	query := `
WITH reference_track AS (
    SELECT 
        id, author_id, title, description, duration, is_blocked, created_at,
        tempo_bpm, chroma_mean, rmse_mean, spectral_centroid, 
        spectral_bandwidth, rolloff, zero_crossing_rate
    FROM tracks 
    WHERE id = $1
),
similar_tracks AS (
    SELECT 
        t.id, t.author_id, t.title, t.description, t.duration, t.is_blocked, t.created_at,
        (
            0.5 * ABS(t.tempo_bpm - r.tempo_bpm) / NULLIF(r.tempo_bpm, 0) +
            1.0 * ABS(t.chroma_mean - r.chroma_mean) / NULLIF(r.chroma_mean, 0) +
            0.8 * ABS(t.rmse_mean - r.rmse_mean) / NULLIF(r.rmse_mean, 0) +
            1.2 * ABS(t.spectral_centroid - r.spectral_centroid) / NULLIF(r.spectral_centroid, 0) +
            1.0 * ABS(t.spectral_bandwidth - r.spectral_bandwidth) / NULLIF(r.spectral_bandwidth, 0) +
            1.0 * ABS(t.rolloff - r.rolloff) / NULLIF(r.rolloff, 0) +
            0.7 * ABS(t.zero_crossing_rate - r.zero_crossing_rate) / NULLIF(r.zero_crossing_rate, 0)
        ) AS similarity_score
    FROM tracks t, reference_track r
    WHERE t.id != r.id AND t.tempo_bpm IS NOT NULL
),
reference_with_score AS (
    SELECT 
        id, author_id, title, description, duration, is_blocked, created_at,
        -1.0 AS similarity_score
    FROM reference_track
)
SELECT 
    id, author_id, title, description, duration, is_blocked, created_at
FROM (
    SELECT * FROM reference_with_score
    UNION ALL
    SELECT id, author_id, title, description, duration, is_blocked, created_at, similarity_score
    FROM similar_tracks
    ORDER BY similarity_score ASC
    LIMIT 10
) AS combined
ORDER BY similarity_score ASC;



	
		`

	rows, err := r.db.Query(query, trackID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	var tracks []Track
	for rows.Next() {
		var t Track
		err := rows.Scan(
			&t.ID,
			&t.Artist,
			&t.Title,
			&t.Description,
			&t.Duration,
			&t.Is_blocked,
			&t.Created_at,
		)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		tracks = append(tracks, t)
	}

	return tracks, nil
}

func (r *Repository) GetRecentTracks(userID int) ([]Track, error) {
	query := `
		SELECT 
			t.id, t.author_id, t.title, t.description, t.duration, t.is_blocked, t.created_at
		FROM track_listens tl
		JOIN tracks t ON tl.track_id = t.id
		WHERE tl.listener_id = $1
		ORDER BY tl.created_at DESC
		LIMIT 10
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []Track
	for rows.Next() {
		var t Track
		err := rows.Scan(
			&t.ID,
			&t.Artist,
			&t.Title,
			&t.Description,
			&t.Duration,
			&t.Is_blocked,
			&t.Created_at,
		)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tracks, nil
}

func (r *Repository) GetTopListenedUsers(userID int) ([]User, error) {
	// Интервал времени для последних 30 дней
	timeframe := "30 days"

	// SQL-запрос для получения наиболее прослушиваемых пользователей
	query := fmt.Sprintf(`
		SELECT u.id, u.username, COUNT(*) AS listen_count
		FROM track_listens tl
		JOIN tracks t ON tl.track_id = t.id
		JOIN users u ON t.author_id = u.id
		WHERE tl.listener_id = $1 AND tl.created_at >= NOW() - INTERVAL '%s'
		GROUP BY u.id
		ORDER BY listen_count DESC
		LIMIT 10;
	`, timeframe)

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var listenCount int
		err := rows.Scan(&user.ID, &user.Username, &listenCount)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if len(users) == 0 {
		return nil, nil // Если нет пользователей, возвращаем пустой срез
	}

	return users, nil
}

// `
// WITH reference_track AS (
//     SELECT
//         id, author_id, title, description, duration, is_blocked, created_at,
//         tempo_bpm, chroma_mean, rmse_mean, spectral_centroid,
//         spectral_bandwidth, rolloff, zero_crossing_rate
//     FROM tracks
//     WHERE id = $1
// ),
// similar_tracks AS (
//     SELECT
//         t.id, t.author_id, t.title, t.description, t.duration, t.is_blocked, t.created_at,
//         (
//             0.5 * ABS(t.tempo_bpm - r.tempo_bpm) / NULLIF(r.tempo_bpm, 0) +
//             1.0 * ABS(t.chroma_mean - r.chroma_mean) / NULLIF(r.chroma_mean, 0) +
//             0.8 * ABS(t.rmse_mean - r.rmse_mean) / NULLIF(r.rmse_mean, 0) +
//             1.2 * ABS(t.spectral_centroid - r.spectral_centroid) / NULLIF(r.spectral_centroid, 0) +
//             1.0 * ABS(t.spectral_bandwidth - r.spectral_bandwidth) / NULLIF(r.spectral_bandwidth, 0) +
//             1.0 * ABS(t.rolloff - r.rolloff) / NULLIF(r.rolloff, 0) +
//             0.7 * ABS(t.zero_crossing_rate - r.zero_crossing_rate) / NULLIF(r.zero_crossing_rate, 0)
//         ) AS similarity_score
//     FROM tracks t, reference_track r
//     WHERE t.id != r.id AND t.tempo_bpm IS NOT NULL
// ),
// reference_with_score AS (
//     SELECT
//         id, author_id, title, description, duration, is_blocked, created_at,
//         -1.0 AS similarity_score
//     FROM reference_track
// )
// SELECT
//     id, author_id, title, description, duration, is_blocked, created_at
// FROM (
//     SELECT * FROM reference_with_score
//     UNION ALL
//     SELECT id, author_id, title, description, duration, is_blocked, created_at, similarity_score
//     FROM similar_tracks
//     ORDER BY similarity_score ASC
//     LIMIT 10
// ) AS combined
// ORDER BY similarity_score ASC;

// 	`

func (r *Repository) GetMyWaveTracks(activity, character, mood string, userID int, excludeTrackIDs []int) ([]Track, error) {

	if excludeTrackIDs == nil {
		excludeTrackIDs = []int{}
	}
	// Сначала получаем ID треков из функции getMyWave
	rows, err := r.db.Query(
		`SELECT id, title, author_id, genre, recommendation_reason 
		FROM getMyWave($1, $2, $3, $4, $5)`,
		activity, character, mood, userID, pq.Array(excludeTrackIDs),
	)

	fmt.Println(activity, character, mood, userID, pq.Array(excludeTrackIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trackIDs []int
	// var trackData = make(map[int]struct {
	// 	title    string
	// 	authorID int
	// 	genre    string
	// 	reason   string
	// })

	for rows.Next() {
		var id, authorID int
		var title, genre, reason string
		if err := rows.Scan(&id, &title, &authorID, &genre, &reason); err != nil {
			return nil, err
		}
		trackIDs = append(trackIDs, id)
		// trackData[id] = struct {
		// 	title    string
		// 	authorID int
		// 	genre    string
		// 	reason   string
		// }{title: title, authorID: authorID, genre: genre, reason: reason}
	}

	fmt.Println(trackIDs)

	if len(trackIDs) == 0 {
		return []Track{}, nil
	}

	// Теперь получаем полную информацию о треках
	query := `
		SELECT 
			t.id, t.title, t.description, t.duration, 
			t.created_at, t.is_blocked, t.updated_at, t.genre,
			u.id, u.username
		FROM tracks t
		JOIN users u ON t.author_id = u.id
		WHERE t.id = ANY($1)
	`

	rows, err = r.db.Query(query, pq.Array(trackIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []Track
	for rows.Next() {
		var track Track
		//var author User
		if err := rows.Scan(
			&track.ID, &track.Title, &track.Description, &track.Duration,
			&track.Created_at, &track.Is_blocked, &track.Updated_at, &track.Genre,
			&track.Author.ID, &track.Author.Username,
		); err != nil {
			return nil, err
		}

		// Добавляем данные из первой выборки
		// if data, ok := trackData[track.ID]; ok {
		// 	track.Title = data.title
		// 	track.Genre = data.genre
		// 	track.RecommendationReason = data.reason
		// 	track.Author = author
		// }

		tracks = append(tracks, track)
	}

	return tracks, nil
}
