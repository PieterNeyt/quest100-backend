package application

import (
	"Quest100Backend/internal/profile/domain"
	"encoding/base64"
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
	GetGraphProfilePicture(token string) (string, error)
	UpdateProfilePicture(profileId uuid.UUID, base64Img string) (*domain.Profile, error)
	DeleteProfilePicture(profileId uuid.UUID) (*domain.Profile, error)
	GiveAwardTo(senderId uuid.UUID, recieverId uuid.UUID, kudoType domain.KudoType, message string) (*domain.Profile, error)
	GetProfiles() (*[]domain.Profile, error)
	GetProfilesWithAward(profileId uuid.UUID) (*[]domain.ProfileAward, error)
	GetProfilesStatistics(profileUUID uuid.UUID) (domain.ProfileStats, error)
	GetCampusByProfileID(id uuid.UUID) (string, error)
}

type profileService struct {
	profileRepo domain.ProfileRepository
}

func NewProfileService(profileRepo domain.ProfileRepository) ProfileService {
	return &profileService{
		profileRepo: profileRepo,
	}
}

func (s *profileService) GetCampusByProfileID(id uuid.UUID) (string, error) {
	return s.profileRepo.GetCampusByProfileID(id)
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

func (s *profileService) GetProfiles() (*[]domain.Profile, error) {
	return s.profileRepo.GetProfiles()
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
	req, _ := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me?$select=id,employeeId,givenName,surname,mail,preferredLanguage,officeLocation", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var user domain.GraphProfile
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return &user, nil
}
func (s *profileService) GetGraphProfilePicture(token string) (string, error) {
	req, _ := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me/photo/$value", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("no profile picture: %d", resp.StatusCode)
	}

	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(imgBytes), nil
}

func (s *profileService) UpdateProfilePicture(profileId uuid.UUID, base64Img string) (*domain.Profile, error) {
	profile, err := s.profileRepo.GetProfileById(profileId)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}
	profile.CustomProfilePicture = &base64Img
	if err := s.profileRepo.UpdateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to update picture: %w", err)
	}
	return profile, nil
}

func (s *profileService) DeleteProfilePicture(profileId uuid.UUID) (*domain.Profile, error) {
	profile, err := s.profileRepo.GetProfileById(profileId)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}
	profile.CustomProfilePicture = nil
	if err := s.profileRepo.UpdateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to delete picture: %w", err)
	}
	return profile, nil
}

func (s *profileService) GiveAwardTo(senderId uuid.UUID, receiverId uuid.UUID, kudoType domain.KudoType, message string) (*domain.Profile, error) {
	profile, err := s.profileRepo.GetProfileById(receiverId)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	if err := s.profileRepo.AddAwardHistoryEntry(senderId, receiverId); err != nil {
		return nil, fmt.Errorf("failed to add award history entry: %w", err)
	}

	if kudos, err := strconv.Atoi(os.Getenv("AWARD_KUDOS")); err == nil {
		if err := profile.AddKudos(kudos, message, kudoType); err != nil {
			return nil, fmt.Errorf("failed to add kudos: %w", err)
		}
	}

	if err := s.profileRepo.UpdateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return profile, nil
}

func (s *profileService) GetProfilesWithAward(profileId uuid.UUID) (*[]domain.ProfileAward, error) {
	profiles, err := s.GetProfiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles: %w", err)
	}

	receivers, err := s.profileRepo.GetSentAwardReceivers(profileId)
	if err != nil {
		return nil, fmt.Errorf("failed to get sent award receivers: %w", err)
	}

	receiverMap := make(map[uuid.UUID]struct{}, len(receivers))
	for _, id := range receivers {
		receiverMap[id] = struct{}{}
	}

	var profileAwards []domain.ProfileAward

	for _, profile := range *profiles {
		if profile.ID == profileId {
			continue
		}

		_, hasSent := receiverMap[profile.ID]

		profileAwards = append(profileAwards, domain.ProfileAward{
			Profile:      profile,
			HasSentAward: hasSent,
		})
	}

	return &profileAwards, nil
}

func (s *profileService) GetProfilesStatistics(profileUUID uuid.UUID) (domain.ProfileStats, error) {
	profileStats, err := s.profileRepo.GetProfileStats(profileUUID)
	if err != nil {
		return domain.ProfileStats{}, fmt.Errorf("failed to get profile stats: %w", err)
	}
	return profileStats, nil
}
