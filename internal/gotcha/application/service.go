package application

import (
	"Quest100Backend/internal/gotcha/domain"
	profileDomain "Quest100Backend/internal/profile/domain"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ProfileSummary struct {
	ID             uuid.UUID
	FirstName      string
	LastName       string
	ProfilePicture *string
}

type PropSummary struct {
	ID   uuid.UUID
	Name string
}

type KillFeedItem struct {
	ID         uuid.UUID
	GameID     uuid.UUID
	PhotoURL   string
	Status     domain.KillStatus
	CreatedAt  time.Time
	ReviewedAt *time.Time
	Hunter     ProfileSummary
	Victim     ProfileSummary
	Prop       *PropSummary
	LikeCount  int
	LikedByMe  bool
}

// TargetInfo holds the resolved target profile and assigned prop for a participant.
type TargetInfo struct {
	Target       *ProfileSummary
	AssignedProp *PropSummary
}

type GotchaService interface {
	CreateGame(campus string, startDate time.Time, killDeadlineHours int) (*domain.GotchaGame, error)
	StartGame(campus string) error
	GetCurrentGame(campus string) (*domain.GotchaGame, error)
	UpdateStartDate(campus string, startDate time.Time, killDeadlineHours int) (*domain.GotchaGame, error)

	OptIn(campus string, profileID uuid.UUID) error
	OptOut(campus string, profileID uuid.UUID) error

	SubmitKill(campus string, hunterID uuid.UUID, photoURL string) (*domain.GotchaKill, error)
	ReviewKill(killID, reviewerID uuid.UUID, approve bool) error

	GetKillFeed(campus string, requestingProfileID uuid.UUID, limit, offset int) ([]*KillFeedItem, error)
	GetPendingKills(campus string, requestingProfileID uuid.UUID) ([]*KillFeedItem, error)
	LikeKill(killID, profileID uuid.UUID) error
	UnlikeKill(killID, profileID uuid.UUID) error

	GetMyStatus(campus string, profileID uuid.UUID) (*domain.Participant, error)
	GetTargetInfo(campus string, profileID uuid.UUID) (*TargetInfo, error)
	GetLeaderboard(campus string) ([]*domain.Participant, error)

	ProcessTimeouts() error
	CheckAndStartGames() error
}

type gotchaService struct {
	gameRepo        domain.GameRepository
	participantRepo domain.ParticipantRepository
	killRepo        domain.KillRepository
	propRepo        domain.PropRepository
	profileRepo     profileDomain.ProfileRepository
}

func NewGotchaService(
	gameRepo domain.GameRepository,
	participantRepo domain.ParticipantRepository,
	killRepo domain.KillRepository,
	propRepo domain.PropRepository,
	profileRepo profileDomain.ProfileRepository,
) GotchaService {
	return &gotchaService{gameRepo, participantRepo, killRepo, propRepo, profileRepo}
}

func (s *gotchaService) GetCurrentGame(campus string) (*domain.GotchaGame, error) {
	return s.gameRepo.GetGameByCampus(campus)
}

func (s *gotchaService) CreateGame(campus string, startDate time.Time, killDeadlineHours int) (*domain.GotchaGame, error) {
	if killDeadlineHours <= 0 {
		killDeadlineHours = 72
	}
	game := &domain.GotchaGame{
		ID:                uuid.New(),
		Campus:            campus,
		Status:            domain.StatusOptIn,
		StartDate:         startDate,
		KillDeadlineHours: killDeadlineHours,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if err := s.gameRepo.SaveGame(game); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}
	return game, nil
}

func (s *gotchaService) UpdateStartDate(campus string, startDate time.Time, killDeadlineHours int) (*domain.GotchaGame, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return nil, fmt.Errorf("no active game for campus %s: %w", campus, err)
	}
	if game.Status == domain.StatusFinished {
		return nil, fmt.Errorf("cannot update a finished game")
	}
	if killDeadlineHours > 0 {
		game.KillDeadlineHours = killDeadlineHours
	}
	game.StartDate = startDate
	game.UpdatedAt = time.Now()
	if err := s.gameRepo.SaveGame(game); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}
	return game, nil
}

func (s *gotchaService) StartGame(campus string) error {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return err
	}
	if game.Status != domain.StatusOptIn {
		return fmt.Errorf("game for campus %s is not in opt-in phase", campus)
	}
	participants, err := s.participantRepo.GetParticipantsByGame(game.ID)
	if err != nil {
		return err
	}
	for _, p := range participants {
		game.Participants = append(game.Participants, *p)
	}
	if err := game.AssignTargets(); err != nil {
		return err
	}

	for i := range game.Participants {
		prop, err := s.propRepo.GetRandomProp()
		if err == nil {
			game.Participants[i].AssignedPropID = &prop.ID
		}
	}
	return s.gameRepo.SaveGame(game)
}

