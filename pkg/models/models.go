package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ============================================================================
// ENUMS & TYPES (Mini-Gwent - Witcher 3 Style)
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

// Nation is an alias for Faction (for compatibility with existing code)
type Nation = Faction

const (
	NationNorthernRealms = FactionNorthernRealms
	NationNilfgaard      = FactionNilfgaard
	NationScoiatael      = FactionScoiatael
	NationMonsters       = FactionMonsters
	NationNeutral        = FactionNeutral
	NationSyndicate      = FactionNeutral // Syndicate not in Mini-Gwent, map to Neutral
)

type CardType string

const (
	CardTypeUnit    CardType = "unit"
	CardTypeSpecial CardType = "special"
	CardTypeHero    CardType = "hero"
	CardTypeLeader  CardType = "leader" // For GWENT: Masters compatibility
)

type RowType string

const (
	RowMelee  RowType = "melee"  // Changed from "close" to match Gwent terminology
	RowRanged RowType = "ranged"
	RowSiege  RowType = "siege"
)

func (r RowType) String() string { return string(r) }
func (r RowType) Emoji() string {
	switch r {
	case RowMelee:
		return "⚔️"
	case RowRanged:
		return "🏹"
	case RowSiege:
		return "🏰"
	default:
		return "❓"
	}
}

// Close is an alias for Melee (for backward compatibility)
const RowClose = RowMelee

type Rarity string

const (
	RarityCommon    Rarity = "common"
	RarityRare      Rarity = "rare"
	RarityEpic      Rarity = "epic"
	RarityLegendary Rarity = "legendary"
)

