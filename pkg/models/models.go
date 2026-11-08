package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ============================================================================
// ENUMS & TYPES (Mini-Gwent)
// ============================================================================

type Faction string

const (
	FactionNorthernRealms Faction = "northern_realms"
	FactionNilfgaard      Faction = "nilfgaard"
	FactionScoiatael      Faction = "scoiatael"
	FactionMonsters       Faction = "monsters"
	FactionNeutral        Faction = "neutral"
)

func (f Faction) String() string { return string(f) }
func (f Faction) Emoji() string {
	switch f {
	case FactionNorthernRealms:
		return "🦁"
	case FactionNilfgaard:
		return "🦅"
	case FactionScoiatael:
		return "🌿"
	case FactionMonsters:
		return "👹"
	default:
		return "⚪"
	}
}

type CardType string

const (
	CardTypeUnit     CardType = "unit"
	CardTypeSpecial  CardType = "special"
	CardTypeHero     CardType = "hero"
)

type RowType string

const (
	RowClose  RowType = "close"
	RowRanged RowType = "ranged"
	RowSiege  RowType = "siege"
)

func (r RowType) String() string { return string(r) }
func (r RowType) Emoji() string {
	switch r {
	case RowClose:
		return "⚔️"
	case RowRanged:
		return "🏹"
	case RowSiege:
		return "🏰"
	default:
		return "❓"
	}
}

type GamePhase string

const (
	PhaseWaiting  GamePhase = "waiting"
	PhaseMulligan GamePhase = "mulligan"
	PhaseRound1   GamePhase = "round_1"
	PhaseRound2   GamePhase = "round_2"
	PhaseRound3   GamePhase = "round_3"
	PhaseFinished GamePhase = "finished"
)

type BattleType string

const (
	BattleTypePvE    BattleType = "pve"
	BattleTypePvP    BattleType = "pvp"
	BattleTypeFriend BattleType = "friend"
)

type PlayerLevel int

const Level1 PlayerLevel = 1

// ============================================================================
// CORE MODELS (Simplified for Mini-Gwent)
// ============================================================================

type Card struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Faction      Faction  `json:"faction"`
	Type         CardType `json:"type"`
	Row          RowType  `json:"row,omitempty"`
	Power        int      `json:"power"`
	IsHero       bool     `json:"is_hero"`
	Abilities    []string `json:"abilities"`
	ArtURL       string   `json:"art_url,omitempty"`
	FlavorText   string   `json:"flavor_text,omitempty"`
	CardGroup    string   `json:"card_group"`
}

type CardInstance struct {
	CardID     string   `json:"card_id"`
	InstanceID string   `json:"instance_id"`
	Power      int      `json:"power"`
	Row        RowType  `json:"row"`
	OwnerID    int64    `json:"owner_id"`
	IsHero     bool     `json:"is_hero"`
	Abilities  []string `json:"abilities"`
}

type Player struct {
	ID                 int64       `json:"id"`
	TelegramID         int64       `json:"telegram_id"`
	Username           string      `json:"username"`
	FirstName          string      `json:"first_name"`
	LastName           string      `json:"last_name"`
	Level              PlayerLevel `json:"level"`
	Experience         int         `json:"experience"`
	Gold               int         `json:"gold"`
	Collection         []Card      `json:"collection"`
	Deck               []string    `json:"deck"`
	LeaderAbility      string      `json:"leader_ability"`
	CardsWon           int         `json:"cards_won"`
	OpponentsDefeated  []string    `json:"opponents_defeated"`
	GamesPlayed        int         `json:"games_played"`
	GamesWon           int         `json:"games_won"`
	GamesLost          int         `json:"games_lost"`
	HighestRank        int         `json:"highest_rank"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

func (p *Player) FactionEmoji() string {
	if len(p.Collection) > 0 {
		return p.Collection[0].Faction.Emoji()
	}
	return "🏴"
}

type Battle struct {
	ID              string       `json:"id"`
	Type            BattleType   `json:"type"`
	Phase           GamePhase    `json:"phase"`
	CurrentRound    int          `json:"current_round"`
	CurrentTurn     int64        `json:"current_turn"`
	Player1ID       int64        `json:"player1_id"`
	Player2ID       int64        `json:"player2_id,omitempty"`
	WinnerID        *int64       `json:"winner_id,omitempty"`
	Player1Board    BoardState   `json:"player1_board"`
	Player2Board    BoardState   `json:"player2_board"`
	Player1Hand     []CardInstance `json:"player1_hand"`
	Player2Hand     []CardInstance `json:"player2_hand"`
	Player1Deck     []CardInstance `json:"player1_deck"`
	Player2Deck     []CardInstance `json:"player2_deck"`
	Player1Graveyard []CardInstance `json:"player1_graveyard"`
	Player2Graveyard []CardInstance `json:"player2_graveyard"`
	Player1LeaderUsed bool       `json:"player1_leader_used"`
	Player2LeaderUsed bool       `json:"player2_leader_used"`
	Player1Passed     bool       `json:"player1_passed"`
	Player2Passed     bool       `json:"player2_passed"`
	Weather           WeatherState `json:"weather"`
	BotDifficulty     string       `json:"bot_difficulty,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}

func (b *Battle) PhaseDisplay() string { return string(b.Phase) }

type BoardState struct {
	Close  []CardInstance `json:"close"`
	Ranged []CardInstance `json:"ranged"`
	Siege  []CardInstance `json:"siege"`
	TotalPower int        `json:"total_power"`
}

type WeatherState struct {
	Active   string `json:"active"`
	Duration int    `json:"duration"`
}

func (w *WeatherState) Display() string {
	switch w.Active {
	case "frost":
		return "❄️ Frost"
	case "fog":
		return "🌫️ Fog"
	case "rain":
		return "🌧️ Rain"
	default:
		return "☀️ Clear"
	}
}

// JSONB helpers
type CardDB []Card

func (c CardDB) Value() (driver.Value, error) { return json.Marshal(c) }
func (c *CardDB) Scan(v interface{}) error {
	b, ok := v.([]byte)
	if !ok {
		return errors.New("type assertion failed")
	}
	return json.Unmarshal(b, c)
}

type CardInstanceDB []CardInstance

func (c CardInstanceDB) Value() (driver.Value, error) { return json.Marshal(c) }
func (c *CardInstanceDB) Scan(v interface{}) error {
	b, ok := v.([]byte)
	if !ok {
		return errors.New("type assertion failed")
	}
	return json.Unmarshal(b, c)
}