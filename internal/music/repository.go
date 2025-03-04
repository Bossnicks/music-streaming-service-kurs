package music

import (
	"database/sql"
	"errors"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AddPlaylist(title string, avatar string, userID int) (int, error) {
	var playlistID int
	query := "INSERT INTO playlists (title, avatar, author_id) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRow(query, title, avatar, userID).Scan(&playlistID)
	if err != nil {
		return 0, err
	}
	return playlistID, nil
}

func (r *Repository) GetTrackByID(id int) (*Track, error) {
	var track Track
	query := "SELECT id, author_id, title, description, duration, created_at FROM tracks WHERE id = $1"
	err := r.db.QueryRow(query, id).Scan(&track.ID, &track.Artist, &track.Title, &track.Description, &track.Duration, &track.Created_at)
	fmt.Println()
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

func (r *Repository) CreateTrack(title, description string, authorID int) (int, error) {
	var id int
	query := "INSERT INTO tracks (author_id, title, description) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRow(query, authorID, title, description).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// AddLike добавляет лайк к треку и возвращает true, если лайк был добавлен
func (r *Repository) AddLike(userID, trackID int) (bool, error) {
	query := "INSERT INTO likes (user_id, track_id) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	res, err := r.db.Exec(query, userID, trackID)
	if err != nil {
		return false, err
	}

	// Проверяем, была ли вставлена новая строка
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
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
	var id int
	query := `INSERT INTO comments (track_id, user_id, text, moment) 
	          VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(query, trackID, userID, text, moment).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) AddTrackListen(listenerID int, trackID int, country string) (int, error) {
	var id int
	query := `INSERT INTO track_listens (listener_id, track_id, country) 
	          VALUES ($1, $2, $3) RETURNING id`

	// Если listenerID == 0, передаём nil
	var listenerIDPtr sql.NullInt32
	if listenerID == 0 {
		listenerIDPtr = sql.NullInt32{Valid: false} // NULL в БД
	} else {
		listenerIDPtr = sql.NullInt32{Int32: int32(listenerID), Valid: true}
	}

	err := r.db.QueryRow(query, listenerIDPtr, trackID, country).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
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

func (r *Repository) GetPlaylistByID(playlistID int) (*Playlist, error) {
	query := `
		SELECT 
			p.id AS playlist_id,
			p.title AS playlist_title,
			p.description AS playlist_description,
			p.created_at AS playlist_created_at,
			p.updated_at AS playlist_updated_at,
			u.id AS author_id,
			u.username AS author_username,
			t.id AS track_id,
			t.title AS track_title,
			t.description AS track_description,
			t.duration AS track_duration,
			t.created_at AS track_created_at,
			u2.id AS track_author_id,
			u2.username AS track_author_username
		FROM playlists p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN tracks_playlists tp ON p.id = tp.playlist_id
		LEFT JOIN tracks t ON tp.track_id = t.id
		LEFT JOIN users u2 ON t.author_id = u2.id
		WHERE p.id = $1`

	rows, err := r.db.Query(query, playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlist Playlist
	var tracks []Track

	for rows.Next() {
		var track Track
		var trackAuthor User

		err := rows.Scan(
			&playlist.ID,
			&playlist.Title,
			&playlist.Description,
			//&playlist.Avatar,
			&playlist.CreatedAt,
			&playlist.UpdatedAt,
			&playlist.Author.ID,
			&playlist.Author.Username,
			//&playlist.Author.Avatar,
			&track.ID,
			&track.Title,
			&track.Description,
			//&track.Avatar,
			&track.Duration,
			&track.Created_at,
			//&track.Updated_at,
			&trackAuthor.ID,
			&trackAuthor.Username,
		)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		//playlist.Avatar = nil
		//playlist.Author.Avatar = nil
		//track.Avatar = nil

		track.Author = trackAuthor
		tracks = append(tracks, track)
	}

	playlist.Tracks = tracks
	return &playlist, nil
}
