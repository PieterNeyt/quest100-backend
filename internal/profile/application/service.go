package application

import (
	"Quest100Backend/internal/profile/domain"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
)

type ProfileService interface {
	HandleAttendance(classId uuid.UUID, profileId uuid.UUID) (*domain.Profile, int, bool, error)
	Sync(graphProfile *domain.GraphProfile) (*domain.Profile, error)
	GetGraphProfile(token string) (*domain.GraphProfile, error)
	GetProfileById(id uuid.UUID) (*domain.Profile, error)
	UpdateProfile(profile *domain.Profile) error
}

type profileService struct {
	profileRepo domain.ProfileRepository
}

func NewProfileService(profileRepo domain.ProfileRepository) ProfileService {
	return &profileService{
		profileRepo: profileRepo,
	}
}

func (s *profileService) HandleAttendance(classId uuid.UUID, profileId uuid.UUID) (*domain.Profile, int, bool, error) {

	profile, err := s.GetProfileById(profileId)
	if err != nil {
		return nil, 0, false, fmt.Errorf("failed to get profile: %w", err)
	}
	if err := profile.RecordAttendance(classId); err != nil {
		var dupErr *domain.DuplicateAttendanceError
		if errors.As(err, &dupErr) {
			return profile, 0, true, nil
		}
		return nil, 0, false, fmt.Errorf("failed to record attendance: %w", err)
	}

	kudos, err := strconv.Atoi(os.Getenv("ATTENDANCE_KUDOS"))
	if err != nil {
		return nil, 0, false, fmt.Errorf("invalid ATTENDANCE_KUDOS value: %w", err)
	}

	if err := profile.AddKudos(kudos, os.Getenv("ATTENDANCE_MESSAGE"), domain.KudoAttendance); err != nil {
		return nil, 0, false, fmt.Errorf("failed to add kudos: %w", err)
	}

	if err := s.UpdateProfile(profile); err != nil {
		return nil, 0, false, fmt.Errorf("failed to update profile: %w", err)
	}

	return profile, kudos, false, nil
}
func (s *profileService) GetProfileById(id uuid.UUID) (*domain.Profile, error) {
	return s.profileRepo.GetProfileById(id)
}

func (s *profileService) UpdateProfile(profile *domain.Profile) error {
	return s.profileRepo.UpdateProfile(profile)
}
func (s *profileService) Sync(graphProfile *domain.GraphProfile) (*domain.Profile, error) {
	profile, err := s.GetProfileById(graphProfile.Id)
	if err != nil {
		profile = domain.CreateProfile(graphProfile)
		if err := s.profileRepo.SaveProfile(profile); err != nil {
			return nil, fmt.Errorf("failed to save profile: %w", err)
		}
		return profile, nil
	}

	if err := profile.Sync(graphProfile); err != nil {
		return nil, fmt.Errorf("failed to sync profile: %w", err)
	}

	if err := s.UpdateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return profile, nil
}

func (s *profileService) GetGraphProfile(token string) (*domain.GraphProfile, error) {
	req, _ := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))

	var user domain.GraphProfile
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return &user, nil
}
