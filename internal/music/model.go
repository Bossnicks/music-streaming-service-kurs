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
	Is_blocked  bool      `json:"is_blocked"`
	Updated_at  time.Time `json:"updated_at"`
	Author      User      `json:"author"`
}

type Comment struct {
	ID        int         `json:"id"`
	Text      string      `json:"text"`
	CreatedAt time.Time   `json:"created_at"`
	Moment    int         `json:"moment"`
	IsHidden  bool        `json:"is_hidden"`
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
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Avatar      string    `json:"avatar"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Author      User      `json:"author"`
	Tracks      []Track   `json:"tracks"`
}

type TrackStatistics struct {
	TotalListens      int
	AverageListenTime float64
	MorningPercent    float64
	AfternoonPercent  float64
	EveningPercent    float64
	NightPercent      float64
	TotalLikes        int
	TotalReposts      int
	TopCountries      []string
}

type Stats struct {
	ListenCount     int     // Количество прослушиваний
	LikeCount       int     // Количество лайков
	UniqueListeners int     // Уникальные слушатели
	Engagement      float64 // Вовлеченность (лайки / прослушивания)
}
