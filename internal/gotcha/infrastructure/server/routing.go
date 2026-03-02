package server

import (
	gotchaAPI "Quest100Backend/internal/gotcha/api"
	"Quest100Backend/internal/gotcha/application"
	gotchaDB "Quest100Backend/internal/gotcha/infrastructure/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupGotchaRoutes(r *gin.RouterGroup, db *gorm.DB) {
	gameRepo := gotchaDB.NewGameRepository(db)
	participantRepo := gotchaDB.NewParticipantRepository(db)
	killRepo := gotchaDB.NewKillRepository(db)
	propRepo := gotchaDB.NewPropRepository(db)

	service := application.NewGotchaService(gameRepo, participantRepo, killRepo, propRepo)
	handler := gotchaAPI.NewGotchaHandler(service)

	g := r.Group("/gotcha")
	{
		g.POST("/opt-in", handler.OptIn)
		g.POST("/kills", handler.SubmitKill)
		g.POST("/kills/:killId/like", handler.LikeKill)
		g.DELETE("/kills/:killId/like", handler.UnlikeKill)
		g.GET("/feed", handler.GetFeed)
		g.GET("/me", handler.GetMyStatus)
		g.GET("/leaderboard", handler.GetLeaderboard)

		g.POST("/games", handler.CreateGame)
		g.PUT("/kills/:killId/review", handler.ReviewKill)
		g.GET("/kills/pending", handler.GetPendingKills)
	}
}
