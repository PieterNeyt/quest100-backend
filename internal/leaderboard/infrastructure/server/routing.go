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
		lbGroup.GET("/courses", leaderboardHandler.GetAllCoursesWithClasses)
		lbGroup.GET("", leaderboardHandler.GetAllLeaderboards)
		lbGroup.GET("/course/:courseId", leaderboardHandler.GetLeaderboardByCourseId)
		lbGroup.GET("/:id", leaderboardHandler.GetLeaderboardByID)
		lbGroup.POST("", leaderboardHandler.CreateLeaderboard)
		lbGroup.PUT("/:id", leaderboardHandler.UpdateLeaderboard)
	}
}
