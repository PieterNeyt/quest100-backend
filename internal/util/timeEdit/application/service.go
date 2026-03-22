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
	GetTodayAgenda(token string, employeeId int, isStudent bool) ([]domain.AgendaItem, error)
}

type timeEditService struct {
}

func NewTimeEditService() TimeEditService {
	return &timeEditService{}
}

func (s *timeEditService) TimeEditTokenReq() (string, error) {
	req, _ := http.NewRequest("POST", os.Getenv("TIME_EDIT_URL")+"organizations/"+os.Getenv("ORG_ID")+"/api-keys/authenticate", nil)
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
	req, _ := http.NewRequest("POST", os.Getenv("TIME_EDIT_URL")+"reservations/find", bytes.NewBuffer(jsonBody))
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

func (s *timeEditService) GetTodayAgenda(token string, employeeId int, isStudent bool) ([]domain.AgendaItem, error) {
	body := domain.CreateAgendaRequestBody(employeeId, isStudent)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal agenda request: %w", err)
	}

	req, _ := http.NewRequest("POST", os.Getenv("TIME_EDIT_URL")+"reservations/find", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request agenda: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var resv domain.ResponseBody
	if err := json.Unmarshal(respBody, &resv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agenda response: %w", err)
	}

	courseIDs := resv.UniqueCourseIDs()

	courseNames, err := s.fetchCourseNames(token, courseIDs)
	if err != nil {
		return nil, err
	}

	items := make([]domain.AgendaItem, 0, len(resv.Results))
	for _, r := range resv.Results {
		item := domain.AgendaItemFromResult(r)
		item.ResolveCourseFromObjects(r.Objects, courseNames)
		items = append(items, item)
	}

	return items, nil
}

func (s *timeEditService) fetchCourseNames(token string, courseExtIDs []string) (map[string]string, error) {
	if len(courseExtIDs) == 0 {
		return map[string]string{}, nil
	}

	body, _ := json.Marshal(map[string]interface{}{
		"idFormat": "EXTERNAL",
		"extIds":   courseExtIDs,
	})

	req, _ := http.NewRequest("POST", os.Getenv("TIME_EDIT_URL")+"objects/find", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch course objects: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result domain.CourseObjectsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal course objects: %w", err)
	}

	return result.ToNameMap(), nil
}
