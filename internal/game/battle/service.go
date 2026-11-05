package battle

import (
	"context"
	"math/rand"
	"time"

	"gwent-bot/internal/database"
	"gwent-bot/internal/game/cards"
	"gwent-bot/pkg/models"
)

type Service struct {
	battleRepo  *database.BattleRepo
	cardService *cards.Service
	playerRepo  *database.PlayerRepo
}

func NewService(battleRepo *database.BattleRepo, cardService *cards.Service, playerRepo *database.PlayerRepo) *Service {
	return &Service{
		battleRepo:  battleRepo,
		cardService: cardService,
		playerRepo:  playerRepo,
	}
}

func (s *Service) CreatePvEBattle(ctx context.Context, playerID int64, difficulty string, playerDeck []string, playerLeaderID string) (*models.Battle, error) {
	player, err := s.playerRepo.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if player == nil {
		return nil, ErrPlayerNotFound
	}

	playerCards, err := s.getPlayerDeckCards(ctx, player, playerDeck)
	if err != nil {
		return nil, err
	}

	botCards, err := s.createBotDeck(ctx, difficulty)
	if err != nil {
		return nil, err
	}

	shuffleCards(playerCards)
	shuffleCards(botCards)

	playerHand := derefCardInstances(playerCards[:10])
	playerDeckRemaining := derefCardInstances(playerCards[10:])
	botHand := derefCardInstances(botCards[:10])
	botDeckRemaining := derefCardInstances(botCards[10:])

	battle := &models.Battle{
		ID:              generateBattleID(),
		Type:            models.BattleTypePvE,
		Phase:           models.PhaseMulligan,
		CurrentRound:    1,
		CurrentTurn:     playerID,
		Player1ID:       playerID,
		Player2ID:       0,
		Player1Hand:     playerHand,
		Player2Hand:     botHand,
		Player1Deck:     playerDeckRemaining,
		Player2Deck:     botDeckRemaining,
		Player1Graveyard: []models.CardInstance{},
		Player2Graveyard: []models.CardInstance{},
		Player1Board: models.BoardState{
			Melee:   []models.CardInstance{},
			Ranged:  []models.CardInstance{},
			Siege:   []models.CardInstance{},
			TotalPower: 0,
		},
		Player2Board: models.BoardState{
			Melee:   []models.CardInstance{},
			Ranged:  []models.CardInstance{},
			Siege:   []models.CardInstance{},
			TotalPower: 0,
		},
		Player1LeaderUsed: false,
		Player2LeaderUsed: false,
		Player1Passed:     false,
		Player2Passed:     false,
		Weather: models.WeatherState{
			Active:   "clear",
			Duration: 0,
			Row:      "",
		},
		BotDifficulty: difficulty,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = s.battleRepo.CreateBattle(ctx, battle)
	if err != nil {
		return nil, err
	}

	return battle, nil
}

func (s *Service) CreatePvPBattle(ctx context.Context, player1ID, player2ID int64) (*models.Battle, error) {
	player1, err := s.playerRepo.GetPlayerByID(ctx, player1ID)
	if err != nil {
		return nil, err
	}
	player2, err := s.playerRepo.GetPlayerByID(ctx, player2ID)
	if err != nil {
		return nil, err
	}
	if player1 == nil || player2 == nil {
		return nil, ErrPlayerNotFound
	}

	player1Cards, err := s.getPlayerDeckCards(ctx, player1, player1.Deck)
	if err != nil {
		return nil, err
	}
	player2Cards, err := s.getPlayerDeckCards(ctx, player2, player2.Deck)
	if err != nil {
		return nil, err
	}

	shuffleCards(player1Cards)
	shuffleCards(player2Cards)

	player1Hand := derefCardInstances(player1Cards[:10])
	player1Deck := derefCardInstances(player1Cards[10:])
	player2Hand := derefCardInstances(player2Cards[:10])
	player2Deck := derefCardInstances(player2Cards[10:])

	battle := &models.Battle{
		ID:              generateBattleID(),
		Type:            models.BattleTypePvP,
		Phase:           models.PhaseMulligan,
		CurrentRound:    1,
		CurrentTurn:     player1ID,
		Player1ID:       player1ID,
		Player2ID:       player2ID,
		Player1Hand:     player1Hand,
		Player2Hand:     player2Hand,
		Player1Deck:     player1Deck,
		Player2Deck:     player2Deck,
		Player1Graveyard: []models.CardInstance{},
		Player2Graveyard: []models.CardInstance{},
		Player1Board: models.BoardState{
			Melee:   []models.CardInstance{},
			Ranged:  []models.CardInstance{},
			Siege:   []models.CardInstance{},
			TotalPower: 0,
		},
		Player2Board: models.BoardState{
			Melee:   []models.CardInstance{},
			Ranged:  []models.CardInstance{},
			Siege:   []models.CardInstance{},
			TotalPower: 0,
		},
		Player1LeaderUsed: false,
		Player2LeaderUsed: false,
		Player1Passed:     false,
		Player2Passed:     false,
		Weather: models.WeatherState{
			Active:   "clear",
			Duration: 0,
			Row:      "",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.battleRepo.CreateBattle(ctx, battle)
	if err != nil {
		return nil, err
	}

	return battle, nil
}

func (s *Service) PlayCard(ctx context.Context, battleID, instanceID, row string, playerID int64) error {
	battle, err := s.battleRepo.GetBattle(ctx, battleID)
	if err != nil {
		return err
	}
	if battle == nil {
		return ErrBattleNotFound
	}

	if battle.Phase == models.PhaseFinished {
		return ErrBattleFinished
	}

	isPlayer1 := battle.Player1ID == playerID
	if !isPlayer1 && battle.Player2ID != playerID {
		return ErrNotYourBattle
	}

	if battle.CurrentTurn != playerID {
		return ErrNotYourTurn
	}

	var hand *[]models.CardInstance
	var board *models.BoardState
	if isPlayer1 {
		hand = &battle.Player1Hand
		board = &battle.Player1Board
	} else {
		hand = &battle.Player2Hand
		board = &battle.Player2Board
	}

	cardIndex := -1
	for i, c := range *hand {
		if c.InstanceID == instanceID {
			cardIndex = i
			break
		}
	}
	if cardIndex == -1 {
		return ErrCardNotInHand
	}

	card := (*hand)[cardIndex]
	card.Row = models.RowType(row)

	*hand = append((*hand)[:cardIndex], (*hand)[cardIndex+1:]...)

	s.addCardToBoard(board, card)
	s.applyDeployAbilities(battle, &card, isPlayer1)
	s.recalculateBoardPower(board)

	battle.UpdatedAt = time.Now()
	return s.battleRepo.UpdateBattle(ctx, battle)
}

func (s *Service) Pass(ctx context.Context, battleID string, playerID int64) error {
	battle, err := s.battleRepo.GetBattle(ctx, battleID)
	if err != nil {
		return err
	}
	if battle == nil {
		return ErrBattleNotFound
	}

	isPlayer1 := battle.Player1ID == playerID
	if isPlayer1 {
		battle.Player1Passed = true
	} else {
		battle.Player2Passed = true
	}

	if battle.Player1Passed && battle.Player2Passed {
		s.endRound(battle)
	} else {
		if isPlayer1 {
			battle.CurrentTurn = battle.Player2ID
		} else {
			battle.CurrentTurn = battle.Player1ID
		}
	}

	battle.UpdatedAt = time.Now()
	return s.battleRepo.UpdateBattle(ctx, battle)
}

func (s *Service) UseLeaderAbility(ctx context.Context, battleID string, playerID int64) error {
	battle, err := s.battleRepo.GetBattle(ctx, battleID)
	if err != nil {
		return err
	}
	if battle == nil {
		return ErrBattleNotFound
	}

	isPlayer1 := battle.Player1ID == playerID
	if isPlayer1 {
		if battle.Player1LeaderUsed {
			return ErrLeaderAlreadyUsed
		}
		battle.Player1LeaderUsed = true
		s.applyLeaderEffect(battle, true)
	} else {
		if battle.Player2LeaderUsed {
			return ErrLeaderAlreadyUsed
		}
		battle.Player2LeaderUsed = true
		s.applyLeaderEffect(battle, false)
	}

	battle.UpdatedAt = time.Now()
	return s.battleRepo.UpdateBattle(ctx, battle)
}

func (s *Service) BotTurn(ctx context.Context, battleID string) error {
	battle, err := s.battleRepo.GetBattle(ctx, battleID)
	if err != nil {
		return err
	}
	if battle == nil || battle.Type != models.BattleTypePvE {
		return ErrInvalidBattle
	}

	if len(battle.Player2Hand) == 0 {
		return s.Pass(ctx, battleID, 0)
	}

	bestIdx := 0
	bestPower := 0
	for i, card := range battle.Player2Hand {
		if card.Power > bestPower {
			bestPower = card.Power
			bestIdx = i
		}
	}

	card := battle.Player2Hand[bestIdx]
	row := string(card.Row)
	if row == "" {
		row = "melee"
	}

	return s.PlayCard(ctx, battleID, card.InstanceID, row, 0)
}

func (s *Service) GetBattle(ctx context.Context, battleID string) (*models.Battle, error) {
	return s.battleRepo.GetBattle(ctx, battleID)
}

func (s *Service) GetActiveBattle(ctx context.Context, playerID int64) (*models.Battle, error) {
	return s.battleRepo.GetActiveBattleForPlayer(ctx, playerID)
}

func (s *Service) getPlayerDeckCards(ctx context.Context, player *models.Player, deckIDs []string) ([]*models.CardInstance, error) {
	var instances []*models.CardInstance
	for _, cardID := range deckIDs {
		card, err := s.cardService.GetCard(ctx, cardID)
		if err != nil || card == nil {
			continue
		}
		instances = append(instances, s.cardService.CreateCardInstance(card, player.ID))
	}
	return instances, nil
}

func (s *Service) createBotDeck(ctx context.Context, difficulty string) ([]*models.CardInstance, error) {
	nations := []models.Nation{
		models.NationNorthernRealms,
		models.NationNilfgaard,
		models.NationScoiatael,
		models.NationMonsters,
		models.NationSyndicate,
	}
	rarityWeights := map[models.Rarity]int{
		models.RarityCommon:    50,
		models.RarityRare:      30,
		models.RarityEpic:      15,
		models.RarityLegendary: 5,
	}

	cards, err := s.cardService.GetRandomCards(ctx, nations, rarityWeights, 25)
	if err != nil {
		return nil, err
	}

	var instances []*models.CardInstance
	for _, card := range cards {
		instances = append(instances, s.cardService.CreateCardInstance(card, 0))
	}
	return instances, nil
}

func (s *Service) addCardToBoard(board *models.BoardState, card models.CardInstance) {
	switch card.Row {
	case models.RowMelee:
		board.Melee = append(board.Melee, card)
	case models.RowRanged:
		board.Ranged = append(board.Ranged, card)
	case models.RowSiege:
		board.Siege = append(board.Siege, card)
	}
}

func (s *Service) recalculateBoardPower(board *models.BoardState) {
	total := 0
	for _, card := range board.Melee {
		total += card.Power
	}
	for _, card := range board.Ranged {
		total += card.Power
	}
	for _, card := range board.Siege {
		total += card.Power
	}
	board.TotalPower = total
}

func (s *Service) applyDeployAbilities(battle *models.Battle, card *models.CardInstance, isPlayer1 bool) {
	_ = battle
	_ = isPlayer1
}

func (s *Service) applyLeaderEffect(battle *models.Battle, isPlayer1 bool) {
	_ = battle
	_ = isPlayer1
}

func (s *Service) endRound(battle *models.Battle) {
	p1Power := battle.Player1Board.TotalPower
	p2Power := battle.Player2Board.TotalPower

	var roundWinner int64
	if p1Power > p2Power {
		roundWinner = battle.Player1ID
	} else if p2Power > p1Power {
		roundWinner = battle.Player2ID
	} else {
		roundWinner = battle.Player1ID
	}

	battle.CurrentRound++
	battle.Player1Passed = false
	battle.Player2Passed = false

	if battle.CurrentRound > 3 {
		battle.Phase = models.PhaseFinished
		now := time.Now()
		battle.FinishedAt = &now
		if p1Power >= p2Power {
			battle.WinnerID = &battle.Player1ID
		} else {
			battle.WinnerID = &battle.Player2ID
		}
		return
	}

	s.drawCards(battle, true, 2)
	s.drawCards(battle, false, 2)

	battle.CurrentTurn = roundWinner
	battle.Phase = models.GamePhase("round_" + string(rune('1'+battle.CurrentRound-1)))
}

func (s *Service) drawCards(battle *models.Battle, isPlayer1 bool, count int) {
	var hand *[]models.CardInstance
	var deck *[]models.CardInstance
	if isPlayer1 {
		hand = &battle.Player1Hand
		deck = &battle.Player1Deck
	} else {
		hand = &battle.Player2Hand
		deck = &battle.Player2Deck
	}

	for i := 0; i < count && len(*deck) > 0 && len(*hand) < 10; i++ {
		card := (*deck)[0]
		*deck = (*deck)[1:]
		*hand = append(*hand, card)
	}
}

func shuffleCards(cards []*models.CardInstance) {
	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
}

func generateBattleID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func derefCardInstances(ptrSlice []*models.CardInstance) []models.CardInstance {
	result := make([]models.CardInstance, len(ptrSlice))
	for i, ptr := range ptrSlice {
		if ptr != nil {
			result[i] = *ptr
		}
	}
	return result
}

var (
	ErrPlayerNotFound     = &BattleError{"player not found"}
	ErrBattleNotFound     = &BattleError{"battle not found"}
	ErrBattleFinished     = &BattleError{"battle already finished"}
	ErrNotYourBattle      = &BattleError{"not your battle"}
	ErrNotYourTurn        = &BattleError{"not your turn"}
	ErrCardNotInHand      = &BattleError{"card not in hand"}
	ErrLeaderAlreadyUsed  = &BattleError{"leader ability already used"}
	ErrInvalidBattle      = &BattleError{"invalid battle"}
)

type BattleError struct {
	msg string
}

func (e *BattleError) Error() string {
	return e.msg
}
