package progression

import (
	"context"
	"math/rand"
	"time"

	"gwent-bot/internal/database"
	"gwent-bot/internal/game/cards"
	"gwent-bot/internal/config"
	"gwent-bot/pkg/models"
)

type Service struct {
	playerRepo  *database.PlayerRepo
	cardService *cards.Service
	cfg         *config.Config
}

func NewService(playerRepo *database.PlayerRepo, cardService *cards.Service, cfg *config.Config) *Service {
	return &Service{
		playerRepo:  playerRepo,
		cardService: cardService,
		cfg:         cfg,
	}
}

func (s *Service) ClaimDailyCoin(ctx context.Context, playerID int64) (int, error) {
	player, err := s.playerRepo.GetPlayerByID(ctx, playerID)
	if err != nil {
		return 0, err
	}
	if player == nil {
		return 0, ErrPlayerNotFound
	}

	now := time.Now()
	if player.DailyCoinClaimedAt != nil {
		nextClaim := player.DailyCoinClaimedAt.Add(s.cfg.DailyCoinCooldown)
		if now.Before(nextClaim) {
			return 0, ErrDailyCoinOnCooldown
		}
	}

	reward := s.cfg.StartingGold + (int(player.Level)-1)*10

	err = s.playerRepo.AddGold(ctx, playerID, reward)
	if err != nil {
		return 0, err
	}

	err = s.playerRepo.ClaimDailyCoin(ctx, playerID)
	if err != nil {
		return 0, err
	}

	return reward, nil
}

func (s *Service) ClaimAxe(ctx context.Context, playerID int64) (int, error) {
	player, err := s.playerRepo.GetPlayerByID(ctx, playerID)
	if err != nil {
		return 0, err
	}
	if player == nil {
		return 0, ErrPlayerNotFound
	}

	now := time.Now()
	if player.AxeClaimedAt != nil {
		nextAxe := player.AxeClaimedAt.Add(s.cfg.AxeCooldown)
		if now.Before(nextAxe) {
			return 0, ErrAxeOnCooldown
		}
	}

	durability := s.cfg.AxeDurability

	err = s.playerRepo.ClaimAxe(ctx, playerID, durability)
	if err != nil {
		return 0, err
	}

	return durability, nil
}

func (s *Service) UnlockMine(ctx context.Context, playerID int64) error {
	player, err := s.playerRepo.GetPlayerByID(ctx, playerID)
	if err != nil {
		return err
	}
	if player == nil {
		return ErrPlayerNotFound
	}

	if player.MineUnlocked {
		return ErrMineAlreadyUnlocked
	}

	cost := s.cfg.MineGoldPerExtract * 10
	err = s.playerRepo.SpendGold(ctx, playerID, cost)
	if err != nil {
		return err
	}

	return s.playerRepo.UnlockMine(ctx, playerID)
}

func (s *Service) ExtractMine(ctx context.Context, playerID int64) (int, int, error) {
	player, err := s.playerRepo.GetPlayerByID(ctx, playerID)
	if err != nil {
		return 0, 0, err
	}
	if player == nil {
		return 0, 0, ErrPlayerNotFound
	}

	if !player.MineUnlocked {
		return 0, 0, ErrMineLocked
	}

	if player.AxeDurability <= 0 {
		return 0, 0, ErrNoAxe
	}

	if player.MineExtractsToday >= s.cfg.MineMaxExtracts {
		return 0, 0, ErrMaxExtractsReached
	}

	baseGold := s.cfg.MineGoldPerExtract
	variance := rand.Intn(21) - 10
	goldEarned := baseGold + variance
	if goldEarned < 10 {
		goldEarned = 10
	}

	err = s.playerRepo.ExtractMineGold(ctx, playerID, goldEarned)
	if err != nil {
		return 0, 0, err
	}

	err = s.playerRepo.UseAxeDurability(ctx, playerID)
	if err != nil {
		return 0, 0, err
	}

	extractsLeft := s.cfg.MineMaxExtracts - player.MineExtractsToday - 1

	return goldEarned, extractsLeft, nil
}

func (s *Service) OpenChest(ctx context.Context, playerID int64, chestID string) ([]*models.Card, int, error) {
	player, err := s.playerRepo.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, 0, err
	}
	if player == nil {
		return nil, 0, ErrPlayerNotFound
	}

	chest := s.getChestByID(chestID)
	if chest == nil {
		return nil, 0, ErrChestNotFound
	}

	if chest.LevelRequired > player.Level {
		return nil, 0, ErrLevelTooLow
	}

	var rewardCards []*models.Card
	for _, reward := range chest.Cards {
		qty := rand.Intn(reward.MaxQty-reward.MinQty+1) + reward.MinQty
		for i := 0; i < qty; i++ {
			card, err := s.cardService.GetCard(ctx, reward.CardID)
			if err == nil && card != nil {
				rewardCards = append(rewardCards, card)
				s.playerRepo.AddCardToCollection(ctx, playerID, card.ID)
			}
		}
	}

	if chest.Gold > 0 {
		s.playerRepo.AddGold(ctx, playerID, chest.Gold)
	}

	return rewardCards, chest.Gold, nil
}

