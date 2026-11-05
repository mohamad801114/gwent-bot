package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ============================================================================
// ENUMS & TYPES
// ============================================================================

type Nation string

const (
	NationNorthernRealms Nation = "Northern Realms"
	NationNilfgaard      Nation = "Nilfgaard"
	NationScoiatael      Nation = "Scoia'tael"
	NationMonsters       Nation = "Monsters"
	NationSyndicate      Nation = "Syndicate"
	NationNeutral        Nation = "Neutral"
)

func (n Nation) String() string {
	return string(n)
}

func (n Nation) Emoji() string {
	switch n {
	case NationNorthernRealms:
		return "???"
	case NationNilfgaard:
		return "??"
	case NationScoiatael:
		return "??"
	case NationMonsters:
		return "??"
	case NationSyndicate:
		return "??"
	default:
		return "?"
	}
}

func (n Nation) Color() string {
	switch n {
	case NationNorthernRealms:
		return "#d4a537"
	case NationNilfgaard:
		return "#e06060"
	case NationScoiatael:
		return "#60e060"
	case NationMonsters:
		return "#c060e0"
	case NationSyndicate:
		return "#f0c060"
	default:
		return "#888888"
	}
}

type Rarity string

const (
	RarityCommon    Rarity = "common"
	RarityRare      Rarity = "rare"
	RarityEpic      Rarity = "epic"
	RarityLegendary Rarity = "legendary"
)

func (r Rarity) String() string {
	return string(r)
}

func (r Rarity) Emoji() string {
	switch r {
	case RarityCommon:
		return "?"
	case RarityRare:
		return "??"
	case RarityEpic:
		return "??"
	case RarityLegendary:
		return "??"
	default:
		return "?"
	}
}

func (r Rarity) Color() string {
	switch r {
	case RarityCommon:
		return "#aaaaaa"
	case RarityRare:
		return "#4a90d9"
	case RarityEpic:
		return "#a357d9"
	case RarityLegendary:
		return "#ffd700"
	default:
		return "#888888"
	}
}

type CardType string

const (
	CardTypeUnit    CardType = "unit"
	CardTypeSpecial CardType = "special"
	CardTypeLeader  CardType = "leader"
	CardTypeArtifact CardType = "artifact"
)

type RowType string

const (
	RowMelee   RowType = "melee"
	RowRanged  RowType = "ranged"
	RowSiege   RowType = "siege"
)

func (r RowType) String() string {
	return string(r)
}

func (r RowType) Emoji() string {
	switch r {
	case RowMelee:
		return "???"
	case RowRanged:
		return "??"
	case RowSiege:
		return "???"
	default:
		return "?"
	}
}

type GamePhase string

const (
	PhaseWaiting    GamePhase = "waiting"
	PhaseMulligan   GamePhase = "mulligan"
	PhaseRound1     GamePhase = "round_1"
	PhaseRound2     GamePhase = "round_2"
	PhaseRound3     GamePhase = "round_3"
	PhaseFinished   GamePhase = "finished"
)

type BattleType string

const (
	BattleTypePvE    BattleType = "pve"
	BattleTypePvP    BattleType = "pvp"
	BattleTypeFriend BattleType = "friend"
)

type PlayerLevel int

const (
	Level1 PlayerLevel = 1
	Level2 PlayerLevel = 2
	Level3 PlayerLevel = 3
)

type CardDB []Card

func (c CardDB) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CardDB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

type CardRowDB []CardInstance

func (c CardRowDB) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CardRowDB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

// ============================================================================
// CORE MODELS
// ============================================================================

type Card struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Nation      Nation     `json:"nation"`
	Rarity      Rarity     `json:"rarity"`
	Type        CardType   `json:"type"`
	Row         RowType    `json:"row,omitempty"`
	Provision   int        `json:"provision"`
	Power       int        `json:"power"`
	Armor       int        `json:"armor,omitempty"`
	Abilities   []Ability  `json:"abilities,omitempty"`
	ArtURL      string     `json:"art_url,omitempty"`
	FlavorText  string     `json:"flavor_text,omitempty"`
	CostGold    int        `json:"cost_gold,omitempty"`
	CostScrap   int        `json:"cost_scrap,omitempty"`
}

type Ability struct {
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Value       int            `json:"value,omitempty"`
	Conditions  map[string]any `json:"conditions,omitempty"`
}

type CardInstance struct {
	CardID     string    `json:"card_id"`
	InstanceID string    `json:"instance_id"`
	Power      int       `json:"power"`
	Armor      int       `json:"armor"`
	Boosted    int       `json:"boosted,omitempty"`
	Damaged    int       `json:"damaged,omitempty"`
	Statuses   []Status  `json:"statuses,omitempty"`
	Row        RowType   `json:"row"`
	OwnerID    int64     `json:"owner_id"`
}

type Status struct {
	Type      string `json:"type"`
	Value     int    `json:"value,omitempty"`
	Duration  int    `json:"duration,omitempty"`
	SourceID  string `json:"source_id,omitempty"`
}

