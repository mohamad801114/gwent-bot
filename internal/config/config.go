package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken       string
	BotMode        string
	WebhookURL     string
	WebhookPort    string
	DatabaseURL    string
	DBMaxConns     int
	DBMinConns     int
	RedisURL       string
	RedisEnabled   bool

	// Game special words
	StartWord         string
	DailyCoinWord     string
	AxeWord           string
	DailyCoinCooldown time.Duration
	AxeCooldown       time.Duration

	// Game Balance
	StartingGold      int
	StartingLevel     int
	ChestCardsLevel1  int
	ChestCardsLevel2  int
	MineGoldPerExtract int
	MineMaxExtracts   int
	AxeDurability     int

	// Rarity Weights
	RarityCommonWeight    int
	RarityRareWeight      int
	RarityEpicWeight      int
	RarityLegendaryWeight int

	Nations []string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		BotToken:       getEnv("BOT_TOKEN", ""),
		BotMode:        getEnv("BOT_MODE", "polling"),
		WebhookURL:     getEnv("WEBHOOK_URL", ""),
		WebhookPort:    getEnv("WEBHOOK_PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		DBMaxConns:     getEnvInt("DB_MAX_CONNS", 25),
		DBMinConns:     getEnvInt("DB_MIN_CONNS", 5),
		RedisURL:       getEnv("REDIS_URL", ""),
		RedisEnabled:   getEnvBool("REDIS_ENABLED", false),

		StartWord:         getEnv("START_WORD", "gwentstart"),
		DailyCoinWord:     getEnv("DAILY_COIN_WORD", "goldmine"),
		AxeWord:           getEnv("AXE_WORD", "sharpaxe"),
		DailyCoinCooldown: getEnvDuration("DAILY_COIN_COOLDOWN_MINUTES", 5) * time.Minute,
		AxeCooldown:       getEnvDuration("AXE_COOLDOWN_HOURS", 1) * time.Hour,

		StartingGold:        getEnvInt("STARTING_GOLD", 100),
		StartingLevel:       getEnvInt("STARTING_LEVEL", 1),
		ChestCardsLevel1:    getEnvInt("CHEST_CARDS_LEVEL_1", 5),
		ChestCardsLevel2:    getEnvInt("CHEST_CARDS_LEVEL_2", 3),
		MineGoldPerExtract:  getEnvInt("MINE_GOLD_PER_EXTRACT", 50),
		MineMaxExtracts:     getEnvInt("MINE_MAX_EXTRACTS", 10),
		AxeDurability:       getEnvInt("AXE_DURABILITY", 10),

		RarityCommonWeight:    getEnvInt("RARITY_COMMON_WEIGHT", 50),
		RarityRareWeight:      getEnvInt("RARITY_RARE_WEIGHT", 30),
		RarityEpicWeight:      getEnvInt("RARITY_EPIC_WEIGHT", 15),
		RarityLegendaryWeight: getEnvInt("RARITY_LEGENDARY_WEIGHT", 5),

		Nations: strings.Split(getEnv("NATIONS", "Northern Realms,Nilfgaard,Scoia'tael,Monsters,Syndicate"), ","),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue int64) time.Duration {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return time.Duration(intVal)
		}
	}
	return time.Duration(defaultValue)
}
