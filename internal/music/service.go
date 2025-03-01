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
