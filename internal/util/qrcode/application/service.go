package application

import (
	"Quest100Backend/internal/util/qrcode/domain"
	"fmt"
)

type QRCodeService interface {
	GenerateQRCode(id string) (string, error)
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
