package application

import (
	"Quest100Backend/internal/profile/domain"
	"fmt"
	"os"
	"strconv"

	"github.com/google/uuid"
)

type ProfileService interface {
	HandleAttendance(classId uuid.UUID, profileId uuid.UUID) (*domain.Profile, error)
}

type profileService struct {
	profileRepo domain.ProfileRepository
}

func NewProfileService(profileRepo domain.ProfileRepository) ProfileService {
	return &profileService{
		profileRepo: profileRepo,
	}
}

func (s *profileService) HandleAttendance(classId uuid.UUID, profileId uuid.UUID) (*domain.Profile, error) {
	// TODO: Implement logic to check if user is in the class

	profile, err := s.profileRepo.GetProfileById(profileId)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	kudos, err := strconv.Atoi(os.Getenv("ATTENDANCE_KUDOS"))
	if err != nil {
		return nil, fmt.Errorf("invalid ATTENDANCE_KUDOS value: %w", err)
	}

	if err := profile.AddKudos(kudos); err != nil {
		return nil, fmt.Errorf("failed to add kudos: %w", err)
	}

	if err := s.profileRepo.UpdateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return profile, nil
}
