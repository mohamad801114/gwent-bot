package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"gwent-bot/pkg/models"
)

type PlayerRepo struct {
	db *DB
}

func NewPlayerRepo(db *DB) *PlayerRepo {
	return &PlayerRepo{db: db}
}

func (r *PlayerRepo) CreatePlayer(ctx context.Context, p *models.Player) error {
	collectionJSON, _ := json.Marshal(p.Collection)
	deckJSON, _ := json.Marshal(p.Deck)

	query := `
		INSERT INTO players (
			telegram_id, username, first_name, last_name, level, experience,
			gold, scrap, meteorite_powder, collection, deck, leader_id,
			daily_coin_claimed_at, axe_claimed_at, axe_durability, mine_unlocked, mine_extracts_today,
			last_mine_extract_at, games_played, games_won, games_lost, highest_rank,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)
		RETURNING id`

	err := r.db.Pool.QueryRow(ctx, query,
		p.TelegramID, p.Username, p.FirstName, p.LastName,
		p.Level, p.Experience,
		p.Gold, p.Scrap, p.MeteoritePowder,
		collectionJSON, deckJSON, p.LeaderID,
		p.DailyCoinClaimedAt, p.AxeClaimedAt, p.AxeDurability, p.MineUnlocked, p.MineExtractsToday,
		p.LastMineExtractAt, p.GamesPlayed, p.GamesWon, p.GamesLost, p.HighestRank,
		p.CreatedAt, p.UpdatedAt,
	).Scan(&p.ID)

	return err
}

func (r *PlayerRepo) GetPlayerByTelegramID(ctx context.Context, telegramID int64) (*models.Player, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, level, experience,
			gold, scrap, meteorite_powder, collection, deck, leader_id,
			daily_coin_claimed_at, axe_claimed_at, axe_durability, mine_unlocked,
			mine_extracts_today, last_mine_extract_at,
			games_played, games_won, games_lost, highest_rank,
			created_at, updated_at
		FROM players WHERE telegram_id = $1`

	var p models.Player
	var collectionJSON, deckJSON []byte
	var dailyCoinClaimedAt, axeClaimedAt, lastMineExtractAt pgtype.Timestamptz

	err := r.db.Pool.QueryRow(ctx, query, telegramID).Scan(
		&p.ID, &p.TelegramID, &p.Username, &p.FirstName, &p.LastName,
		&p.Level, &p.Experience,
		&p.Gold, &p.Scrap, &p.MeteoritePowder,
		&collectionJSON, &deckJSON, &p.LeaderID,
		&dailyCoinClaimedAt, &axeClaimedAt, &p.AxeDurability, &p.MineUnlocked,
		&p.MineExtractsToday, &lastMineExtractAt,
		&p.GamesPlayed, &p.GamesWon, &p.GamesLost, &p.HighestRank,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(collectionJSON) > 0 {
		json.Unmarshal(collectionJSON, &p.Collection)
	}
	if len(deckJSON) > 0 {
		json.Unmarshal(deckJSON, &p.Deck)
	}
	if dailyCoinClaimedAt.Valid {
		t := dailyCoinClaimedAt.Time
		p.DailyCoinClaimedAt = &t
	}
	if axeClaimedAt.Valid {
		t := axeClaimedAt.Time
		p.AxeClaimedAt = &t
	}
	if lastMineExtractAt.Valid {
		t := lastMineExtractAt.Time
		p.LastMineExtractAt = &t
	}

	return &p, nil
}

func (r *PlayerRepo) GetPlayerByID(ctx context.Context, id int64) (*models.Player, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, level, experience,
			gold, scrap, meteorite_powder, collection, deck, leader_id,
			daily_coin_claimed_at, axe_claimed_at, axe_durability, mine_unlocked,
			mine_extracts_today, last_mine_extract_at,
			games_played, games_won, games_lost, highest_rank,
			created_at, updated_at
		FROM players WHERE id = $1`

	var p models.Player
	var collectionJSON, deckJSON []byte
	var dailyCoinClaimedAt, axeClaimedAt, lastMineExtractAt pgtype.Timestamptz

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.TelegramID, &p.Username, &p.FirstName, &p.LastName,
		&p.Level, &p.Experience,
		&p.Gold, &p.Scrap, &p.MeteoritePowder,
		&collectionJSON, &deckJSON, &p.LeaderID,
		&dailyCoinClaimedAt, &axeClaimedAt, &p.AxeDurability, &p.MineUnlocked,
		&p.MineExtractsToday, &lastMineExtractAt,
		&p.GamesPlayed, &p.GamesWon, &p.GamesLost, &p.HighestRank,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(collectionJSON) > 0 {
		json.Unmarshal(collectionJSON, &p.Collection)
	}
	if len(deckJSON) > 0 {
		json.Unmarshal(deckJSON, &p.Deck)
	}
	if dailyCoinClaimedAt.Valid {
		t := dailyCoinClaimedAt.Time
		p.DailyCoinClaimedAt = &t
	}
	if axeClaimedAt.Valid {
		t := axeClaimedAt.Time
		p.AxeClaimedAt = &t
	}
	if lastMineExtractAt.Valid {
		t := lastMineExtractAt.Time
		p.LastMineExtractAt = &t
	}

	return &p, nil
}

