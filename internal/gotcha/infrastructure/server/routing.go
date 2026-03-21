package server

import (
	gotchaAPI "Quest100Backend/internal/gotcha/api"
	"Quest100Backend/internal/gotcha/application"
	"Quest100Backend/internal/infrastructure/auth"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupGotchaRoutes(r *gin.RouterGroup, service application.GotchaService) {
	handler := gotchaAPI.NewGotchaHandler(service)

	// auto-start games + process timeouts
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := service.CheckAndStartGames(); err != nil {
				log.Printf("CheckAndStartGames error: %v", err)
			}
			if err := service.ProcessTimeouts(); err != nil {
				log.Printf("ProcessTimeouts error: %v", err)
			}
		}
	}()

	g := r.Group("/gotcha")
	{
		// Game
		g.GET("/game", handler.GetCurrentGame)
		g.POST("/games", auth.RequireRole(auth.Lector), handler.CreateGame)

		// History
		g.GET("/games/history", handler.GetGameHistory)
		g.GET("/games/:gameId/end-screen", handler.GetEndScreenByID)

		// Participation
		g.POST("/opt-in", handler.OptIn)
		g.DELETE("/opt-in", handler.OptOut)
		g.GET("/me", handler.GetMyStatus)
		g.GET("/me/target", handler.GetTargetInfo)

		// Kills
		g.POST("/kills", handler.SubmitKill)
		g.PUT("/kills/:killId/review", auth.RequireRole(auth.Lector), handler.ReviewKill)
		g.POST("/kills/:killId/like", handler.LikeKill)
		g.DELETE("/kills/:killId/like", handler.UnlikeKill)
		g.GET("/kills/pending/next", handler.GetNextPendingKill)
		g.GET("/kills/pending/count", handler.GetPendingKillCount)
		g.GET("/kills/pending", handler.GetPendingKills)

		g.GET("/feed", handler.GetFeed)
		g.GET("/leaderboard", handler.GetLeaderboard)
		g.GET("/end-screen", handler.GetEndScreen)

		g.GET("/props", auth.RequireRole(auth.Lector), handler.GetAllProps)
		g.POST("/props", auth.RequireRole(auth.Lector), handler.CreateProp)
		g.PUT("/props/:propId", auth.RequireRole(auth.Lector), handler.UpdateProp)
		g.DELETE("/props/:propId", auth.RequireRole(auth.Lector), handler.DeleteProp)
	}
}
