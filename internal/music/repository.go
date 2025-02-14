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
