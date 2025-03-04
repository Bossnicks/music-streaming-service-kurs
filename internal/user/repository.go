package user

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

func (r *Repository) CreateUser(user *User) error {
	query := "INSERT INTO users (username, email, password, avatar) VALUES ($1, $2, $3, $4) RETURNING id, created_at"
	err := r.db.QueryRow(query, user.Username, user.Email, user.Password, user.Avatar).
		Scan(&user.ID, &user.CreatedAt)
	return err
}

func (r *Repository) GetUserByEmail(email string) (*User, error) {
	var user User
	query := "SELECT id, username, email, password, avatar, role, created_at, can_comment FROM users WHERE email = $1"
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Avatar, &user.Role, &user.CreatedAt, &user.CanComment)
	if err == sql.ErrNoRows {
		return nil, errors.New("пользователь не найден")
	}
	return &user, err
}

func (r *Repository) GetUserByID(userID int) (*User, error) {
	fmt.Println(userID)
	var user User
	query := "SELECT id, username, email, password, avatar, role, created_at, token, can_comment FROM users WHERE id = $1"
	err := r.db.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Avatar, &user.Role, &user.CreatedAt, &user.Token, &user.CanComment)
	fmt.Println(err)
	if err == sql.ErrNoRows {
		return nil, errors.New("пользователь не найден")
	}
	return &user, err
}

func (r *Repository) UpdateUserToken(userID int, token string) error {
	query := "UPDATE users SET token = $1 WHERE id = $2"
	_, err := r.db.Exec(query, token, userID)
	return err
}

func (r *Repository) UpdateUser(userID int, user *UpdateUserRequest) error {
	query := "UPDATE users SET username = COALESCE($1, username), email = COALESCE($2, email), password = COALESCE($3, password), avatar = COALESCE($4, avatar) WHERE id = $5"
	_, err := r.db.Exec(query, user.Username, user.Email, user.Password, user.Avatar, userID)
	return err
}

func (r *Repository) GetAvatar(userID int) ([]byte, error) {
	var avatar []byte
	query := "SELECT avatar FROM users WHERE id = $1"
	err := r.db.QueryRow(query, userID).Scan(&avatar)
	if err == sql.ErrNoRows {
		return nil, errors.New("аватар не найден")
	}
	return avatar, err
}

func (r *Repository) FollowUser(userID, followingUserID int) error {
	_, err := r.db.Exec("INSERT INTO follows (following_user_id, followed_user_id) VALUES ($1, $2)", userID, followingUserID)
	return err
}

func (r *Repository) UnfollowUser(userID, followingUserID int) error {
	_, err := r.db.Exec("DELETE FROM follows WHERE following_user_id = $1 AND followed_user_id = $2", userID, followingUserID)
	return err
}

func (r *Repository) GetFollowers(userID int) ([]int, error) {
	rows, err := r.db.Query("SELECT following_user_id FROM follows WHERE followed_user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followers []int
	for rows.Next() {
		var followerID int
		if err := rows.Scan(&followerID); err != nil {
			return nil, err
		}
		followers = append(followers, followerID)
	}

	return followers, nil
}

func (r *Repository) GetFollowing(userID int) ([]int, error) {
	rows, err := r.db.Query("SELECT followed_user_id FROM follows WHERE following_user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var following []int
	for rows.Next() {
		var followingID int
		if err := rows.Scan(&followingID); err != nil {
			return nil, err
		}
		following = append(following, followingID)
	}

	return following, nil
}

func (r *Repository) IsUserSubscribed(userID, targetID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (
		SELECT 1 FROM follows WHERE following_user_id = $1 AND followed_user_id = $2
	)`
	err := r.db.QueryRow(query, userID, targetID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Блокировка комментариев пользователя
func (r *Repository) BlockComments(userID int) error {
	_, err := r.db.Exec("UPDATE users SET can_comment = FALSE WHERE id = $1", userID)
	return err
}

// Разблокировка комментариев пользователя
func (r *Repository) UnblockComments(userID int) error {
	_, err := r.db.Exec("UPDATE users SET can_comment = TRUE WHERE id = $1", userID)
	return err
}

func (r *Repository) IsCommentAbilityBlocked(userID int) (bool, error) {
	var canComment bool
	query := `SELECT can_comment FROM users WHERE id = $1`
	err := r.db.QueryRow(query, userID).Scan(&canComment)
	if err != nil {
		return false, err
	}
	return !canComment, nil
}

func (r *Repository) GetUserFeed(userID int) ([]FeedItem, error) {
	query := `
		(
			SELECT r.id, r.user_id, u.username, 'repost' AS type, r.track_id AS target_id, t.title AS target_name, r.created_at
			FROM reposts r
			JOIN users u ON r.user_id = u.id
			JOIN tracks t ON r.track_id = t.id
			JOIN follows f ON r.user_id = f.followed_user_id
			WHERE f.following_user_id = $1
		)
		UNION ALL
		(
			SELECT t.id, t.author_id, u.username, 'upload' AS type, t.id AS target_id, t.title AS target_name, t.created_at
			FROM tracks t
			JOIN users u ON t.author_id = u.id
			JOIN follows f ON t.author_id = f.followed_user_id
			WHERE f.following_user_id = $1
		)
		UNION ALL
		(
			SELECT p.id, p.author_id, u.username, 'playlist' AS type, p.id AS target_id, p.title AS target_name, p.created_at
			FROM playlists p
			JOIN users u ON p.author_id = u.id
			JOIN follows f ON p.author_id = f.followed_user_id
			WHERE f.following_user_id = $1
		)
		ORDER BY created_at DESC
		LIMIT 50
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feed []FeedItem
	for rows.Next() {
		var item FeedItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.UserName, &item.Type, &item.TargetID, &item.TargetName, &item.CreatedAt); err != nil {
			return nil, err
		}
		feed = append(feed, item)
	}
	return feed, nil
}
