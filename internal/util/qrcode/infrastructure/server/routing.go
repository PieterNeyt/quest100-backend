package server

import (
	"Quest100Backend/internal/infrastructure/auth"
	"Quest100Backend/internal/util/qrcode/api"
	"Quest100Backend/internal/util/qrcode/application"

	"github.com/gin-gonic/gin"
)

func SetupQRCodeRoutes(r *gin.RouterGroup, qrServ application.QRCodeService) {
	qrHandler := api.NewQRCodeHandler(qrServ)

	qrGroup := r.Group("/qrcode")
	{
		qrGroup.POST("/generate", auth.RequireRole(auth.Lector), qrHandler.GenerateQRCode)
	}
}
