package server

import (
	commRouting "Quest100Backend/internal/communication/infrastructure/server"
	eventRouting "Quest100Backend/internal/event/infrastructure/server"
	gotchaRouting "Quest100Backend/internal/gotcha/infrastructure/server"
	"Quest100Backend/internal/infrastructure/auth"
	profileRouting "Quest100Backend/internal/profile/infrastructure/server"
	qrcodeRouting "Quest100Backend/internal/util/qrcode/infrastructure/server"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-Graph-Token"},
		AllowCredentials: true,
	}))

	api := r.Group("/api", auth.AuthMiddleware())
	profileRouting.SetupProfileRoutes(api, s.profileServ)
	eventRouting.SetupEventRoutes(api, s.eventServ)
	qrcodeRouting.SetupQRCodeRoutes(api, s.qrCodeServ)
	commRouting.SetupCommunicationsRoutes(api, s.chatServ)
	gotchaRouting.SetupGotchaRoutes(api, s.gotchaServ, s.profileServ)

	commRouting.SetupWebSocketRoutes(r, s.chatServ, s.hub)

	r.GET("/debug/ws", func(c *gin.Context) {
		// Calling the method we just created
		snapshot := s.hub.GetSnapshot()

		c.JSON(200, snapshot)
	})
	return r
}
