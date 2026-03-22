package server

import (
	"Quest100Backend/internal/minigame/nerdle/api"
	"Quest100Backend/internal/minigame/nerdle/application"

	"github.com/gin-gonic/gin"
)

func SetupNerdleRoutes(r *gin.RouterGroup, nerdleServ application.NerdleService) {
	handler := api.NewNerdleHandler(nerdleServ)

	nerdleGroup := r.Group("/nerdle")
	{
		nerdleGroup.GET("/session", handler.GetSession)
		nerdleGroup.POST("/guess", handler.SubmitGuess)
	}
}
