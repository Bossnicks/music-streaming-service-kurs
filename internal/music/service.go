package music

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) AddPlaylist(title string, avatar string, userID int) (int, error) {
	return s.repo.AddPlaylist(title, avatar, userID)
}

func (s *Service) GetTrack(id int) (*Track, error) {
	return s.repo.GetTrackByID(id)
}

func (s *Service) GetUserPlaylists(userID int) ([]Playlist, error) {
	return s.repo.GetUserPlaylists(userID)
}

func (s *Service) CreateTrack(title, description string, authorID int) (int, error) {
	return s.repo.CreateTrack(title, description, authorID)
}

func (s *Service) AddLike(userID, trackID int) (bool, error) {
	return s.repo.AddLike(userID, trackID)
}

func (s *Service) RemoveLike(userID, trackID int) (bool, error) {
	return s.repo.RemoveLike(userID, trackID)
}

func (s *Service) GetLikeCount(trackID int) (int, error) {
	return s.repo.GetLikeCount(trackID)
}

func (s *Service) IsTrackLiked(userID, trackID int) (bool, error) {
	return s.repo.IsTrackLiked(userID, trackID)
}

func (s *Service) AddRepost(userID, trackID int) (bool, error) {
	return s.repo.AddRepost(userID, trackID)
}

func (s *Service) RemoveRepost(userID, trackID int) (bool, error) {
	return s.repo.RemoveRepost(userID, trackID)
}

func (s *Service) GetRepostCount(trackID int) (int, error) {
	return s.repo.GetRepostCount(trackID)
}

func (s *Service) IsTrackReposted(userID, trackID int) (bool, error) {
	return s.repo.IsTrackReposted(userID, trackID)
}
