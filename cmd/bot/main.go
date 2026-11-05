package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tele "gopkg.in/telebot.v3"

	"gwent-bot/internal/config"
	"gwent-bot/internal/database"
	"gwent-bot/internal/game/battle"
	"gwent-bot/internal/game/cards"
	"gwent-bot/internal/game/progression"
	"gwent-bot/internal/telegram/handlers"
	"gwent-bot/internal/telegram/keyboards"
	"gwent-bot/internal/telegram/render"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.RunMigrations(ctx); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	playerRepo := database.NewPlayerRepo(db)
	cardRepo := database.NewCardRepo(db)
	battleRepo := database.NewBattleRepo(db)

	cardService := cards.NewService(cardRepo)
	battleService := battle.NewService(battleRepo, cardService, playerRepo)
	progressionService := progression.NewService(playerRepo, cardService, cfg)

	msgRenderer := render.New(cfg)
	keyboardBuilder := keyboards.New(cfg)

	botToken := cfg.BotToken
	if botToken == "" {
		log.Fatal("BOT_TOKEN is required")
	}

	pref := tele.Settings{
		Token:  botToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	if cfg.BotMode == "webhook" && cfg.WebhookURL != "" {
		pref.Poller = &tele.Webhook{
			Listen: ":" + cfg.WebhookPort,
			Endpoint: &tele.WebhookEndpoint{
				PublicURL: cfg.WebhookURL,
			},
		}
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	h := handlers.New(
		bot,
		playerRepo,
		cardRepo,
		battleRepo,
		cardService,
		battleService,
		progressionService,
		msgRenderer,
		keyboardBuilder,
		cfg,
	)

	h.RegisterRoutes()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		bot.Stop()
	}()

	log.Println("Gwent Bot started!")
	if cfg.BotMode == "webhook" {
		log.Printf("Webhook mode on port %s", cfg.WebhookPort)
	} else {
		log.Println("Polling mode")
	}

	bot.Start()
}