func (s *gotchaService) CheckAndStartGames() error {
	games, err := s.gameRepo.GetAllOptInGames()
	if err != nil {
		return err
	}
	for _, game := range games {
		if game.StartDate.IsZero() || game.StartDate.After(time.Now()) {
			continue
		}
		if err := s.StartGame(game.Campus); err != nil {
			fmt.Printf("auto-start failed for campus %s: %v\n", game.Campus, err)
		}
	}
	return nil
}

func (s *gotchaService) OptIn(campus string, profileID uuid.UUID) error {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		game, err = s.CreateGame(campus, time.Time{}, 72)
		if err != nil {
			return fmt.Errorf("failed to create game for campus %s: %w", campus, err)
		}
	}
	if game.Status != domain.StatusOptIn {
		return fmt.Errorf("opt-in period has ended")
	}
	existing, _ := s.participantRepo.GetParticipant(game.ID, profileID)
	if existing != nil {
		return &domain.AlreadyOptedInError{ProfileID: profileID}
	}
	p := &domain.Participant{
		ID:        uuid.New(),
		GameID:    game.ID,
		ProfileID: profileID,
		IsAlive:   true,
		OptedInAt: time.Now(),
	}
	return s.participantRepo.SaveParticipant(p)
}

func (s *gotchaService) OptOut(campus string, profileID uuid.UUID) error {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return fmt.Errorf("no active game for campus %s: %w", campus, err)
	}
	if game.Status != domain.StatusOptIn {
		return fmt.Errorf("cannot opt out after game has started")
	}
	existing, _ := s.participantRepo.GetParticipant(game.ID, profileID)
	if existing == nil {
		return fmt.Errorf("not opted in")
	}
	return s.participantRepo.DeleteParticipant(game.ID, profileID)
}

func (s *gotchaService) SubmitKill(campus string, hunterID uuid.UUID, photoURL string) (*domain.GotchaKill, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return nil, fmt.Errorf("no active game for campus %s: %w", campus, err)
	}
	if game.Status != domain.StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	hunter, err := s.participantRepo.GetParticipant(game.ID, hunterID)
	if err != nil || hunter == nil {
		return nil, &domain.NotParticipantError{ProfileID: hunterID}
	}
	if !hunter.IsAlive {
		return nil, fmt.Errorf("you are eliminated and cannot submit kills")
	}
	if hunter.TargetID == nil {
		return nil, fmt.Errorf("you have no assigned target")
	}

	kill := &domain.GotchaKill{
		ID:        uuid.New(),
		GameID:    game.ID,
		HunterID:  hunterID,
		VictimID:  *hunter.TargetID,
		PhotoURL:  photoURL,
		PropID:    hunter.AssignedPropID,
		Status:    domain.KillPending,
		CreatedAt: time.Now(),
	}
	if err := s.killRepo.SaveKill(kill); err != nil {
		return nil, fmt.Errorf("failed to submit kill: %w", err)
	}
	return kill, nil
}

func (s *gotchaService) ReviewKill(killID, reviewerID uuid.UUID, approve bool) error {
	kill, err := s.killRepo.GetKillByID(killID)
	if err != nil {
		return fmt.Errorf("kill not found: %w", err)
	}
	if kill.Status != domain.KillPending {
		return fmt.Errorf("kill already reviewed")
	}
	now := time.Now()
	kill.ReviewedBy = &reviewerID
	kill.ReviewedAt = &now

	if !approve {
		kill.Status = domain.KillDenied
		return s.killRepo.SaveKill(kill)
	}

	kill.Status = domain.KillApproved
	game, err := s.gameRepo.GetGameByID(kill.GameID)
	if err != nil {
		return err
	}
	participants, err := s.participantRepo.GetParticipantsByGame(game.ID)
	if err != nil {
		return err
	}
	for _, p := range participants {
		game.Participants = append(game.Participants, *p)
	}
	if err := game.ProcessKill(kill.HunterID, kill.VictimID); err != nil {
		return err
	}

	if err := s.assignNewProp(game, kill.HunterID); err != nil {
		fmt.Printf("could not assign new prop to hunter %s: %v\n", kill.HunterID, err)
	}

	if err := s.killRepo.SaveKill(kill); err != nil {
		return err
	}
	return s.gameRepo.SaveGame(game)
}

func (s *gotchaService) assignNewProp(game *domain.GotchaGame, profileID uuid.UUID) error {
	prop, err := s.propRepo.GetRandomProp()
	if err != nil {
		return err
	}
	for i := range game.Participants {
		if game.Participants[i].ProfileID == profileID {
			game.Participants[i].AssignedPropID = &prop.ID
			return s.participantRepo.SaveParticipant(&game.Participants[i])
		}
	}
	return fmt.Errorf("participant not found in game")
}

func (s *gotchaService) GetKillFeed(campus string, requestingProfileID uuid.UUID, limit, offset int) ([]*KillFeedItem, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return nil, err
	}

	kills, err := s.killRepo.GetKillFeed(game.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	return s.hydrateKills(kills, requestingProfileID)
}

