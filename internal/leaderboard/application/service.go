package application

import (
	"Quest100Backend/internal/leaderboard/domain"
)

type LeaderboardService interface {
}

type leaderboardService struct {
	leaderboardRepo domain.LeaderboardRepository
}

func NewLeaderboardService(leaderboardRepo domain.LeaderboardRepository) LeaderboardService {
	return &leaderboardService{
		leaderboardRepo: leaderboardRepo,
	}
}
