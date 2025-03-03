package music

import "time"

type Track struct {
	ID          int       `json:"id"`
	Artist      string    `json:"author_id"`
	Title       string    `json:"title"`
	Avatar      string    `json:"avatar"`
	Description string    `json:"description"`
	Duration    int       `json:"duration"`
	Created_at  time.Time `json:"created_at"`
}

type Comment struct {
	ID        int         `json:"id"`
	Text      string      `json:"text"`
	CreatedAt time.Time   `json:"created_at"`
	Moment    int         `json:"moment"`
	User      UserComment `json:"usercomment"`
}

type UserComment struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type User struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Avatar     string `json:"avatar"`
	Popularity int    `json:"popularity"`
}

type Playlist struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Avatar string `json:"avatar"`
}