func (s *Service) AddXP(ctx context.Context, playerID int64, amount int) (bool, models.PlayerLevel, error) {
	player, err := s.playerRepo.GetPlayerByID(ctx, playerID)
	if err != nil {
		return false, 0, err
	}
	if player == nil {
		return false, 0, ErrPlayerNotFound
	}

	newXP := player.Experience + amount
	newLevel := calculateLevel(newXP)

	leveledUp := false
	if newLevel > player.Level {
		leveledUp = true
		err = s.playerRepo.LevelUp(ctx, playerID, newLevel)
		if err != nil {
			return false, 0, err
		}
	}

	err = s.playerRepo.IncrementExperience(ctx, playerID, amount)
	if err != nil {
		return false, 0, err
	}

	return leveledUp, newLevel, nil
}

func (s *Service) RecordBattleResult(ctx context.Context, playerID int64, won bool) error {
	player, err := s.playerRepo.GetPlayerByID(ctx, playerID)
	if err != nil {
		return err
	}
	if player == nil {
		return ErrPlayerNotFound
	}

	err = s.playerRepo.RecordBattleResult(ctx, playerID, won)
	if err != nil {
		return err
	}

	xpReward := 10
	if won {
		xpReward = 25
	}
	s.AddXP(ctx, playerID, xpReward)

	goldReward := 5
	if won {
		goldReward = 15
	}
	s.playerRepo.AddGold(ctx, playerID, goldReward)

	return nil
}

func (s *Service) GetStarterRewards(ctx context.Context, playerID int64, nation models.Nation) ([]*models.Card, error) {
	starterCards, err := s.cardService.GetStarterDeck(ctx, nation)
	if err != nil {
		return nil, err
	}

	for _, card := range starterCards {
		s.playerRepo.AddCardToCollection(ctx, playerID, card.ID)
	}

	return starterCards, nil
}

func (s *Service) getChestByID(chestID string) *models.Chest {
	chests := map[string]*models.Chest{
		"starter": {
			ID:            "starter",
			Name:          "Starter Chest",
			Description:   "Basic cards for new players",
			LevelRequired: 1,
			Cards: []models.ChestReward{
				{CardID: "nr_001", MinQty: 1, MaxQty: 2, Weight: 50},
				{CardID: "ng_001", MinQty: 1, MaxQty: 2, Weight: 50},
			},
			Gold:     50,
			Cooldown: 24 * time.Hour,
		},
		"level_1": {
			ID:            "level_1",
			Name:          "Novice Chest",
			Description:   "Cards for level 1+ players",
			LevelRequired: 1,
			Cards: []models.ChestReward{
				{CardID: "nr_002", MinQty: 1, MaxQty: 1, Weight: 40},
				{CardID: "st_001", MinQty: 1, MaxQty: 1, Weight: 30},
				{CardID: "mo_001", MinQty: 1, MaxQty: 1, Weight: 30},
			},
			Gold:     100,
			Cooldown: 12 * time.Hour,
		},
		"level_2": {
			ID:            "level_2",
			Name:          "Adept Chest",
			Description:   "Better cards for level 2+ players",
			LevelRequired: 2,
			Cards: []models.ChestReward{
				{CardID: "nr_003", MinQty: 1, MaxQty: 1, Weight: 30},
				{CardID: "sy_001", MinQty: 1, MaxQty: 1, Weight: 35},
				{CardID: "ng_002", MinQty: 1, MaxQty: 1, Weight: 35},
			},
			Gold:     200,
			Cooldown: 8 * time.Hour,
		},
	}

	return chests[chestID]
}

func calculateLevel(xp int) models.PlayerLevel {
	level := xp / 100
	if level < 1 {
		level = 1
	}
	if level > 30 {
		level = 30
	}
	return models.PlayerLevel(level)
}

var (
	ErrPlayerNotFound        = &ProgressionError{"player not found"}
	ErrDailyCoinOnCooldown   = &ProgressionError{"daily coin on cooldown"}
	ErrAxeOnCooldown         = &ProgressionError{"axe on cooldown"}
	ErrMineAlreadyUnlocked   = &ProgressionError{"mine already unlocked"}
	ErrMineLocked            = &ProgressionError{"mine is locked"}
	ErrNoAxe                 = &ProgressionError{"no axe available"}
	ErrMaxExtractsReached    = &ProgressionError{"max extracts reached today"}
	ErrChestNotFound         = &ProgressionError{"chest not found"}
	ErrLevelTooLow           = &ProgressionError{"level too low for this chest"}
)

type ProgressionError struct {
	msg string
}

func (e *ProgressionError) Error() string {
	return e.msg
}
