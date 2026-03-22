package server

import (
	"Quest100Backend/internal/minigame/sudoku/api"
	"Quest100Backend/internal/minigame/sudoku/application"

	"github.com/gin-gonic/gin"
)

func SetupSudokuRoutes(r *gin.RouterGroup, sudokuServ application.SudokuService) {
	handler := api.NewSudokuHandler(sudokuServ)

	sudokuGroup := r.Group("/sudoku")
	{
		sudokuGroup.GET("/session", handler.GetSession)
		sudokuGroup.POST("/move", handler.SubmitMove)
	}
}
