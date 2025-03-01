package music

import "time"

type Track struct {
	ID          int       `json:"id"`
	Artist      string    `json:"author_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Duration    int       `json:"duration"`
	Created_at  time.Time `json:"created_at"`
}

type Playlist struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Avatar string `json:"avatar"`
}
