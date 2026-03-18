package server

import (
	"Quest100Backend/internal/leaderboard/api"
	"Quest100Backend/internal/leaderboard/application"

	"github.com/gin-gonic/gin"
)

func SetupLeaderboardRoutes(r *gin.RouterGroup, lbServ application.LeaderboardService) {
	leaderboardHandler := api.NewLeaderboardHandler(lbServ)
	lbGroup := r.Group("/leaderboard")
	{

	}
}
