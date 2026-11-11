package database

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"gwent-bot/pkg/models"
)

type BattleRepo struct {
	db *DB
}

func NewBattleRepo(db *DB) *BattleRepo {
	return &BattleRepo{db: db}
}

func (r *BattleRepo) CreateBattle(ctx context.Context, b *models.Battle) error {
	p1BoardJSON, _ := json.Marshal(b.Player1Board)
	p2BoardJSON, _ := json.Marshal(b.Player2Board)
	p1HandJSON, _ := json.Marshal(b.Player1Hand)
	p2HandJSON, _ := json.Marshal(b.Player2Hand)
	p1DeckJSON, _ := json.Marshal(b.Player1Deck)
	p2DeckJSON, _ := json.Marshal(b.Player2Deck)
	p1GraveJSON, _ := json.Marshal(b.Player1Graveyard)
	p2GraveJSON, _ := json.Marshal(b.Player2Graveyard)
	weatherJSON, _ := json.Marshal(b.Weather)

	query := `
		INSERT INTO battles (
			id, type, phase, current_round, current_turn, player1_id, player2_id,
			player1_board, player2_board, player1_hand, player2_hand,
			player1_deck, player2_deck, player1_graveyard, player2_graveyard,
			player1_leader_used, player2_leader_used, player1_passed, player2_passed,
			weather, bot_difficulty, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)
		RETURNING id`

	_, err := r.db.Pool.Exec(ctx, query,
		b.ID, b.Type, b.Phase, b.CurrentRound, b.CurrentTurn,
		b.Player1ID, b.Player2ID,
		p1BoardJSON, p2BoardJSON, p1HandJSON, p2HandJSON,
		p1DeckJSON, p2DeckJSON, p1GraveJSON, p2GraveJSON,
		b.Player1LeaderUsed, b.Player2LeaderUsed, b.Player1Passed, b.Player2Passed,
		weatherJSON, b.BotDifficulty, b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (r *BattleRepo) GetBattle(ctx context.Context, battleID string) (*models.Battle, error) {
	query := `
		SELECT id, type, phase, current_round, current_turn, player1_id, player2_id, winner_id,
			player1_board, player2_board, player1_hand, player2_hand,
			player1_deck, player2_deck, player1_graveyard, player2_graveyard,
			player1_leader_used, player2_leader_used, player1_passed, player2_passed,
			weather, bot_difficulty, created_at, updated_at, finished_at
		FROM battles WHERE id = $1`

	var b models.Battle
	var p1BoardJSON, p2BoardJSON, p1HandJSON, p2HandJSON []byte
	var p1DeckJSON, p2DeckJSON, p1GraveJSON, p2GraveJSON, weatherJSON []byte
	var winnerID pgtype.Int8
	var botDifficulty pgtype.Text
	var finishedAt pgtype.Timestamptz

	err := r.db.Pool.QueryRow(ctx, query, battleID).Scan(
		&b.ID, &b.Type, &b.Phase, &b.CurrentRound, &b.CurrentTurn,
		&b.Player1ID, &b.Player2ID, &winnerID,
		&p1BoardJSON, &p2BoardJSON, &p1HandJSON, &p2HandJSON,
		&p1DeckJSON, &p2DeckJSON, &p1GraveJSON, &p2GraveJSON,
		&b.Player1LeaderUsed, &b.Player2LeaderUsed, &b.Player1Passed, &b.Player2Passed,
		&weatherJSON, &botDifficulty, &b.CreatedAt, &b.UpdatedAt, &finishedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if winnerID.Valid {
		b.WinnerID = &winnerID.Int64
	}
	if botDifficulty.Valid {
		b.BotDifficulty = botDifficulty.String
	}
	if finishedAt.Valid {
		t := finishedAt.Time
		b.FinishedAt = &t
	}

	json.Unmarshal(p1BoardJSON, &b.Player1Board)
	json.Unmarshal(p2BoardJSON, &b.Player2Board)
	json.Unmarshal(p1HandJSON, &b.Player1Hand)
	json.Unmarshal(p2HandJSON, &b.Player2Hand)
	json.Unmarshal(p1DeckJSON, &b.Player1Deck)
	json.Unmarshal(p2DeckJSON, &b.Player2Deck)
	json.Unmarshal(p1GraveJSON, &b.Player1Graveyard)
	json.Unmarshal(p2GraveJSON, &b.Player2Graveyard)
	json.Unmarshal(weatherJSON, &b.Weather)

	return &b, nil
}

func (r *BattleRepo) UpdateBattle(ctx context.Context, b *models.Battle) error {
	p1BoardJSON, _ := json.Marshal(b.Player1Board)
	p2BoardJSON, _ := json.Marshal(b.Player2Board)
	p1HandJSON, _ := json.Marshal(b.Player1Hand)
	p2HandJSON, _ := json.Marshal(b.Player2Hand)
	p1DeckJSON, _ := json.Marshal(b.Player1Deck)
	p2DeckJSON, _ := json.Marshal(b.Player2Deck)
	p1GraveJSON, _ := json.Marshal(b.Player1Graveyard)
	p2GraveJSON, _ := json.Marshal(b.Player2Graveyard)
	weatherJSON, _ := json.Marshal(b.Weather)

	b.UpdatedAt = time.Now()

	query := `
		UPDATE battles SET
			phase = $2, current_round = $3, current_turn = $4, winner_id = $5,
			player1_board = $6, player2_board = $7, player1_hand = $8, player2_hand = $9,
			player1_deck = $10, player2_deck = $11, player1_graveyard = $12, player2_graveyard = $13,
			player1_leader_used = $14, player2_leader_used = $15, player1_passed = $16, player2_passed = $17,
			weather = $18, updated_at = $19, finished_at = $20
		WHERE id = $1`

	var winnerID, finishedAt any
	if b.WinnerID != nil {
		winnerID = *b.WinnerID
	}
	if b.FinishedAt != nil {
		finishedAt = *b.FinishedAt
	}

	_, err := r.db.Pool.Exec(ctx, query,
		b.ID, b.Phase, b.CurrentRound, b.CurrentTurn, winnerID,
		p1BoardJSON, p2BoardJSON, p1HandJSON, p2HandJSON,
		p1DeckJSON, p2DeckJSON, p1GraveJSON, p2GraveJSON,
		b.Player1LeaderUsed, b.Player2LeaderUsed, b.Player1Passed, b.Player2Passed,
		weatherJSON, b.UpdatedAt, finishedAt,
	)
	return err
}

func (r *BattleRepo) GetActiveBattleForPlayer(ctx context.Context, playerID int64) (*models.Battle, error) {
	query := `
		SELECT id FROM battles 
		WHERE (player1_id = $1 OR player2_id = $1) 
		AND phase NOT IN ($2, $3)
		ORDER BY created_at DESC LIMIT 1`

	var battleID string
	err := r.db.Pool.QueryRow(ctx, query, playerID, models.PhaseFinished, models.PhaseWaiting).Scan(&battleID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return r.GetBattle(ctx, battleID)
}

func (r *BattleRepo) DeleteBattle(ctx context.Context, battleID string) error {
	query := `DELETE FROM battles WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, battleID)
	return err
}