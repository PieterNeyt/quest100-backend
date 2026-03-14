package server

import (
	comApp "Quest100Backend/internal/communication/application"
	database4 "Quest100Backend/internal/communication/infrastructure/database"
	server2 "Quest100Backend/internal/communication/infrastructure/server"
	eventApp "Quest100Backend/internal/event/application"
	database3 "Quest100Backend/internal/event/infrastructure/database"
	gotchaApp "Quest100Backend/internal/gotcha/application"
	gotchaDB "Quest100Backend/internal/gotcha/infrastructure/database"
	"Quest100Backend/internal/infrastructure/database"
	"Quest100Backend/internal/infrastructure/schedular"
	profileApp "Quest100Backend/internal/profile/application"
	database2 "Quest100Backend/internal/profile/infrastructure/database"
	utilApp "Quest100Backend/internal/util/qrcode/application"
	"Quest100Backend/internal/util/qrcode/domain"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Server struct {
	port        int
	db          database.Service
	hub         *server2.Hub
	profileServ profileApp.ProfileService
	eventServ   eventApp.EventService
	chatServ    comApp.ChatService
	qrCodeServ  utilApp.QRCodeService
	gotchaServ  gotchaApp.GotchaService
}

func NewServer() *http.Server {
	db := database.New()

	pRepo := database2.NewProfileRepository(db.GetDB())
	eRepo := database3.NewEventRepository(db.GetDB())
	cRepo := database4.NewChatRepository(db.GetDB())
	qrGen := domain.NewQRCodeGenerator(256)
	gGameRepo := gotchaDB.NewGameRepository(db.GetDB())
	gPartRepo := gotchaDB.NewParticipantRepository(db.GetDB())
	gKillRepo := gotchaDB.NewKillRepository(db.GetDB())
	gPropRepo := gotchaDB.NewPropRepository(db.GetDB())

	pServ := profileApp.NewProfileService(pRepo)
	cServ := comApp.NewChatService(pServ, cRepo, eRepo)
	eServ := eventApp.NewEventService(eRepo, cServ)
	qServ := utilApp.NewQRCodeService(qrGen, pServ)
	gServ := gotchaApp.NewGotchaService(gGameRepo, gPartRepo, gKillRepo, gPropRepo, pServ)

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	newServer := &Server{
		port:        port,
		db:          db,
		hub:         server2.NewHub(),
		profileServ: pServ,
		eventServ:   eServ,
		chatServ:    cServ,
		qrCodeServ:  qServ,
		gotchaServ:  gServ,
	}

	schedular.StartDailyTableCleanup(
		newServer.db.GetDB(),
		[]string{"award_history_entries"},
		"02:00",
	)

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go newServer.hub.Run() // Start the hub in a background routine

	return server
}
