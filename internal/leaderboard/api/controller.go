package api

import (
	"Quest100Backend/internal/leaderboard/application"
)

type LeaderboardHandler struct {
	leaderboardService application.LeaderboardService
}

func NewLeaderboardHandler(leaderboardService application.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardService: leaderboardService,
	}
}
