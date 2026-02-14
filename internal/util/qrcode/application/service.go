package application

import (
	"Quest100Backend/internal/util/qrcode/domain"
	"fmt"
	"os"
)

type QRCodeService interface {
	GenerateQRCode(id string) (string, error)
	GenerateAttendanceQRCode(classID string) (string, error)
}

type qrCodeService struct {
	qrGenerator domain.QRCodeGenerator
}

func NewQRCodeService(qrGenerator domain.QRCodeGenerator) QRCodeService {
	return &qrCodeService{
		qrGenerator: qrGenerator,
	}
}

func (s *qrCodeService) GenerateQRCode(id string) (string, error) {
	if id == "" {
		return "", &domain.InvalidQRCodeDataError{Message: "ID cannot be empty"}
	}

	qrCode, err := s.qrGenerator.GenerateQRCode(id)
	if err != nil {
		return "", fmt.Errorf("failed to generate qr code: %w", err)
	}

	return qrCode, nil
}

func (s *qrCodeService) GenerateAttendanceQRCode(classID string) (string, error) {
	if classID == "" {
		return "", &domain.InvalidQRCodeDataError{Message: "Class ID cannot be empty"}
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200" // fallback
	}

	qrCode, err := s.qrGenerator.GenerateAttendanceQRCode(classID, frontendURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate attendance qr code: %w", err)
	}

	return qrCode, nil
}