func (r *PlayerRepo) UpdatePlayer(ctx context.Context, p *models.Player) error {
	collectionJSON, _ := json.Marshal(p.Collection)
	deckJSON, _ := json.Marshal(p.Deck)

	p.UpdatedAt = time.Now()

	query := `
		UPDATE players SET
			username = $2, first_name = $3, last_name = $4, level = $5, experience = $6,
			gold = $7, scrap = $8, meteorite_powder = $9, collection = $10, deck = $11,
			leader_id = $12, daily_coin_claimed_at = $13, axe_claimed_at = $14,
			axe_durability = $15, mine_unlocked = $16, mine_extracts_today = $17,
			last_mine_extract_at = $18, games_played = $19, games_won = $20,
			games_lost = $21, highest_rank = $22, updated_at = $23
		WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query,
		p.ID, p.Username, p.FirstName, p.LastName, p.Level, p.Experience,
		p.Gold, p.Scrap, p.MeteoritePowder,
		collectionJSON, deckJSON, p.LeaderID,
		p.DailyCoinClaimedAt, p.AxeClaimedAt, p.AxeDurability, p.MineUnlocked,
		p.MineExtractsToday, p.LastMineExtractAt,
		p.GamesPlayed, p.GamesWon, p.GamesLost, p.HighestRank,
		p.UpdatedAt,
	)
	return err
}

func (r *PlayerRepo) AddCardToCollection(ctx context.Context, playerID int64, cardID string) error {
	// This is a simplified approach - in production you would use proper JSONB operations
	// For now, fetch, modify, save
	p, err := r.GetPlayerByID(ctx, playerID)
	if err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("player not found")
	}

	// Check if already has card
	found := false
	for _, c := range p.Collection {
		if c.ID == cardID {
			found = true
			break
		}
	}
	if !found {
		p.Collection = append(p.Collection, models.Card{ID: cardID})
	}
	return r.UpdatePlayer(ctx, p)
}

func (r *PlayerRepo) AddGold(ctx context.Context, playerID int64, amount int) error {
	query := `UPDATE players SET gold = gold + $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, playerID, amount)
	return err
}

func (r *PlayerRepo) SpendGold(ctx context.Context, playerID int64, amount int) error {
	query := `UPDATE players SET gold = gold - $2, updated_at = NOW() WHERE id = $1 AND gold >= $2`
	result, err := r.db.Pool.Exec(ctx, query, playerID, amount)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("insufficient gold")
	}
	return nil
}

func (r *PlayerRepo) IncrementExperience(ctx context.Context, playerID int64, amount int) error {
	query := `UPDATE players SET experience = experience + $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, playerID, amount)
	return err
}

func (r *PlayerRepo) LevelUp(ctx context.Context, playerID int64, newLevel models.PlayerLevel) error {
	query := `UPDATE players SET level = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, playerID, newLevel)
	return err
}

func (r *PlayerRepo) UnlockMine(ctx context.Context, playerID int64) error {
	query := `UPDATE players SET mine_unlocked = true, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, playerID)
	return err
}

func (r *PlayerRepo) ClaimDailyCoin(ctx context.Context, playerID int64) error {
	now := time.Now()
	query := `UPDATE players SET daily_coin_claimed_at = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, playerID, now)
	return err
}

func (r *PlayerRepo) ClaimAxe(ctx context.Context, playerID int64, durability int) error {
	now := time.Now()
	query := `UPDATE players SET axe_claimed_at = $2, axe_durability = $3, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, playerID, now, durability)
	return err
}

func (r *PlayerRepo) UseAxeDurability(ctx context.Context, playerID int64) error {
	query := `UPDATE players SET axe_durability = axe_durability - 1, updated_at = NOW() WHERE id = $1 AND axe_durability > 0`
	result, err := r.db.Pool.Exec(ctx, query, playerID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("axe broken or not available")
	}
	return nil
}

func (r *PlayerRepo) ExtractMineGold(ctx context.Context, playerID int64, goldAmount int) error {
	now := time.Now()
	query := `
		UPDATE players SET
			gold = gold + $2,
			mine_extracts_today = mine_extracts_today + 1,
			last_mine_extract_at = $3,
			updated_at = NOW()
		WHERE id = $1 AND mine_unlocked = true AND mine_extracts_today < $4`
	result, err := r.db.Pool.Exec(ctx, query, playerID, goldAmount, now, 10) // Max 10 extracts
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("cannot extract: mine locked or max extracts reached")
	}
	return err
}

func (r *PlayerRepo) ResetDailyMineExtracts(ctx context.Context) error {
	query := `UPDATE players SET mine_extracts_today = 0 WHERE mine_unlocked = true`
	_, err := r.db.Pool.Exec(ctx, query)
	return err
}

func (r *PlayerRepo) RecordBattleResult(ctx context.Context, playerID int64, won bool) error {
	query := `
		UPDATE players SET
			games_played = games_played + 1,
			games_won = games_won + $2,
			games_lost = games_lost + $3,
			updated_at = NOW()
		WHERE id = $1`
	wonInt := 0
	lostInt := 0
	if won {
		wonInt = 1
	} else {
		lostInt = 1
	}
	_, err := r.db.Pool.Exec(ctx, query, playerID, wonInt, lostInt)
	return err
}
