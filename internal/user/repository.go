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
	query := "SELECT id, username, email, password, avatar, role, created_at FROM users WHERE email = $1"
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Avatar, &user.Role, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("пользователь не найден")
	}
	return &user, err
}

func (r *Repository) GetUserByID(userID int) (*User, error) {
	fmt.Println(userID)
	var user User
	query := "SELECT id, username, email, password, avatar, role, created_at, token FROM users WHERE id = $1"
	err := r.db.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Avatar, &user.Role, &user.CreatedAt, &user.Token)
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
