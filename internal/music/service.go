package music

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) AddPlaylist(title string, description string, userID int) (int, error) {
	return s.repo.AddPlaylist(title, description, userID)
}

func (s *Service) UpdatePlaylist(playlistID int, title, description string, userID int) error {
	return s.repo.UpdatePlaylist(playlistID, title, description, userID)
}

func (s *Service) DeletePlaylist(playlistID int, userID int) error {
	return s.repo.DeletePlaylist(playlistID, userID)
}

func (s *Service) GetTrack(id int) (*Track, error) {
	return s.repo.GetTrackByID(id)
}

func (s *Service) GetUserPlaylists(userID int) ([]Playlist, error) {
	return s.repo.GetUserPlaylists(userID)
}

func (s *Service) CreateTrack(title, description, genre string, authorID int) (int, error) {
	return s.repo.CreateTrack(title, description, genre, authorID)
}

func (s *Service) AddLike(userID, trackID int) (bool, error) {
	return s.repo.AddLike(userID, trackID)
}

func (s *Service) AddSongToPlaylist(playlistId, trackID int) (bool, error) {
	return s.repo.AddSongToPlaylist(playlistId, trackID)
}

func (s *Service) RemoveLike(userID, trackID int) (bool, error) {
	return s.repo.RemoveLike(userID, trackID)
}

func (s *Service) GetLikeCount(trackID int) (int, error) {
	return s.repo.GetLikeCount(trackID)
}

func (s *Service) GetFavorites(userID int) ([]Track, error) {
	return s.repo.GetFavorites(userID)
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

func (s *Service) GetCommentsByTrackID(trackID int, isAdmin bool) ([]Comment, error) {
	return s.repo.GetCommentsByTrackID(trackID, isAdmin)
}

func (s *Service) AddComment(trackID, userID int, text string, moment int) (int, error) {
	return s.repo.AddComment(trackID, userID, text, moment)
}

func (s *Service) AddTrackListen(listenerID int, trackID int, country string, device string, duration int, parts []TrackParts) (int, error) {
	return s.repo.AddTrackListen(listenerID, trackID, country, device, duration, parts)
}

func (s *Service) GetTrackPartsByTrackID(trackID int) ([]TrackPartsAverage, error) {
	return s.repo.GetTrackPartsByTrackID(trackID)
}

func (s *Service) GetTrackListens(trackID int) (int, error) {
	return s.repo.GetTrackListens(trackID)
}

func (s *Service) GetTopUsersByPopularity() ([]User, error) {
	return s.repo.GetTopUsersByPopularity()
}

func (s *Service) GetUser(userID int) (*User, error) {
	return s.repo.GetUserByID(userID)
}

func (s *Service) GetArtistTracks(artistID, page int) ([]Track, error) {
	return s.repo.GetArtistTracks(artistID, page)
}

func (s *Service) HideComment(commentID int) error {
	return s.repo.HideComment(commentID)
}

func (s *Service) UnhideComment(commentID int) error {
	return s.repo.UnhideComment(commentID)
}

func (s *Service) GetPlaylistByID(playlistID int, isAdmin bool) (*Playlist, error) {
	return s.repo.GetPlaylistByID(playlistID, isAdmin)
}

func (s *Service) HideTrack(commentID int) error {
	return s.repo.HideTrack(commentID)
}

func (s *Service) UnhideTrack(commentID int) error {
	return s.repo.UnhideTrack(commentID)
}

func (s *Service) GetSongStatistics(trackID int) (*TrackStatistics, error) {
	return s.repo.GetSongStatistics(trackID)
}

// service/statistics.go

func (s *Service) GetGlobalStatistics(days int) (int, int, int, int, error) {
	return s.repo.GetGlobalStatistics(days)
}

func (s *Service) UpdateTrackFeatures(trackID int, features *AudioFeatures) error {
	return s.repo.UpdateTrackFeatures(trackID, features)
}

func (s *Service) GetTopListenedTracks(userID int) ([]Track, error) {
	return s.repo.GetTopListenedTracks(userID)
}

func (s *Service) GetRecommendationByAI(trackID int) ([]Track, error) {
	return s.repo.GetRecommendationByAI(trackID)
}

func (s *Service) GetRecentTracks(userID int) ([]Track, error) {
	return s.repo.GetRecentTracks(userID)
}

func (s *Service) GetTopListenedUsers(userID int) ([]User, error) {
	return s.repo.GetTopListenedUsers(userID)
}

func (s *Service) GetMyWave(activity string, character string, mood string, userId int, excludeTrackIDs []int) ([]Track, error) {
	return s.repo.GetMyWaveTracks(activity, character, mood, userId, excludeTrackIDs)
}

func (s *Service) UpdateTrack(id int, title, description, genre string, userID int) error {
	return s.repo.UpdateTrack(id, title, description, genre, userID)
}

func (s *Service) DeleteTrack(id, userID int) error {
	return s.repo.DeleteTrack(id, userID)
}