func (r Rarity) Emoji() string {
	switch r {
	case RarityCommon:
		return "⚪"
	case RarityRare:
		return "🔵"
	case RarityEpic:
		return "🟣"
	case RarityLegendary:
		return "🟡"
	default:
		return "⚪"
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
// CORE MODELS (Mini-Gwent)
// ============================================================================

type Card struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Faction     Faction  `json:"faction"`
	Nation      Nation   `json:"nation"` // Alias for Faction
	Type        CardType `json:"type"`
	Row         RowType  `json:"row,omitempty"`
	Power       int      `json:"power"`
	Armor       int      `json:"armor"`       // For GWENT: Masters compatibility
	Provision   int      `json:"provision"`   // For GWENT: Masters compatibility
	IsHero      bool     `json:"is_hero"`
	Abilities   []string `json:"abilities"`   // Keyword names for Mini-Gwent
	ArtURL      string   `json:"art_url,omitempty"`
	FlavorText  string   `json:"flavor_text,omitempty"`
	CardGroup   string   `json:"card_group"`  // close, ranged, siege, special, hero
	Rarity      Rarity   `json:"rarity"`      // For GWENT: Masters compatibility
	CostGold    int      `json:"cost_gold"`   // For shop
	CostScrap   int      `json:"cost_scrap"`  // For shop
}

type CardInstance struct {
	CardID     string     `json:"card_id"`
	InstanceID string     `json:"instance_id"`
	Power      int        `json:"power"`
	BasePower  int        `json:"base_power"`
	Armor      int        `json:"armor"`
	Boosted    int        `json:"boosted"`
	Damaged    int        `json:"damaged"`
	Row        RowType    `json:"row"`
	OwnerID    int64      `json:"owner_id"`
	IsHero     bool       `json:"is_hero"`
	Abilities  []string   `json:"abilities"`
	Statuses   []Status   `json:"statuses"`
	Card       *Card      `json:"-"` // Reference to base card (not serialized)
}

type Status struct {
	Name     string `json:"name"`
	Duration int    `json:"duration"`
	Value    int    `json:"value"`
}

func (s Status) Emoji() string {
	switch s.Name {
	case "frost":
		return "❄️"
	case "fog":
		return "🌫️"
	case "rain":
		return "🌧️"
	case "boost":
		return "🟢"
	case "damage":
		return "🔴"
	case "armor":
		return "🛡️"
	case "shield":
		return "🛡️"
	case "poison":
		return "☠️"
	case "bleeding":
		return "🩸"
	case "vitality":
		return "💚"
	default:
		return "❓"
	}
}

type Player struct {
	ID                 int64        `json:"id"`
	TelegramID         int64        `json:"telegram_id"`
	Username           string       `json:"username"`
	FirstName          string       `json:"first_name"`
	LastName           string       `json:"last_name"`
	Level              PlayerLevel  `json:"level"`
	Experience         int          `json:"experience"`
	Gold               int          `json:"gold"`
	Scrap              int          `json:"scrap"`
	MeteoritePowder    int          `json:"meteorite_powder"`
	Collection         []Card       `json:"collection"`
	Deck               []string     `json:"deck"`
	LeaderID           string       `json:"leader_id"`
	LeaderAbility      string       `json:"leader_ability"`
	AxeDurability      int          `json:"axe_durability"`
	MineUnlocked       bool         `json:"mine_unlocked"`
	MineExtractsToday  int          `json:"mine_extracts_today"`
	LastMineExtractAt  *time.Time   `json:"last_mine_extract_at"`
	DailyCoinClaimedAt *time.Time   `json:"daily_coin_claimed_at"`
	AxeClaimedAt       *time.Time   `json:"axe_claimed_at"`
	CardsWon           int          `json:"cards_won"`
	OpponentsDefeated  []string     `json:"opponents_defeated"`
	GamesPlayed        int          `json:"games_played"`
	GamesWon           int          `json:"games_won"`
	GamesLost          int          `json:"games_lost"`
	HighestRank        int          `json:"highest_rank"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
}

func (p *Player) FactionEmoji() string {
	if len(p.Collection) > 0 {
		return p.Collection[0].Faction.Emoji()
	}
	return "🏴"
}

func (p *Player) NationEmoji() string {
	return p.FactionEmoji()
}

type Battle struct {
	ID                string        `json:"id"`
	Type              BattleType    `json:"type"`
	Phase             GamePhase     `json:"phase"`
	CurrentRound      int           `json:"current_round"`
	CurrentTurn       int64         `json:"current_turn"`
	Player1ID         int64         `json:"player1_id"`
	Player2ID         int64         `json:"player2_id,omitempty"`
	WinnerID          *int64        `json:"winner_id,omitempty"`
	Player1Board      BoardState    `json:"player1_board"`
	Player2Board      BoardState    `json:"player2_board"`
	Player1Hand       []CardInstance `json:"player1_hand"`
	Player2Hand       []CardInstance `json:"player2_hand"`
	Player1Deck       []CardInstance `json:"player1_deck"`
	Player2Deck       []CardInstance `json:"player2_deck"`
	Player1Graveyard  []CardInstance `json:"player1_graveyard"`
	Player2Graveyard  []CardInstance `json:"player2_graveyard"`
	Player1LeaderUsed bool          `json:"player1_leader_used"`
	Player2LeaderUsed bool          `json:"player2_leader_used"`
	Player1Passed     bool          `json:"player1_passed"`
	Player2Passed     bool          `json:"player2_passed"`
	Weather           WeatherState  `json:"weather"`
	BotDifficulty     string        `json:"bot_difficulty,omitempty"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
	FinishedAt        *time.Time    `json:"finished_at,omitempty"`
}

func (b *Battle) PhaseDisplay() string { return string(b.Phase) }

type BoardState struct {
	Melee     []CardInstance `json:"melee"`
	Ranged    []CardInstance `json:"ranged"`
	Siege     []CardInstance `json:"siege"`
	TotalPower int           `json:"total_power"`
}

// Aliases for backward compatibility
func (b *BoardState) Close() []CardInstance { return b.Melee }
func (b *BoardState) SetClose(cards []CardInstance) { b.Melee = cards }

type WeatherState struct {
	Active   string `json:"active"`   // clear, frost, fog, rain
	Duration int    `json:"duration"` // rounds remaining
	Row      string `json:"row"`      // melee, ranged, siege, or "" for all
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

func (w *WeatherState) Emoji() string {
	switch w.Active {
	case "frost":
		return "❄️"
	case "fog":
		return "🌫️"
	case "rain":
		return "🌧️"
	default:
		return "☀️"
	}
}

// Progression models
type ChestReward struct {
	CardID  string `json:"card_id"`
	MinQty  int    `json:"min_qty"`
	MaxQty  int    `json:"max_qty"`
	Weight  int    `json:"weight"`
}

type Chest struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	LevelRequired int            `json:"level_required"`
	Cards         []ChestReward  `json:"cards"`
	Gold          int            `json:"gold"`
	Cooldown      time.Duration  `json:"cooldown"`
}

// JSONB helpers for database
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

type StringArray []string

func (s StringArray) Value() (driver.Value, error) { return json.Marshal(s) }
func (s *StringArray) Scan(v interface{}) error {
	b, ok := v.([]byte)
	if !ok {
		return errors.New("type assertion failed")
	}
	return json.Unmarshal(b, s)
}

type CardRowDB []CardInstance

func (c CardRowDB) Value() (driver.Value, error) { return json.Marshal(c) }
func (c *CardRowDB) Scan(v interface{}) error {
	b, ok := v.([]byte)
	if !ok {
		return errors.New("type assertion failed")
	}
	return json.Unmarshal(b, c)
}