package application

import (
	"Quest100Backend/internal/util/timeEdit/domain"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type TimeEditService interface {
	TimeEditTokenReq() (string, error)
	TimeEditReservationsReq(token string, typeID domain.TypeID, id int) (int, error)
}

type timeEditService struct {
}

func NewTimeEditService() TimeEditService {
	return &timeEditService{}
}

func (s *timeEditService) TimeEditTokenReq() (string, error) {
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

func (s *timeEditService) TimeEditReservationsReq(token string, typeID domain.TypeID, employeeId int) (int, error) {
	body := domain.CreateRequestBody(typeID, employeeId)

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
