package application

import (
	"Quest100Backend/internal/gotcha/domain"
	profileService "Quest100Backend/internal/profile/application"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type GotchaService interface {
	CreateGame(campus string, startDate time.Time, killDeadlineHours int, prizePhotoBase64, prizeDescEN, prizeDescNL string) (*domain.GotchaGame, error)
	StartGame(campus string) error
	GetCurrentGame(profileID uuid.UUID) (*domain.GotchaGame, error)
	UpdateStartDate(campus string, startDate time.Time, killDeadlineHours int, prizePhotoBase64, prizeDescEN, prizeDescNL string) (*domain.GotchaGame, error)

	OptIn(profileID uuid.UUID) error
	OptOut(profileID uuid.UUID) error

	SubmitKill(hunterID uuid.UUID, photoBase64 string) (*domain.GotchaKill, error)
	ReviewKill(killID, reviewerID uuid.UUID, approve bool) error

	GetNextPendingKill(requestingProfileID uuid.UUID) (*KillFeedItem, error)
	GetPendingKillCount(profileID uuid.UUID) (int, error)

	GetKillFeed(requestingProfileID uuid.UUID, limit, offset int) ([]*KillFeedItem, error)
	GetPendingKills(requestingProfileID uuid.UUID) ([]*KillFeedItem, error)
	LikeKill(killID, profileID uuid.UUID) error
	UnlikeKill(killID, profileID uuid.UUID) error

	GetMyStatus(profileID uuid.UUID) (*domain.Participant, error)
	GetTargetInfo(profileID uuid.UUID) (*TargetInfo, error)
	GetLeaderboard(profileID uuid.UUID) ([]*domain.Participant, error)

	GetEndScreen(profileID uuid.UUID) (*EndScreen, error)
	GetEndScreenByID(gameID uuid.UUID) (*EndScreen, error)

	GetGameHistory(profileID uuid.UUID) ([]*GameSummary, error)

	GetAllPropsByGame(profileID uuid.UUID) ([]*domain.GotchaProp, error)
	CreateProp(profileID uuid.UUID, nameEN, nameNL string) (*domain.GotchaProp, error)
	UpdateProp(id uuid.UUID, nameEN, nameNL string) (*domain.GotchaProp, error)
	DeleteProp(id uuid.UUID) error

	ProcessTimeouts() error
	CheckAndStartGames() error
}

type gotchaService struct {
	gameRepo        domain.GameRepository
	participantRepo domain.ParticipantRepository
	killRepo        domain.KillRepository
	propRepo        domain.PropRepository
	profileService  profileService.ProfileService
}

func NewGotchaService(
	gameRepo domain.GameRepository,
	participantRepo domain.ParticipantRepository,
	killRepo domain.KillRepository,
	propRepo domain.PropRepository,
	profileService profileService.ProfileService,
) GotchaService {
	return &gotchaService{gameRepo, participantRepo, killRepo, propRepo, profileService}
}

func (s *gotchaService) campusFor(profileID uuid.UUID) (string, error) {
	return s.profileService.GetCampusByProfileID(profileID)
}

// Game

func (s *gotchaService) GetCurrentGame(profileID uuid.UUID) (*domain.GotchaGame, error) {
	campus, err := s.campusFor(profileID)
	if err != nil {
		return nil, fmt.Errorf("could not determine campus: %w", err)
	}
	return s.gameRepo.GetGameByCampus(campus)
}

func (s *gotchaService) CreateGame(campus string, startDate time.Time, killDeadlineHours int, prizePhotoBase64, prizeDescEN, prizeDescNL string) (*domain.GotchaGame, error) {
	existing, err := s.gameRepo.GetActiveGameByCampus(campus)
	if err == nil && existing != nil {
		return s.UpdateStartDate(campus, startDate, killDeadlineHours, prizePhotoBase64, prizeDescEN, prizeDescNL)
	}

	if killDeadlineHours <= 0 {
		killDeadlineHours = 72
	}
	game := &domain.GotchaGame{
		ID:                 uuid.New(),
		Campus:             campus,
		Status:             domain.StatusOptIn,
		StartDate:          startDate,
		KillDeadlineHours:  killDeadlineHours,
		PrizePhotoBase64:   prizePhotoBase64,
		PrizeDescriptionEN: prizeDescEN,
		PrizeDescriptionNL: prizeDescNL,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	if err := s.gameRepo.SaveGame(game); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}
	return game, nil
}

func (s *gotchaService) UpdateStartDate(campus string, startDate time.Time, killDeadlineHours int, prizePhotoBase64, prizeDescEN, prizeDescNL string) (*domain.GotchaGame, error) {
	game, err := s.gameRepo.GetActiveGameByCampus(campus)
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
	if prizePhotoBase64 != "" {
		game.PrizePhotoBase64 = prizePhotoBase64
	}
	if prizeDescEN != "" {
		game.PrizeDescriptionEN = prizeDescEN
	}
	if prizeDescNL != "" {
		game.PrizeDescriptionNL = prizeDescNL
	}
	game.UpdatedAt = time.Now()
	if err := s.gameRepo.SaveGame(game); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}
	return game, nil
}

