package application

import (
	"Quest100Backend/internal/profile/application"
	"Quest100Backend/internal/util/qrcode/domain"
	application2 "Quest100Backend/internal/util/timeEdit/application"
	domain2 "Quest100Backend/internal/util/timeEdit/domain"
	"fmt"
	"os"

	"github.com/google/uuid"
)

type QRCodeService interface {
	GenerateAttendanceQRCode(profileId uuid.UUID) (string, error)
}

type qrCodeService struct {
	qrGenerator     domain.QRCodeGenerator
	profileService  application.ProfileService
	timeEditService application2.TimeEditService
}

func NewQRCodeService(qrGenerator domain.QRCodeGenerator, profileService application.ProfileService, timeEditService application2.TimeEditService) QRCodeService {
	return &qrCodeService{
		qrGenerator:     qrGenerator,
		profileService:  profileService,
		timeEditService: timeEditService,
	}
}

func (s *qrCodeService) GenerateAttendanceQRCode(profileId uuid.UUID) (string, error) {
	profile, err := s.profileService.GetProfileById(profileId)
	if err != nil {
		return "", fmt.Errorf("failed to get profile: %w", err)
	}
	token, err := s.timeEditService.TimeEditTokenReq()
	if err != nil {
		return "", err
	}
	classID, err := s.timeEditService.TimeEditReservationsReq(token, domain2.Lector, profile.EmployeeID)
	if err != nil {
		return "", err
	}

	frontendURL := os.Getenv("FRONTEND_URL")

	qrCode, err := s.qrGenerator.GenerateAttendanceQRCode(classID, frontendURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate attendance qr code: %w", err)
	}

	return qrCode, nil
}
