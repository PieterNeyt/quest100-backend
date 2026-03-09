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
	ID     uuid.UUID
	NameEN string
	NameNL string
}

type KillFeedItem struct {
	ID          uuid.UUID
	GameID      uuid.UUID
	PhotoBase64 string
	Status      domain.KillStatus
	CreatedAt   time.Time
	ReviewedAt  *time.Time
	Hunter      ProfileSummary
	Victim      ProfileSummary
	Prop        *PropSummary
	LikeCount   int
	LikedByMe   bool
}

type TargetInfo struct {
	Target       *ProfileSummary
	AssignedProp *PropSummary
	KillDeadline *time.Time
}

type EndScreenKillNode struct {
	KillID           uuid.UUID
	Hunter           ProfileSummary
	Victim           ProfileSummary
	Prop             *PropSummary
	PhotoBase64      string
	CreatedAt        time.Time
	LikeCount        int
	TargetAssignedAt *time.Time
}

type EndScreenStats struct {
	TotalKills        int
	TotalParticipants int
	FastestKillSecs   int
	MostKillsName     string
	MostKillsCount    int
}

type EndScreen struct {
	Winner             *ProfileSummary
	WinnerKillCount    int
	PrizePhotoBase64   string
	PrizeDescriptionEN string
	PrizeDescriptionNL string
	Stats              EndScreenStats
	Kills              []EndScreenKillNode
}

type GotchaService interface {
	CreateGame(campus string, startDate time.Time, killDeadlineHours int, prizePhotoBase64, prizeDescEN, prizeDescNL string) (*domain.GotchaGame, error)
	StartGame(campus string) error
	GetCurrentGame(campus string) (*domain.GotchaGame, error)
	UpdateStartDate(campus string, startDate time.Time, killDeadlineHours int, prizePhotoBase64, prizeDescEN, prizeDescNL string) (*domain.GotchaGame, error)

	OptIn(campus string, profileID uuid.UUID) error
	OptOut(campus string, profileID uuid.UUID) error

	SubmitKill(campus string, hunterID uuid.UUID, photoBase64 string) (*domain.GotchaKill, error)
	ReviewKill(killID, reviewerID uuid.UUID, approve bool) error

	GetNextPendingKill(campus string, requestingProfileID uuid.UUID) (*KillFeedItem, error)
	GetPendingKillCount(campus string) (int, error)

	GetKillFeed(campus string, requestingProfileID uuid.UUID, limit, offset int) ([]*KillFeedItem, error)
	GetPendingKills(campus string, requestingProfileID uuid.UUID) ([]*KillFeedItem, error)
	LikeKill(killID, profileID uuid.UUID) error
	UnlikeKill(killID, profileID uuid.UUID) error

	GetMyStatus(campus string, profileID uuid.UUID) (*domain.Participant, error)
	GetTargetInfo(campus string, profileID uuid.UUID) (*TargetInfo, error)
	GetLeaderboard(campus string) ([]*domain.Participant, error)
	GetEndScreen(campus string) (*EndScreen, error)

	GetAllPropsByGame(campus string) ([]*domain.GotchaProp, error)
	CreateProp(campus string, nameEN, nameNL string) (*domain.GotchaProp, error)
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

//  Game

func (s *gotchaService) GetCurrentGame(campus string) (*domain.GotchaGame, error) {
	return s.gameRepo.GetGameByCampus(campus)
}

func (s *gotchaService) CreateGame(campus string, startDate time.Time, killDeadlineHours int, prizePhotoBase64, prizeDescEN, prizeDescNL string) (*domain.GotchaGame, error) {
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
	game, err := s.gameRepo.GetGameByCampus(campus)
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

//  Opt-in / out

func (s *gotchaService) OptIn(campus string, profileID uuid.UUID) error {
	game, err := s.gameRepo.GetGameByCampus(campus)
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

// Kills

func (s *gotchaService) SubmitKill(campus string, hunterID uuid.UUID, photoBase64 string) (*domain.GotchaKill, error) {
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
	now := time.Now()
	hunter.PendingKillAt = &now
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

	now := time.Now()
	hunter.PendingKillAt = nil

	if hunter.IsAlive && !hunter.KillDeadline.IsZero() && hunter.KillDeadline.Before(now) {

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

	hunter := game.FindParticipant(kill.HunterID)
	victim := game.FindParticipant(kill.VictimID)

	if hunter == nil || !hunter.IsAlive {
		kill.Status = domain.KillDenied
		return s.killRepo.SaveKill(kill)
	}

	if hunter.TargetID == nil || *hunter.TargetID != kill.VictimID {
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

	hunter.PendingKillAt = nil
	victim.PendingKillAt = nil

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

func (s *gotchaService) GetNextPendingKill(campus string, requestingProfileID uuid.UUID) (*KillFeedItem, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
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

func (s *gotchaService) GetPendingKillCount(campus string) (int, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
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

func (s *gotchaService) GetKillFeed(campus string, requestingProfileID uuid.UUID, limit, offset int) ([]*KillFeedItem, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return nil, err
	}
	kills, err := s.killRepo.GetApprovedKillFeed(game.ID, limit, offset)
	if err != nil {
		return nil, err
	}
	return s.hydrateKills(kills, requestingProfileID)
}

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

// Target info

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
			info.AssignedProp = &PropSummary{ID: prop.ID, NameEN: prop.NameEN, NameNL: prop.NameNL}
		}
	}
	if !participant.KillDeadline.IsZero() {
		info.KillDeadline = &participant.KillDeadline
	}
	return info, nil
}

// End screen

func (s *gotchaService) GetEndScreen(campus string) (*EndScreen, error) {
	game, err := s.gameRepo.GetGameByFinishedCampus(campus)
	if err != nil {
		return nil, fmt.Errorf("no finished game for campus %s: %w", campus, err)
	}
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
		p, err := s.profileRepo.GetProfileById(id)
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
		// Fetch full kill with likes preloaded to get LikeCount
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

	return &EndScreen{
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
		Kills: killNodes,
	}, nil
}

// Hydrate kills

func (s *gotchaService) hydrateKills(kills []*domain.GotchaKill, requestingProfileID uuid.UUID) ([]*KillFeedItem, error) {
	type cached struct{ summary ProfileSummary }
	profileCache := map[uuid.UUID]*cached{}
	getProfile := func(id uuid.UUID) ProfileSummary {
		if c, ok := profileCache[id]; ok {
			return c.summary
		}
		p, err := s.profileRepo.GetProfileById(id)
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

// Status / leaderboard

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

//  Timeouts

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

func (s *gotchaService) CreateProp(campus string, nameEN, nameNL string) (*domain.GotchaProp, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return nil, err
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

func (s *gotchaService) GetAllPropsByGame(campus string) ([]*domain.GotchaProp, error) {
	game, err := s.gameRepo.GetGameByCampus(campus)
	if err != nil {
		return nil, err
	}
	return s.propRepo.GetPropsByGame(game.ID)
}
