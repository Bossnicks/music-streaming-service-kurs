package music

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetTrack(id int) (*Track, error) {
	return s.repo.GetTrackByID(id)
}
