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
	query := "SELECT id, title, avatar FROM playlists WHERE author_id = $1"
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []Playlist
	for rows.Next() {
		var p Playlist
		if err := rows.Scan(&p.ID, &p.Title, &p.Avatar); err != nil {
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

func (r *Repository) GetCommentsByTrackID(trackID int) ([]Comment, error) {
	var comments []Comment
	query := `
		SELECT 
			c.id AS comment_id,
			c.text AS comment_text,
			c.moment AS comment_moment,
			c.created_at AS comment_date,
			u.id AS user_id,
			u.username,
			u.avatar
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.track_id = $1 
		  AND (c.is_hidden IS NULL OR c.is_hidden = false)
		ORDER BY c.created_at ASC`

	rows, err := r.db.Query(query, trackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.ID, &comment.Text, &comment.Moment, &comment.CreatedAt, &comment.User.ID, &comment.User.Username, &comment.User.Avatar)
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
