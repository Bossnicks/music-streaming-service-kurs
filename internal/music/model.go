package music

import "time"

type Track struct {
	ID                   int       `json:"id"`
	Artist               string    `json:"author_id"`
	Title                string    `json:"title"`
	Avatar               string    `json:"avatar"`
	Description          string    `json:"description"`
	Duration             int       `json:"duration"`
	Created_at           time.Time `json:"created_at"`
	Is_blocked           bool      `json:"is_blocked"`
	Updated_at           time.Time `json:"updated_at"`
	Author               User      `json:"author"`
	Genre                string    `json:"genre"`
	RecommendationReason string    `json:"recommendation_reason"`
}

type GetMyWaveRequest struct {
	Activity        string `json:"activity"`
	Character       string `json:"character"`
	Mood            string `json:"mood"`
	UserID          int    `json:"user_id"`
	ExcludeTrackIDs []int  `json:"exclude_track_ids"`
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

type TrackParts struct {
	StartTime int `json:"start_time"`
	EndTime   int `json:"end_time"`
}

type TrackPartsAverage struct {
	StartTime int `json:"avg_start"`
	EndTime   int `json:"avg_end"`
	Count     int `json:"count"`
}

type AudioFeatures struct {
	DurationSec       float64 `json:"duration_sec"`
	TempoBPM          float64 `json:"tempo_bpm"`
	ChromaMean        float64 `json:"chroma_mean"`
	RMSEMean          float64 `json:"rmse_mean"`
	SpectralCentroid  float64 `json:"spectral_centroid"`
	SpectralBandwidth float64 `json:"spectral_bandwidth"`
	Rolloff           float64 `json:"rolloff"`
	ZeroCrossingRate  float64 `json:"zero_crossing_rate"`
	//MFCC              []float64
}
