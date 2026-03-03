package server

import (
	"Quest100Backend/internal/infrastructure/auth"
	"Quest100Backend/internal/util/qrcode/api"
	"Quest100Backend/internal/util/qrcode/application"
	"Quest100Backend/internal/util/qrcode/domain"

	"github.com/gin-gonic/gin"
)

func SetupQRCodeRoutes(r *gin.RouterGroup) {

	qrGenerator := domain.NewQRCodeGenerator(256)
	qrService := application.NewQRCodeService(qrGenerator)
	qrHandler := api.NewQRCodeHandler(qrService)

	qrGroup := r.Group("/qrcode")
	{
		qrGroup.POST("/generate", auth.RequireRole(auth.Lector), qrHandler.GenerateQRCode)
	}
}
