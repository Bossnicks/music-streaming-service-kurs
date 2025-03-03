package user

import (
	"errors"

	"github.com/Bossnicks/music-streaming-service-kurs/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RegisterUser(user *User) error {
	checkuser, _ := s.repo.GetUserByEmail(user.Email)
	if checkuser != nil {
		return errors.New("пользователь уже существует")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return s.repo.CreateUser(user)
}

func (s *Service) Authenticate(email, password string) (string, *User, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return "", nil, errors.New("неверные учетные данные")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", nil, errors.New("неверные учетные данные")
	}

	token, err := auth.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return "", nil, err
	}

	//s.repo.UpdateUserToken(user.ID, token)

	err = s.repo.UpdateUserToken(user.ID, token)
	if err != nil {
		return "", nil, err
	}

	sanitizedUser := &User{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Password:  "",
		Avatar:    user.Avatar,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: nil,
	}

	return token, sanitizedUser, nil // Возвращаем и токен, и пользователя
}

func (s *Service) UpdateUserToken(userID int, newToken string) error {
	return s.repo.UpdateUserToken(userID, newToken)
}

func (s *Service) GetUser(userID int) (*User, error) {
	return s.repo.GetUserByID(userID)
}

func (s *Service) UpdateUser(userID int, req UpdateUserRequest) error {
	return s.repo.UpdateUser(userID, &req)
}

func (s *Service) GetAvatar(userID int) ([]byte, error) {
	return s.repo.GetAvatar(userID)
}

func (s *Service) FollowUser(userID, followingUserID int) error {
	if userID == followingUserID {
		return errors.New("you cannot follow yourself")
	}
	return s.repo.FollowUser(userID, followingUserID)
}

func (s *Service) UnfollowUser(userID, followingUserID int) error {
	return s.repo.UnfollowUser(userID, followingUserID)
}

func (s *Service) GetFollowers(userID int) ([]int, error) {
	return s.repo.GetFollowers(userID)
}

func (s *Service) GetFollowing(userID int) ([]int, error) {
	return s.repo.GetFollowing(userID)
}

func (s *Service) IsUserSubscribed(userID, targetID int) (bool, error) {
	return s.repo.IsUserSubscribed(userID, targetID)
}

func (s *Service) BlockComments(userID int) error {
	return s.repo.BlockComments(userID)
}

func (s *Service) UnblockComments(userID int) error {
	return s.repo.UnblockComments(userID)
}

func (s *Service) IsCommentAbilityBlocked(userID int) (bool, error) {
	return s.repo.IsCommentAbilityBlocked(userID)
}