func (s *gotchaService) StartGame(campus string) error {
	game, err := s.gameRepo.GetActiveGameByCampus(campus)
	if err != nil {
		return err
	}
	if game.Status != domain.StatusOptIn {
		return fmt.Errorf("game for campus %s is already active or finished", campus)
	}

	props, _ := s.propRepo.GetPropsByGame(game.ID)
	if len(props) == 0 {
		return fmt.Errorf("cannot start game without any props defined")
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
		prop, err := s.propRepo.GetRandomProp(game.ID)
		if err == nil {
			game.Participants[i].AssignedPropID = &prop.ID
		}
		if err := s.participantRepo.SaveParticipant(&game.Participants[i]); err != nil {
			return err
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

// Opt-in / out

func (s *gotchaService) OptIn(profileID uuid.UUID) error {
	campus, err := s.campusFor(profileID)
	if err != nil {
		return fmt.Errorf("could not determine campus: %w", err)
	}
	game, err := s.gameRepo.GetActiveGameByCampus(campus)
	if err != nil {
		game, err = s.CreateGame(campus, time.Time{}, 72, "", "", "")
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

func (s *gotchaService) OptOut(profileID uuid.UUID) error {
	campus, err := s.campusFor(profileID)
	if err != nil {
		return fmt.Errorf("could not determine campus: %w", err)
	}
	game, err := s.gameRepo.GetActiveGameByCampus(campus)
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

// Kills

func (s *gotchaService) SubmitKill(hunterID uuid.UUID, photoBase64 string) (*domain.GotchaKill, error) {
	campus, err := s.campusFor(hunterID)
	if err != nil {
		return nil, fmt.Errorf("could not determine campus: %w", err)
	}
	game, err := s.gameRepo.GetActiveGameByCampus(campus)
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
	if err := hunter.CanSubmitKill(); err != nil {
		return nil, err
	}

	hasPending, err := s.killRepo.HasPendingKill(game.ID, hunterID)
	if err != nil {
		return nil, fmt.Errorf("could not check pending kills: %w", err)
	}
	if hasPending {
		return nil, fmt.Errorf("you already have a kill waiting for review")
	}

	kill := &domain.GotchaKill{
		ID:          uuid.New(),
		GameID:      game.ID,
		HunterID:    hunterID,
		VictimID:    *hunter.TargetID,
		PhotoBase64: photoBase64,
		PropID:      hunter.AssignedPropID,
		Status:      domain.KillPending,
		CreatedAt:   time.Now(),
	}
	if err := s.killRepo.SaveKill(kill); err != nil {
		return nil, fmt.Errorf("failed to submit kill: %w", err)
	}

	hunter.MarkPendingKill()
	if err := s.participantRepo.SaveParticipant(hunter); err != nil {
		return nil, fmt.Errorf("failed to update hunter pending status: %w", err)
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
	oldest, err := s.killRepo.GetOldestPendingKill(kill.GameID)
	if err != nil {
		return fmt.Errorf("could not determine review order: %w", err)
	}
	if oldest.ID != killID {
		return fmt.Errorf("another kill must be reviewed first (FIFO order)")
	}
	now := time.Now()
	kill.ReviewedBy = &reviewerID
	kill.ReviewedAt = &now
	if approve {
		return s.approveKill(kill, reviewerID)
	}
	return s.denyKill(kill)
}

func (s *gotchaService) denyKill(kill *domain.GotchaKill) error {
	kill.Status = domain.KillDenied
	hunter, err := s.participantRepo.GetParticipant(kill.GameID, kill.HunterID)
	if err != nil || hunter == nil {
		return s.killRepo.SaveKill(kill)
	}

	if hunter.IsTimedOut() {
		game, err := s.gameRepo.GetGameByID(kill.GameID)
		if err == nil {
			participants, _ := s.participantRepo.GetParticipantsByGame(game.ID)
			for _, p := range participants {
				game.Participants = append(game.Participants, *p)
			}
			_ = game.ProcessTimeout(kill.HunterID)
			for i := range game.Participants {
				_ = s.participantRepo.SaveParticipant(&game.Participants[i])
			}
			_ = s.gameRepo.SaveGame(game)
		}
	} else {
		hunter.ClearPendingKill()
		_ = s.participantRepo.SaveParticipant(hunter)
	}

	return s.killRepo.SaveKill(kill)
}

func (s *gotchaService) approveKill(kill *domain.GotchaKill, reviewerID uuid.UUID) error {
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

	if err := game.ValidateKillApproval(kill.HunterID, kill.VictimID); err != nil {
		kill.Status = domain.KillDenied
		return s.killRepo.SaveKill(kill)
	}

	kill.Status = domain.KillApproved
	if err := game.ProcessKill(kill.HunterID, kill.VictimID); err != nil {
		return err
	}
	if err := s.assignNewProp(game, kill.HunterID); err != nil {
		fmt.Printf("could not assign new prop: %v\n", err)
	}

	hunter := game.FindParticipant(kill.HunterID)
	victim := game.FindParticipant(kill.VictimID)

	hunter.ClearPendingKill()
	victim.ClearPendingKill()

	if err := s.participantRepo.SaveParticipant(hunter); err != nil {
		return err
	}
	if err := s.participantRepo.SaveParticipant(victim); err != nil {
		return err
	}

	pendingKills, err := s.killRepo.GetPendingKillsByHunter(kill.GameID, kill.VictimID)
	if err == nil {
		for _, pk := range pendingKills {
			now := time.Now()
			pk.Status = domain.KillDenied
			pk.ReviewedBy = &reviewerID
			pk.ReviewedAt = &now
			_ = s.killRepo.SaveKill(pk)
		}
	}

	if err := s.killRepo.SaveKill(kill); err != nil {
		return err
	}
	return s.gameRepo.SaveGame(game)
}

func (s *gotchaService) GetNextPendingKill(requestingProfileID uuid.UUID) (*KillFeedItem, error) {
	campus, err := s.campusFor(requestingProfileID)
	if err != nil {
		return nil, fmt.Errorf("could not determine campus: %w", err)
	}
	game, err := s.gameRepo.GetActiveGameByCampus(campus)
	if err != nil {
		return nil, err
	}
	kill, err := s.killRepo.GetOldestPendingKill(game.ID)
	if err != nil {
		return nil, err
	}
	items, err := s.hydrateKills([]*domain.GotchaKill{kill}, requestingProfileID)
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return items[0], nil
}

func (s *gotchaService) GetPendingKillCount(profileID uuid.UUID) (int, error) {
	campus, err := s.campusFor(profileID)
	if err != nil {
		return 0, fmt.Errorf("could not determine campus: %w", err)
	}
	game, err := s.gameRepo.GetActiveGameByCampus(campus)
	if err != nil {
		return 0, err
	}
	return s.killRepo.CountPendingKills(game.ID)
}

func (s *gotchaService) assignNewProp(game *domain.GotchaGame, profileID uuid.UUID) error {
	prop, err := s.propRepo.GetRandomProp(game.ID)
	if err != nil {
		return fmt.Errorf("could not assign new prop: no props found for game %s: %w", game.ID, err)
	}
	for i := range game.Participants {
		if game.Participants[i].ProfileID == profileID {
			game.Participants[i].AssignedPropID = &prop.ID
			return s.participantRepo.SaveParticipant(&game.Participants[i])
		}
	}
	return fmt.Errorf("participant %s not found in game %s", profileID, game.ID)
}

func (s *gotchaService) GetKillFeed(requestingProfileID uuid.UUID, limit, offset int) ([]*KillFeedItem, error) {
	campus, err := s.campusFor(requestingProfileID)
	if err != nil {
		return nil, fmt.Errorf("could not determine campus: %w", err)
	}
	game, err := s.gameRepo.GetActiveGameByCampus(campus)
	if err != nil {
		return nil, err
	}
	kills, err := s.killRepo.GetApprovedKillFeed(game.ID, limit, offset)
	if err != nil {
		return nil, err
	}
	return s.hydrateKills(kills, requestingProfileID)
}

func (s *gotchaService) GetPendingKills(requestingProfileID uuid.UUID) ([]*KillFeedItem, error) {
	campus, err := s.campusFor(requestingProfileID)
	if err != nil {
		return nil, fmt.Errorf("could not determine campus: %w", err)
	}
	game, err := s.gameRepo.GetActiveGameByCampus(campus)
	if err != nil {
		return nil, err
	}
	kills, err := s.killRepo.GetPendingKills(game.ID)
	if err != nil {
		return nil, err
	}
	return s.hydrateKills(kills, requestingProfileID)
}

// Target info

func (s *gotchaService) GetTargetInfo(profileID uuid.UUID) (*TargetInfo, error) {
	campus, err := s.campusFor(profileID)
	if err != nil {
		return nil, fmt.Errorf("could not determine campus: %w", err)
	}
	game, err := s.gameRepo.GetActiveGameByCampus(campus)
	if err != nil {
		return nil, err
	}
	participant, err := s.participantRepo.GetParticipant(game.ID, profileID)
	if err != nil || participant == nil {
		return nil, fmt.Errorf("not a participant")
	}
	info := &TargetInfo{}
	if participant.TargetID != nil {
		targetProfile, err := s.profileService.GetProfileById(*participant.TargetID)
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
			info.AssignedProp = &PropSummary{ID: prop.ID, NameEN: prop.NameEN, NameNL: prop.NameNL}
		}
	}
	if !participant.KillDeadline.IsZero() {
		info.KillDeadline = &participant.KillDeadline
	}
	return info, nil
}

// End screen

func (s *gotchaService) GetEndScreen(profileID uuid.UUID) (*EndScreen, error) {
	campus, err := s.campusFor(profileID)
	if err != nil {
		return nil, fmt.Errorf("could not determine campus: %w", err)
	}
	game, err := s.gameRepo.GetGameByFinishedCampus(campus)
	if err != nil {
		return nil, fmt.Errorf("no finished game for campus %s: %w", campus, err)
	}
	return s.buildEndScreen(game)
}

func (s *gotchaService) GetEndScreenByID(gameID uuid.UUID) (*EndScreen, error) {
	game, err := s.gameRepo.GetGameByID(gameID)
	if err != nil {
		return nil, fmt.Errorf("game not found: %w", err)
	}
	if game.Status != domain.StatusFinished {
		return nil, fmt.Errorf("game is not finished yet")
	}
	return s.buildEndScreen(game)
}

func (s *gotchaService) buildEndScreen(game *domain.GotchaGame) (*EndScreen, error) {
	participants, err := s.participantRepo.GetParticipantsByGame(game.ID)
	if err != nil {
		return nil, err
	}
	participantMap := map[uuid.UUID]*domain.Participant{}
	for _, p := range participants {
		participantMap[p.ProfileID] = p
	}

	type cached struct{ summary ProfileSummary }
	profileCache := map[uuid.UUID]*cached{}
	getProfile := func(id uuid.UUID) ProfileSummary {
		if c, ok := profileCache[id]; ok {
			return c.summary
		}
		p, err := s.profileService.GetProfileById(id)
		e := &cached{}
		if err == nil {
			e.summary = ProfileSummary{ID: p.ID, FirstName: p.FirstName, LastName: p.LastName, ProfilePicture: p.CustomProfilePicture}
		} else {
			e.summary = ProfileSummary{ID: id, FirstName: "Unknown"}
		}
		profileCache[id] = e
		return e.summary
	}

	allKills, err := s.killRepo.GetApprovedKillsByGame(game.ID)
	if err != nil {
		return nil, err
	}

	killCounts := map[uuid.UUID]int{}
	for _, k := range allKills {
		killCounts[k.HunterID]++
	}

	var winner *ProfileSummary
	winnerKillCount := 0
	if game.WinnerID != nil {
		ws := getProfile(*game.WinnerID)
		winner = &ws
		winnerKillCount = killCounts[*game.WinnerID]
	}

	mostKillsID := uuid.Nil
	mostKillsCount := 0
	for pid, cnt := range killCounts {
		if cnt > mostKillsCount {
			mostKillsCount = cnt
			mostKillsID = pid
		}
	}
	mostKillsName := ""
	if mostKillsID != uuid.Nil {
		p := getProfile(mostKillsID)
		mostKillsName = p.FirstName + " " + p.LastName
	}

	fastestSecs := 0
	if len(allKills) > 0 && !game.StartDate.IsZero() {
		earliest := allKills[0].CreatedAt
		for _, k := range allKills {
			if k.CreatedAt.Before(earliest) {
				earliest = k.CreatedAt
			}
		}
		if diff := earliest.Sub(game.StartDate); diff > 0 {
			fastestSecs = int(diff.Seconds())
		}
	}

	killNodes := make([]EndScreenKillNode, 0, len(allKills))
	for _, k := range allKills {
		fullKill, err := s.killRepo.GetKillByID(k.ID)
		likeCount := 0
		if err == nil {
			likeCount = len(fullKill.Likes)
		}
		var targetAssignedAt *time.Time
		if p, ok := participantMap[k.HunterID]; ok && !p.KillDeadline.IsZero() {
			assigned := p.KillDeadline.Add(-time.Duration(game.KillDeadlineHours) * time.Hour)
			targetAssignedAt = &assigned
		}
		node := EndScreenKillNode{
			KillID:           k.ID,
			Hunter:           getProfile(k.HunterID),
			Victim:           getProfile(k.VictimID),
			PhotoBase64:      k.PhotoBase64,
			CreatedAt:        k.CreatedAt,
			LikeCount:        likeCount,
			TargetAssignedAt: targetAssignedAt,
		}
		if k.PropID != nil {
			if prop, err := s.propRepo.GetPropByID(*k.PropID); err == nil {
				node.Prop = &PropSummary{ID: prop.ID, NameEN: prop.NameEN, NameNL: prop.NameNL}
			}
		}
		killNodes = append(killNodes, node)
	}

	awards := buildAwards(killNodes)

	return &EndScreen{
		GameID:             game.ID,
		Winner:             winner,
		WinnerKillCount:    winnerKillCount,
		PrizePhotoBase64:   game.PrizePhotoBase64,
		PrizeDescriptionEN: game.PrizeDescriptionEN,
		PrizeDescriptionNL: game.PrizeDescriptionNL,
		Stats: EndScreenStats{
			TotalKills:        len(allKills),
			TotalParticipants: len(participants),
			FastestKillSecs:   fastestSecs,
			MostKillsName:     mostKillsName,
			MostKillsCount:    mostKillsCount,
		},
		Kills:  killNodes,
		Awards: awards,
	}, nil
}

func buildAwards(kills []EndScreenKillNode) []GameAward {
	if len(kills) == 0 {
		return nil
	}

	sortedKills := make([]EndScreenKillNode, len(kills))
	copy(sortedKills, kills)
	sortByTime(sortedKills)

	intPtr := func(v int) *int { return &v }

	// Build helper maps
	killCountMap := map[uuid.UUID]int{}
	for _, k := range kills {
		killCountMap[k.Hunter.ID]++
	}

	profileMap := map[uuid.UUID]ProfileSummary{}
	for _, k := range kills {
		profileMap[k.Hunter.ID] = k.Hunter
		profileMap[k.Victim.ID] = k.Victim
	}

	killsByHunter := map[uuid.UUID][]EndScreenKillNode{}
	for _, k := range kills {
		killsByHunter[k.Hunter.ID] = append(killsByHunter[k.Hunter.ID], k)
	}

	var awards []GameAward

	// First Blood – hunter of the first kill
	firstBloodHunter := sortedKills[0].Hunter
	awards = append(awards, GameAward{
		ID:             "first-blood",
		Category:       AwardCategorySkill,
		TitleKey:       "gotcha.awards.firstBlood.title",
		DescriptionKey: "gotcha.awards.firstBlood.desc",
		Profile:        &firstBloodHunter,
	})

	// Serial Killer – most kills overall
	var serialKillerID uuid.UUID
	serialKillerCount := 0
	for id, cnt := range killCountMap {
		if cnt > serialKillerCount {
			serialKillerCount = cnt
			serialKillerID = id
		}
	}
	if serialKillerID != uuid.Nil {
		p := profileMap[serialKillerID]
		awards = append(awards, GameAward{
			ID:             "serial-killer",
			Category:       AwardCategorySkill,
			TitleKey:       "gotcha.awards.serialKiller.title",
			DescriptionKey: "gotcha.awards.serialKiller.desc",
			Profile:        &p,
			Count:          intPtr(serialKillerCount),
		})
	}

	fastestGap := time.Duration(1<<63 - 1)
	slowestGap := time.Duration(0)
	var speedDemonProfile *ProfileSummary
	var patientHunterProfile *ProfileSummary

	for id, hKills := range killsByHunter {
		if len(hKills) < 2 {
			continue
		}
		sortByTime(hKills)
		for i := 1; i < len(hKills); i++ {
			gap := hKills[i].CreatedAt.Sub(hKills[i-1].CreatedAt)
			if gap < fastestGap {
				fastestGap = gap
				p := profileMap[id]
				speedDemonProfile = &p
			}
			if gap > slowestGap {
				slowestGap = gap
				p := profileMap[id]
				patientHunterProfile = &p
			}
		}
	}

	if speedDemonProfile != nil {
		mins := int(fastestGap.Minutes())
		if mins < 1 {
			mins = 1
		}
		awards = append(awards, GameAward{
			ID:             "speed-demon",
			Category:       AwardCategorySkill,
			TitleKey:       "gotcha.awards.speedDemon.title",
			DescriptionKey: "gotcha.awards.speedDemon.desc",
			Profile:        speedDemonProfile,
			Count:          intPtr(mins),
		})
	}

	if patientHunterProfile != nil {
		hours := int(slowestGap.Hours())
		awards = append(awards, GameAward{
			ID:             "patient-hunter",
			Category:       AwardCategorySkill,
			TitleKey:       "gotcha.awards.patientHunter.title",
			DescriptionKey: "gotcha.awards.patientHunter.desc",
			Profile:        patientHunterProfile,
			Count:          intPtr(hours),
		})
	}

	// Best Disguise – kill with the most likes
	var mostLikedKill *EndScreenKillNode
	for i := range kills {
		if mostLikedKill == nil || kills[i].LikeCount > mostLikedKill.LikeCount {
			mostLikedKill = &kills[i]
		}
	}
	if mostLikedKill != nil && mostLikedKill.LikeCount > 0 {
		p := mostLikedKill.Hunter
		awards = append(awards, GameAward{
			ID:             "best-disguise",
			Category:       AwardCategorySocial,
			TitleKey:       "gotcha.awards.bestDisguise.title",
			DescriptionKey: "gotcha.awards.bestDisguise.desc",
			Profile:        &p,
			Count:          intPtr(mostLikedKill.LikeCount),
		})
	}

	// Deadliest Weapon – most-used prop (EN name stored; frontend can look up NL if needed)
	type propCount struct {
		nameEN string
		nameNL string
		count  int
	}
	propCounts := map[uuid.UUID]*propCount{}
	for _, k := range kills {
		if k.Prop == nil {
			continue
		}
		if _, ok := propCounts[k.Prop.ID]; !ok {
			propCounts[k.Prop.ID] = &propCount{nameEN: k.Prop.NameEN, nameNL: k.Prop.NameNL}
		}
		propCounts[k.Prop.ID].count++
	}
	var bestProp *propCount
	for _, pc := range propCounts {
		if bestProp == nil || pc.count > bestProp.count {
			bestProp = pc
		}
	}
	if bestProp != nil {
		awards = append(awards, GameAward{
			ID:             "deadliest-weapon",
			Category:       AwardCategoryProp,
			TitleKey:       "gotcha.awards.deadliestWeapon.title",
			DescriptionKey: "gotcha.awards.deadliestWeapon.desc",
			PropName:       bestProp.nameEN,
			Count:          intPtr(bestProp.count),
		})
	}

	// First Victim – victim of the first kill
	firstVictim := sortedKills[0].Victim
	awards = append(awards, GameAward{
		ID:             "first-victim",
		Category:       AwardCategoryMeme,
		TitleKey:       "gotcha.awards.firstVictim.title",
		DescriptionKey: "gotcha.awards.firstVictim.desc",
		Profile:        &firstVictim,
	})

	// Unlucky – eliminated within 2 hours of the very first kill
	firstKillTime := sortedKills[0].CreatedAt
	var unluckyProfiles []ProfileSummary
	for i := 1; i < len(sortedKills); i++ {
		if sortedKills[i].CreatedAt.Sub(firstKillTime) < 2*time.Hour {
			unluckyProfiles = append(unluckyProfiles, sortedKills[i].Victim)
		}
	}
	if len(unluckyProfiles) > 0 {
		awards = append(awards, GameAward{
			ID:             "unlucky",
			Category:       AwardCategoryMeme,
			TitleKey:       "gotcha.awards.unlucky.title",
			DescriptionKey: "gotcha.awards.unlucky.desc",
			Profiles:       unluckyProfiles,
		})
	}

	// AFK Victim – participated but never made a kill before being eliminated
	hunterIDs := map[uuid.UUID]bool{}
	for _, k := range kills {
		hunterIDs[k.Hunter.ID] = true
	}
	var afkProfiles []ProfileSummary
	seen := map[uuid.UUID]bool{}
	for _, p := range profileMap {
		if !hunterIDs[p.ID] && !seen[p.ID] {
			afkProfiles = append(afkProfiles, p)
			seen[p.ID] = true
		}
	}
	if len(afkProfiles) > 0 {
		awards = append(awards, GameAward{
			ID:             "afk-victim",
			Category:       AwardCategoryMeme,
			TitleKey:       "gotcha.awards.afkVictim.title",
			DescriptionKey: "gotcha.awards.afkVictim.desc",
			Profiles:       afkProfiles,
		})
	}

	// Final Victim – victim of the last kill
	finalVictim := sortedKills[len(sortedKills)-1].Victim
	awards = append(awards, GameAward{
		ID:             "final-victim",
		Category:       AwardCategoryGame,
		TitleKey:       "gotcha.awards.finalVictim.title",
		DescriptionKey: "gotcha.awards.finalVictim.desc",
		Profile:        &finalVictim,
	})

	// Bloodiest Day – calendar day with the most kills
	dayCounts := map[string]int{}
	for _, k := range kills {
		day := k.CreatedAt.Format("2 Jan") // e.g. "3 Mar"
		dayCounts[day]++
	}
	bestDay := ""
	bestDayCount := 0
	for day, cnt := range dayCounts {
		if cnt > bestDayCount {
			bestDayCount = cnt
			bestDay = day
		}
	}
	if bestDay != "" {
		awards = append(awards, GameAward{
			ID:             "bloodiest-day",
			Category:       AwardCategoryGame,
			TitleKey:       "gotcha.awards.bloodiestDay.title",
			DescriptionKey: "gotcha.awards.bloodiestDay.desc",
			Day:            bestDay,
			Count:          intPtr(bestDayCount),
		})
	}

	return awards
}

func sortByTime(kills []EndScreenKillNode) {
	for i := 1; i < len(kills); i++ {
		for j := i; j > 0 && kills[j].CreatedAt.Before(kills[j-1].CreatedAt); j-- {
			kills[j], kills[j-1] = kills[j-1], kills[j]
		}
	}
}

// History

func (s *gotchaService) GetGameHistory(profileID uuid.UUID) ([]*GameSummary, error) {
	campus, err := s.campusFor(profileID)
	if err != nil {
		return nil, fmt.Errorf("could not determine campus: %w", err)
	}
	games, err := s.gameRepo.GetGameHistoryByCampus(campus)
	if err != nil {
		return nil, err
	}

	type cached struct{ summary ProfileSummary }
	profileCache := map[uuid.UUID]*cached{}
	getProfile := func(id uuid.UUID) *ProfileSummary {
		if c, ok := profileCache[id]; ok {
			return &c.summary
		}
		p, err := s.profileService.GetProfileById(id)
		e := &cached{}
		if err == nil {
			e.summary = ProfileSummary{ID: p.ID, FirstName: p.FirstName, LastName: p.LastName, ProfilePicture: p.CustomProfilePicture}
		} else {
			e.summary = ProfileSummary{ID: id, FirstName: "Unknown"}
		}
		profileCache[id] = e
		return &e.summary
	}

	summaries := make([]*GameSummary, 0, len(games))
	for _, g := range games {
		participants, _ := s.participantRepo.GetParticipantsByGame(g.ID)
		kills, _ := s.killRepo.GetApprovedKillsByGame(g.ID)

		killCounts := map[uuid.UUID]int{}
		for _, k := range kills {
			killCounts[k.HunterID]++
		}

		summary := &GameSummary{
			ID:                 g.ID,
			Campus:             g.Campus,
			Status:             g.Status,
			StartDate:          g.StartDate,
			UpdatedAt:          g.UpdatedAt,
			WinnerID:           g.WinnerID,
			TotalParticipants:  len(participants),
			TotalKills:         len(kills),
			PrizeDescriptionEN: g.PrizeDescriptionEN,
			PrizeDescriptionNL: g.PrizeDescriptionNL,
		}
		if g.WinnerID != nil {
			summary.Winner = getProfile(*g.WinnerID)
			summary.WinnerKillCount = killCounts[*g.WinnerID]
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

// Hydrate kills

func (s *gotchaService) hydrateKills(kills []*domain.GotchaKill, requestingProfileID uuid.UUID) ([]*KillFeedItem, error) {
	type cached struct{ summary ProfileSummary }
	profileCache := map[uuid.UUID]*cached{}
	getProfile := func(id uuid.UUID) ProfileSummary {
		if c, ok := profileCache[id]; ok {
			return c.summary
		}
		p, err := s.profileService.GetProfileById(id)
		e := &cached{}
		if err == nil {
			e.summary = ProfileSummary{ID: p.ID, FirstName: p.FirstName, LastName: p.LastName, ProfilePicture: p.CustomProfilePicture}
		} else {
			e.summary = ProfileSummary{ID: id, FirstName: "Unknown"}
		}
		profileCache[id] = e
		return e.summary
	}

	items := make([]*KillFeedItem, 0, len(kills))
	for _, k := range kills {
		item := &KillFeedItem{
			ID:          k.ID,
			GameID:      k.GameID,
			PhotoBase64: k.PhotoBase64,
			Status:      k.Status,
			CreatedAt:   k.CreatedAt,
			ReviewedAt:  k.ReviewedAt,
			Hunter:      getProfile(k.HunterID),
			Victim:      getProfile(k.VictimID),
			LikeCount:   len(k.Likes),
		}
		if k.PropID != nil {
			if prop, err := s.propRepo.GetPropByID(*k.PropID); err == nil {
				item.Prop = &PropSummary{ID: prop.ID, NameEN: prop.NameEN, NameNL: prop.NameNL}
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

// Likes

func (s *gotchaService) LikeKill(killID, profileID uuid.UUID) error {
	return s.killRepo.SaveKillLike(&domain.GotchaKillLike{KillID: killID, ProfileID: profileID, LikedAt: time.Now()})
}

func (s *gotchaService) UnlikeKill(killID, profileID uuid.UUID) error {
	return s.killRepo.DeleteKillLike(killID, profileID)
}

// Status

func (s *gotchaService) GetMyStatus(profileID uuid.UUID) (*domain.Participant, error) {
	campus, err := s.campusFor(profileID)
	if err != nil {
		return nil, fmt.Errorf("could not determine campus: %w", err)
	}
	game, err := s.gameRepo.GetActiveGameByCampus(campus)
	if err != nil {
		return nil, err
	}
	return s.participantRepo.GetParticipant(game.ID, profileID)
}

func (s *gotchaService) GetLeaderboard(profileID uuid.UUID) ([]*domain.Participant, error) {
	campus, err := s.campusFor(profileID)
	if err != nil {
		return nil, fmt.Errorf("could not determine campus: %w", err)
	}
	game, err := s.gameRepo.GetActiveGameByCampus(campus)
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
		for i := range game.Participants {
			_ = s.participantRepo.SaveParticipant(&game.Participants[i])
		}
		_ = s.gameRepo.SaveGame(game)
	}
	return nil
}

// Prop management

func (s *gotchaService) CreateProp(profileID uuid.UUID, nameEN, nameNL string) (*domain.GotchaProp, error) {
	campus, err := s.campusFor(profileID)
	if err != nil {
		return nil, fmt.Errorf("could not determine campus: %w", err)
	}
	game, err := s.gameRepo.GetActiveGameByCampus(campus)
	if err != nil {
		game, err = s.CreateGame(campus, time.Time{}, 72, "", "", "")
		if err != nil {
			return nil, fmt.Errorf("failed to create game for prop: %w", err)
		}
	}
	if game.Status != domain.StatusOptIn {
		return nil, fmt.Errorf("cannot add props after game has started")
	}
	prop := &domain.GotchaProp{
		ID:     uuid.New(),
		GameID: game.ID,
		NameEN: nameEN,
		NameNL: nameNL,
	}
	if err := s.propRepo.SaveProp(prop); err != nil {
		return nil, err
	}
	return prop, nil
}

func (s *gotchaService) UpdateProp(id uuid.UUID, nameEN, nameNL string) (*domain.GotchaProp, error) {
	prop, err := s.propRepo.GetPropByID(id)
	if err != nil {
		return nil, err
	}
	game, err := s.gameRepo.GetGameByID(prop.GameID)
	if err == nil && game.Status != domain.StatusOptIn {
		return nil, fmt.Errorf("cannot update props after game has started")
	}
	prop.NameEN = nameEN
	prop.NameNL = nameNL
	if err := s.propRepo.SaveProp(prop); err != nil {
		return nil, err
	}
	return prop, nil
}

func (s *gotchaService) DeleteProp(id uuid.UUID) error {
	prop, err := s.propRepo.GetPropByID(id)
	if err != nil {
		return err
	}
	game, err := s.gameRepo.GetGameByID(prop.GameID)
	if err == nil && game.Status != domain.StatusOptIn {
		return fmt.Errorf("cannot delete props after game has started")
	}
	return s.propRepo.DeleteProp(id)
}

func (s *gotchaService) GetAllPropsByGame(profileID uuid.UUID) ([]*domain.GotchaProp, error) {
	campus, err := s.campusFor(profileID)
	if err != nil {
		return []*domain.GotchaProp{}, nil
	}
	game, err := s.gameRepo.GetActiveGameByCampus(campus)
	if err != nil {
		return []*domain.GotchaProp{}, nil
	}
	return s.propRepo.GetPropsByGame(game.ID)
}