// GetPendingKills returns all PENDING kills for the campus game, enriched with profile data.
func (s *gotchaService) GetPendingKills(campus string, requestingProfileID uuid.UUID) ([]*KillFeedItem, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return nil, err
	}

	kills, err := s.killRepo.GetPendingKills(game.ID)
	if err != nil {
		return nil, err
	}

	return s.hydrateKills(kills, requestingProfileID)
}

// GetTargetInfo resolves the current participant's target profile and assigned prop.
func (s *gotchaService) GetTargetInfo(campus string, profileID uuid.UUID) (*TargetInfo, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return nil, err
	}

	participant, err := s.participantRepo.GetParticipant(game.ID, profileID)
	if err != nil || participant == nil {
		return nil, fmt.Errorf("not a participant")
	}

	info := &TargetInfo{}

	if participant.TargetID != nil {
		targetProfile, err := s.profileRepo.GetProfileById(*participant.TargetID)
		if err == nil {
			info.Target = &ProfileSummary{
				ID:             targetProfile.ID,
				FirstName:      targetProfile.FirstName,
				LastName:       targetProfile.LastName,
				ProfilePicture: targetProfile.CustomProfilePicture,
			}
		}
	}

	if participant.AssignedPropID != nil {
		prop, err := s.propRepo.GetPropByID(*participant.AssignedPropID)
		if err == nil {
			info.AssignedProp = &PropSummary{
				ID:   prop.ID,
				Name: prop.Name,
			}
		}
	}

	return info, nil
}

// hydrateKills resolves hunter/victim profiles and prop details for a slice of kills.
func (s *gotchaService) hydrateKills(kills []*domain.GotchaKill, requestingProfileID uuid.UUID) ([]*KillFeedItem, error) {
	profileCache := map[uuid.UUID]*profileSummaryOrEmpty{}

	getProfile := func(id uuid.UUID) ProfileSummary {
		if cached, ok := profileCache[id]; ok {
			return cached.summary
		}
		p, err := s.profileRepo.GetProfileById(id)
		entry := &profileSummaryOrEmpty{}
		if err == nil {
			entry.summary = ProfileSummary{
				ID:             p.ID,
				FirstName:      p.FirstName,
				LastName:       p.LastName,
				ProfilePicture: p.CustomProfilePicture,
			}
		} else {
			entry.summary = ProfileSummary{ID: id, FirstName: "Unknown"}
		}
		profileCache[id] = entry
		return entry.summary
	}

	items := make([]*KillFeedItem, 0, len(kills))
	for _, k := range kills {
		item := &KillFeedItem{
			ID:         k.ID,
			GameID:     k.GameID,
			PhotoURL:   k.PhotoURL,
			Status:     k.Status,
			CreatedAt:  k.CreatedAt,
			ReviewedAt: k.ReviewedAt,
			Hunter:     getProfile(k.HunterID),
			Victim:     getProfile(k.VictimID),
			LikeCount:  len(k.Likes),
		}
		if k.PropID != nil {
			prop, err := s.propRepo.GetPropByID(*k.PropID)
			if err == nil {
				item.Prop = &PropSummary{ID: prop.ID, Name: prop.Name}
			}
		}
		if requestingProfileID != uuid.Nil {
			liked, _ := s.killRepo.HasLiked(k.ID, requestingProfileID)
			item.LikedByMe = liked
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *gotchaService) LikeKill(killID, profileID uuid.UUID) error {
	return s.killRepo.SaveKillLike(&domain.GotchaKillLike{
		KillID:    killID,
		ProfileID: profileID,
		LikedAt:   time.Now(),
	})
}

func (s *gotchaService) UnlikeKill(killID, profileID uuid.UUID) error {
	return s.killRepo.DeleteKillLike(killID, profileID)
}

func (s *gotchaService) GetMyStatus(campus string, profileID uuid.UUID) (*domain.Participant, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return nil, err
	}
	return s.participantRepo.GetParticipant(game.ID, profileID)
}

func (s *gotchaService) GetLeaderboard(campus string) ([]*domain.Participant, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return nil, err
	}
	return s.participantRepo.GetParticipantsByGame(game.ID)
}

func (s *gotchaService) ProcessTimeouts() error {
	expired, err := s.participantRepo.GetExpiredParticipants(time.Now())
	if err != nil {
		return err
	}
	for _, victim := range expired {
		game, err := s.gameRepo.GetGameByID(victim.GameID)
		if err != nil {
			continue
		}
		participants, _ := s.participantRepo.GetParticipantsByGame(game.ID)
		for _, p := range participants {
			game.Participants = append(game.Participants, *p)
		}
		_ = game.ProcessTimeout(victim.ProfileID)
		_ = s.gameRepo.SaveGame(game)
	}
	return nil
}

type profileSummaryOrEmpty struct {
	summary ProfileSummary
}
