package application

import (
	"Quest100Backend/internal/profile/application"
	"Quest100Backend/internal/util/qrcode/domain"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type QRCodeService interface {
	GenerateAttendanceQRCode(profileId uuid.UUID) (string, error)
}

type qrCodeService struct {
	qrGenerator    domain.QRCodeGenerator
	profileService application.ProfileService
}

func NewQRCodeService(qrGenerator domain.QRCodeGenerator, profileService application.ProfileService) QRCodeService {
	return &qrCodeService{
		qrGenerator:    qrGenerator,
		profileService: profileService,
	}
}

func (s *qrCodeService) GenerateAttendanceQRCode(profileId uuid.UUID) (string, error) {
	profile, err := s.profileService.GetProfileById(profileId)
	if err != nil {
		return "", fmt.Errorf("failed to get profile: %w", err)
	}
	token, err := timeEditTokenReq()
	if err != nil {
		return "", err
	}
	classID, err := timeEditReservationsReq(token, domain.Lector, profile.EmployeeID)
	if err != nil {
		return "", err
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

func timeEditTokenReq() (string, error) {
	req, _ := http.NewRequest("POST", "https://api.test.timeedit.net/v1/organizations/"+os.Getenv("ORG_ID")+"/api-keys/authenticate", nil)
	req.Header.Set("Authorization", os.Getenv("API_KEY"))
	req.Header.Set("X-Region", "EU_EES")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var token map[string]string
	if err := json.Unmarshal(body, &token); err != nil {
		return "", fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return token["token"], nil
}

func timeEditReservationsReq(token string, typeID domain.TypeID, id int) (int, error) {
	timeNow := time.Now().Unix()
	body := domain.RequestBody{
		Date: domain.Date{
			StartDate: timeNow,
			EndDate:   timeNow,
		},
		IDFormat: "EXTERNAL",
		SearchObjects: []domain.SearchObject{
			{
				TypeID:   typeID,
				ObjectID: fmt.Sprintf("person_%d", id),
			},
		},
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "https://api.test.timeedit.net/v1/reservations/find", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to request reservations: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var resv = domain.ResponseBody{}
	if err := json.Unmarshal(respBody, &resv); err != nil {
		return 0, fmt.Errorf("failed to unmarshal reservation response: %w", err)
	}
	if resv.TotalResults != 1 {
		return 0, fmt.Errorf("failed to request reservation response: expected 1 result, got %d", resv.TotalResults)
	}
	return resv.Results[0].ID, nil
}
