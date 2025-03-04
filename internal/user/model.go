package user

import "time"

type User struct {
	ID         int        `json:"id"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	Password   string     `json:"password,omitempty"`
	Avatar     []byte     `json:"avatar"`
	Role       string     `json:"role"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
	Token      string     `json:"token"`
	CanComment bool       `json:"can_comment"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Avatar   []byte `json:"avatar"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	Avatar   *[]byte `json:"avatar"`
}

type FeedItem struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	UserName   string    `json:"user_name"`
	Type       string    `json:"type"` // "repost", "upload", "playlist"
	TargetID   int       `json:"target_id"`
	TargetName string    `json:"target_name"`
	CreatedAt  time.Time `json:"created_at"`
}