func (s *Status) Emoji() string {
	switch s.Type {
	case "boost":
		return "🟢"
	case "damage":
		return "🔴"
	case "shield":
		return "🛡️"
	case "poison":
		return "☠️"
	case "vitality":
		return "💚"
	case "bleeding":
		return "🩸"
	case "lock":
		return "🔒"
	case "veil":
		return "🌫️"
	default:
		return "❓"
	}
}

type Player struct {
	ID              int64      `json:"id"`
	TelegramID      int64      `json:"telegram_id"`
	Username        string     `json:"username"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	Level           PlayerLevel `json:"level"`
	Experience      int        `json:"experience"`
	Gold            int        `json:"gold"`
	Scrap           int        `json:"scrap"`
	MeteoritePowder int        `json:"meteorite_powder"`

	// Collection
	Collection      CardDB     `json:"collection"`
	Deck            []string   `json:"deck"`
	LeaderID        string     `json:"leader_id"`

	// Progression
	DailyCoinClaimedAt *time.Time `json:"daily_coin_claimed_at,omitempty"`
	AxeClaimedAt       *time.Time `json:"axe_claimed_at,omitempty"`
	AxeDurability      int        `json:"axe_durability"`
	MineUnlocked       bool       `json:"mine_unlocked"`
	MineExtractsToday  int        `json:"mine_extracts_today"`
	LastMineExtractAt  *time.Time `json:"last_mine_extract_at,omitempty"`

	// Stats
	GamesPlayed   int `json:"games_played"`
	GamesWon      int `json:"games_won"`
	GamesLost     int `json:"games_lost"`
	HighestRank   int `json:"highest_rank"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *Player) NationEmoji() string {
	// Return the emoji of the first nation in collection, or default
	if len(p.Collection) > 0 {
		return p.Collection[0].Nation.Emoji()
	}
	return "🏴"
}

type Battle struct {
	ID           string       `json:"id"`
	Type         BattleType   `json:"type"`
	Phase        GamePhase    `json:"phase"`
	CurrentRound int          `json:"current_round"`
	CurrentTurn  int64        `json:"current_turn"` // Player ID whose turn it is
	Player1ID    int64        `json:"player1_id"`
	Player2ID    int64        `json:"player2_id,omitempty"` // nil for PvE
	WinnerID     *int64       `json:"winner_id,omitempty"`

	// Board state
	Player1Board BoardState `json:"player1_board"`
	Player2Board BoardState `json:"player2_board"`

	// Hands
	Player1Hand []CardInstance `json:"player1_hand"`
	Player2Hand []CardInstance `json:"player2_hand"`

	// Decks (remaining cards)
	Player1Deck []CardInstance `json:"player1_deck"`
	Player2Deck []CardInstance `json:"player2_deck"`

	// Graveyards
	Player1Graveyard []CardInstance `json:"player1_graveyard"`
	Player2Graveyard []CardInstance `json:"player2_graveyard"`

	// Leader abilities
	Player1LeaderUsed bool `json:"player1_leader_used"`
	Player2LeaderUsed bool `json:"player2_leader_used"`

	// Pass states
	Player1Passed bool `json:"player1_passed"`
	Player2Passed bool `json:"player2_passed"`

	// Weather
	Weather WeatherState `json:"weather"`

	// For PvE
	BotDifficulty string `json:"bot_difficulty,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

func (b *Battle) PhaseDisplay() string {
	return string(b.Phase)
}

type BoardState struct {
	Melee   CardRowDB `json:"melee"`
	Ranged  CardRowDB `json:"ranged"`
	Siege   CardRowDB `json:"siege"`
	TotalPower int    `json:"total_power"`
}

type WeatherState struct {
	Active   string `json:"active"`   // frost, fog, rain, clear
	Duration int    `json:"duration"` // rounds remaining
	Row      RowType `json:"row"`     // affected row for frost/fog/rain
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

type Chest struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	LevelRequired PlayerLevel `json:"level_required"`
	Cards       []ChestReward `json:"cards"`
	Gold        int        `json:"gold"`
	Cooldown    time.Duration `json:"cooldown"`
}

type ChestReward struct {
	CardID  string `json:"card_id"`
	MinQty  int    `json:"min_qty"`
	MaxQty  int    `json:"max_qty"`
	Weight  int    `json:"weight"`
}

type ShopItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CardID      string `json:"card_id"`
	CostGold    int    `json:"cost_gold"`
	CostScrap   int    `json:"cost_scrap"`
	Stock       int    `json:"stock"` // -1 for unlimited
	LevelRequired PlayerLevel `json:"level_required"`
}

type MineState struct {
	PlayerID    int64     `json:"player_id"`
	GoldExtracted int     `json:"gold_extracted"`
	ExtractsLeft  int     `json:"extracts_left"`
	AxeDurability int     `json:"axe_durability"`
	LastExtractAt *time.Time `json:"last_extract_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GameSettings struct {
	StartWord          string
	DailyCoinWord      string
	AxeWord            string
	DailyCoinCooldown  time.Duration
	AxeCooldown        time.Duration
	StartingGold       int
	StartingLevel      PlayerLevel
	ChestCardsLevel1   int
	ChestCardsLevel2   int
	MineGoldPerExtract int
	MineMaxExtracts    int
	AxeDurability      int
	RarityWeights      map[Rarity]int
	Nations            []Nation
}
