package server

import (
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

	api := r.Group("/api", AuthMiddleware())
	profileRouting.SetupProfileRoutes(api, s.db.GetDB())
	qrcodeRouting.SetupQRCodeRoutes(api)

	r.GET("/ws", s.handleWebSocket)

	return r
}
