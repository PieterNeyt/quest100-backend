package server

import (
	"Quest100Backend/internal/minigame/minesweeper/api"
	"Quest100Backend/internal/minigame/minesweeper/application"

	"github.com/gin-gonic/gin"
)

func SetupMinesweeperRoutes(r *gin.RouterGroup, minesweeperServ application.MinesweeperService) {
	handler := api.NewMinesweeperHandler(minesweeperServ)

	minesweeperGroup := r.Group("/minesweeper")
	{
		minesweeperGroup.GET("/session", handler.GetSession)
		minesweeperGroup.POST("/move", handler.SubmitMove)
	}
}
