package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"

	"gwent-bot/internal/config"
	"gwent-bot/internal/database"
	"gwent-bot/internal/game/battle"
	"gwent-bot/internal/game/cards"
	"gwent-bot/internal/game/progression"
	"gwent-bot/internal/telegram/keyboards"
	"gwent-bot/internal/telegram/render"
	"gwent-bot/pkg/models"
)

type Handlers struct {
	bot                *tele.Bot
	playerRepo         *database.PlayerRepo
	cardRepo           *database.CardRepo
	battleRepo         *database.BattleRepo
	cardService        *cards.Service
	battleService      *battle.Service
	progressionService *progression.Service
	renderer           *render.Renderer
	keyboardBuilder    *keyboards.KeyboardBuilder
	cfg                *config.Config
}

func New(
	bot *tele.Bot,
	playerRepo *database.PlayerRepo,
	cardRepo *database.CardRepo,
	battleRepo *database.BattleRepo,
	cardService *cards.Service,
	battleService *battle.Service,
	progressionService *progression.Service,
	renderer *render.Renderer,
	keyboardBuilder *keyboards.KeyboardBuilder,
	cfg *config.Config,
) *Handlers {
	return &Handlers{
		bot:                bot,
		playerRepo:         playerRepo,
		cardRepo:           cardRepo,
		battleRepo:         battleRepo,
		cardService:        cardService,
		battleService:      battleService,
		progressionService: progressionService,
		renderer:           renderer,
		keyboardBuilder:    keyboardBuilder,
		cfg:                cfg,
	}
}

func (h *Handlers) RegisterRoutes() {
	h.bot.Handle("/start", h.OnStart)
	h.bot.Handle("/help", h.OnHelp)
	h.bot.Handle("/profile", h.OnProfile)
	h.bot.Handle("/deck", h.OnDeck)
	h.bot.Handle("/collection", h.OnCollection)
	h.bot.Handle("/battle", h.OnBattle)
	h.bot.Handle("/shop", h.OnShop)
	h.bot.Handle("/mine", h.OnMine)
	h.bot.Handle("/chest", h.OnChest)
	h.bot.Handle("/leader", h.OnLeader)

	h.bot.Handle(&tele.Btn{Unique: "main_menu"}, h.OnMainMenu)
	h.bot.Handle(&tele.Btn{Unique: "battle_menu"}, h.OnBattleMenu)
	h.bot.Handle(&tele.Btn{Unique: "show_deck"}, h.OnShowDeck)
	h.bot.Handle(&tele.Btn{Unique: "show_collection"}, h.OnShowCollection)
	h.bot.Handle(&tele.Btn{Unique: "show_profile"}, h.OnShowProfile)
	h.bot.Handle(&tele.Btn{Unique: "show_help"}, h.OnShowHelp)
	h.bot.Handle(&tele.Btn{Unique: "shop_menu"}, h.OnShopMenu)
	h.bot.Handle(&tele.Btn{Unique: "mine_menu"}, h.OnMineMenu)
	h.bot.Handle(&tele.Btn{Unique: "chest_menu"}, h.OnChestMenu)
	h.bot.Handle(&tele.Btn{Unique: "leader_menu"}, h.OnLeaderMenu)

	h.bot.Handle(&tele.Btn{Unique: "battle_pve_easy"}, h.OnBattlePvEEasy)
	h.bot.Handle(&tele.Btn{Unique: "battle_pve_medium"}, h.OnBattlePvEMedium)
	h.bot.Handle(&tele.Btn{Unique: "battle_pve_hard"}, h.OnBattlePvEHard)
	h.bot.Handle(&tele.Btn{Unique: "battle_pvp"}, h.OnBattlePvP)
	h.bot.Handle(&tele.Btn{Unique: "battle_friend"}, h.OnBattleFriend)

	h.bot.Handle(tele.OnCallback, h.OnCallback)

	h.bot.Handle(tele.OnText, h.OnText)
}

func (h *Handlers) OnStart(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil {
		log.Printf("Error getting player: %v", err)
		return c.Send(h.renderer.ErrorMessage("Failed to load profile"))
	}

	if player == nil {
		player = &models.Player{
			TelegramID:       user.ID,
			Username:         user.Username,
			FirstName:        user.FirstName,
			LastName:         user.LastName,
			Level:            models.Level1,
			Experience:       0,
			Gold:             h.cfg.StartingGold,
			Scrap:            0,
			MeteoritePowder:  0,
			Collection:       []models.Card{},
			Deck:             []string{},
			LeaderID:         "",
			AxeDurability:    0,
			MineUnlocked:     false,
			MineExtractsToday: 0,
			GamesPlayed:      0,
			GamesWon:         0,
			GamesLost:        0,
			HighestRank:      0,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		err = h.playerRepo.CreatePlayer(ctx, player)
		if err != nil {
			log.Printf("Error creating player: %v", err)
			return c.Send(h.renderer.ErrorMessage("Failed to create profile"))
		}

		starterCards, err := h.progressionService.GetStarterRewards(ctx, player.ID, models.NationNorthernRealms)
		if err != nil {
			log.Printf("Error giving starter cards: %v", err)
		}

		deckIDs := make([]string, len(starterCards))
		for i, card := range starterCards {
			deckIDs[i] = card.ID
		}
		player.Deck = deckIDs
		h.playerRepo.UpdatePlayer(ctx, player)

		msg := h.renderer.WelcomeMessage(player)
		return c.Send(msg, h.keyboardBuilder.MainMenu(), tele.ModeHTML)
	}

	msg := h.renderer.WelcomeMessage(player)
	return c.Send(msg, h.keyboardBuilder.MainMenu(), tele.ModeHTML)
}

func (h *Handlers) OnHelp(c tele.Context) error {
	msg := h.renderer.HelpMessage()
	return c.Send(msg, tele.ModeHTML)
}

func (h *Handlers) OnProfile(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	msg := h.renderer.ProfileMessage(player)
	return c.Send(msg, h.keyboardBuilder.MainMenu(), tele.ModeHTML)
}

func (h *Handlers) OnDeck(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	var deckCards []models.Card
	for _, cardID := range player.Deck {
		card, err := h.cardService.GetCard(ctx, cardID)
		if err == nil && card != nil {
			deckCards = append(deckCards, *card)
		}
	}

	msg := h.renderer.DeckMessage(player, deckCards)
	return c.Send(msg, h.keyboardBuilder.DeckManagement(player.Deck, deckCards, 25), tele.ModeHTML)
}

func (h *Handlers) OnCollection(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	allCards, err := h.cardService.GetAllCards(ctx)
	if err != nil {
		return c.Send(h.renderer.ErrorMessage("Failed to load collection"), tele.ModeHTML)
	}

	ownedMap := make(map[string]bool)
	for _, card := range player.Collection {
		ownedMap[card.ID] = true
	}

	var ownedCards []models.Card
	for _, card := range allCards {
		if ownedMap[card.ID] {
			ownedCards = append(ownedCards, *card)
		}
	}

	msg := h.renderer.CollectionMessage(player, ownedCards, 1, 1)
	return c.Send(msg, h.keyboardBuilder.Pagination(1, 1, "collection"), tele.ModeHTML)
}

func (h *Handlers) OnBattle(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	activeBattle, err := h.battleService.GetActiveBattle(ctx, player.ID)
	if err == nil && activeBattle != nil {
		return h.showBattleState(c, activeBattle, player)
	}

	msg := "âš”ï¸ <b>Battle Menu</b>\n\nChoose your battle type:"
	return c.Send(msg, h.keyboardBuilder.BattleMenu(), tele.ModeHTML)
}

func (h *Handlers) OnShop(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	msg := h.renderer.ShopMessage(player)
	return c.Send(msg, h.keyboardBuilder.ShopMenu(), tele.ModeHTML)
}

func (h *Handlers) OnMine(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	msg := h.renderer.MineMessage(player, h.cfg)
	return c.Send(msg, h.keyboardBuilder.MineMenu(player.MineUnlocked, h.cfg.MineMaxExtracts-player.MineExtractsToday, player.AxeDurability), tele.ModeHTML)
}

func (h *Handlers) OnChest(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	chests := h.getAvailableChests(player.Level)
	msg := h.renderer.ChestMessage(chests, player.Level)
	return c.Send(msg, h.keyboardBuilder.ChestMenu(chests, player.Level), tele.ModeHTML)
}

func (h *Handlers) OnLeader(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	leadersPtr, err := h.cardRepo.GetCardsByType(ctx, models.CardTypeLeader)
	if err != nil {
		return c.Send(h.renderer.ErrorMessage("Failed to load leaders"), tele.ModeHTML)
	}

	// Convert []*models.Card to []models.Card
	leaders := make([]models.Card, len(leadersPtr))
	for i, l := range leadersPtr {
		if l != nil {
			leaders[i] = *l
		}
	}

	msg := h.renderer.LeaderMessage(leaders, player.LeaderID)
	return c.Send(msg, h.keyboardBuilder.LeaderSelection(leaders, player.LeaderID), tele.ModeHTML)
}

func (h *Handlers) OnMainMenu(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	msg := h.renderer.WelcomeMessage(player)
	return c.Edit(msg, h.keyboardBuilder.MainMenu(), tele.ModeHTML)
}

func (h *Handlers) OnBattleMenu(c tele.Context) error {
	msg := "âš”ï¸ <b>Battle Menu</b>\n\nChoose your battle type:"
	return c.Edit(msg, h.keyboardBuilder.BattleMenu(), tele.ModeHTML)
}

func (h *Handlers) OnShowDeck(c tele.Context) error {
	return h.OnDeck(c)
}

func (h *Handlers) OnShowCollection(c tele.Context) error {
	return h.OnCollection(c)
}

func (h *Handlers) OnShowProfile(c tele.Context) error {
	return h.OnProfile(c)
}

func (h *Handlers) OnShowHelp(c tele.Context) error {
	msg := h.renderer.HelpMessage()
	return c.Edit(msg, h.keyboardBuilder.MainMenu(), tele.ModeHTML)
}

func (h *Handlers) OnShopMenu(c tele.Context) error {
	return h.OnShop(c)
}

func (h *Handlers) OnMineMenu(c tele.Context) error {
	return h.OnMine(c)
}

func (h *Handlers) OnChestMenu(c tele.Context) error {
	return h.OnChest(c)
}

func (h *Handlers) OnLeaderMenu(c tele.Context) error {
	return h.OnLeader(c)
}

func (h *Handlers) OnBattlePvEEasy(c tele.Context) error {
	return h.createPvEBattle(c, "easy")
}

func (h *Handlers) OnBattlePvEMedium(c tele.Context) error {
	return h.createPvEBattle(c, "medium")
}

func (h *Handlers) OnBattlePvEHard(c tele.Context) error {
	return h.createPvEBattle(c, "hard")
}

func (h *Handlers) createPvEBattle(c tele.Context, difficulty string) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	battle, err := h.battleService.CreatePvEBattle(ctx, player.ID, difficulty, player.Deck, player.LeaderID)
	if err != nil {
		log.Printf("Error creating PvE battle: %v", err)
		return c.Send(h.renderer.ErrorMessage("Failed to create battle"), tele.ModeHTML)
	}

	return h.showBattleState(c, battle, player)
}

func (h *Handlers) OnBattlePvP(c tele.Context) error {
	return c.Send(h.renderer.InfoMessage("PvP matchmaking coming soon!"), tele.ModeHTML)
}

func (h *Handlers) OnBattleFriend(c tele.Context) error {
	return c.Send(h.renderer.InfoMessage("Friend battles coming soon!"), tele.ModeHTML)
}

func (h *Handlers) showBattleState(c tele.Context, b *models.Battle, player *models.Player) error {
	opponentName := "Bot"
	if b.Type == models.BattleTypePvP && b.Player2ID > 0 {
		opp, err := h.playerRepo.GetPlayerByID(context.Background(), b.Player2ID)
		if err == nil && opp != nil {
			opponentName = opp.FirstName
		}
	}

	msg := h.renderer.BattleStateMessage(b, player.ID, player.FirstName, opponentName)
	markup := h.keyboardBuilder.BattleAction(b.ID, b.CurrentTurn == player.ID, len(b.Player1Hand) > 0, len(b.Player1Hand))

	if c.Callback() != nil {
		return c.Edit(msg, markup, tele.ModeHTML)
	}
	return c.Send(msg, markup, tele.ModeHTML)
}

func (h *Handlers) OnCallback(c tele.Context) error {
	data := c.Callback().Data

	if strings.HasPrefix(data, "battle_") {
		return h.handleBattleCallback(c, data)
	}

	if strings.HasPrefix(data, "shop_") {
		return h.handleShopCallback(c, data)
	}

	if strings.HasPrefix(data, "mine_") {
		return h.handleMineCallback(c, data)
	}

	if strings.HasPrefix(data, "chest_") {
		return h.handleChestCallback(c, data)
	}

	if strings.HasPrefix(data, "leader_") {
		return h.handleLeaderCallback(c, data)
	}

	if strings.HasPrefix(data, "deck_") {
		return h.handleDeckCallback(c, data)
	}

	if strings.HasPrefix(data, "confirm_") {
		return h.handleConfirmCallback(c, data)
	}

	return nil
}

func (h *Handlers) handleBattleCallback(c tele.Context, data string) error {
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return nil
	}

	action := parts[0]
	battleID := parts[1]

	ctx := context.Background()
	user := c.Sender()
	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	battle, err := h.battleService.GetBattle(ctx, battleID)
	if err != nil || battle == nil {
		return c.Send(h.renderer.ErrorMessage("Battle not found"), tele.ModeHTML)
	}

	switch action {
	case "battle_play_card":
		var hand []models.CardInstance
		if battle.Player1ID == player.ID {
			hand = battle.Player1Hand
		} else {
			hand = battle.Player2Hand
		}
		msg := "ðŸƒ <b>Select a card to play:</b>"
		return c.Edit(msg, h.keyboardBuilder.CardSelection(hand, battleID), tele.ModeHTML)

	case "battle_select_card":
		if len(parts) < 3 {
			return nil
		}
		instanceID := parts[2]
		msg := "ðŸ“ <b>Choose row:</b>"
		return c.Edit(msg, h.keyboardBuilder.RowSelection(instanceID, battleID), tele.ModeHTML)

	case "battle_row":
		if len(parts) < 4 {
			return nil
		}
		instanceID := parts[2]
		row := parts[3]
		err := h.battleService.PlayCard(ctx, battleID, instanceID, row, player.ID)
		if err != nil {
			return c.Send(h.renderer.ErrorMessage(err.Error()), tele.ModeHTML)
		}
		return h.showBattleState(c, battle, player)

	case "battle_pass":
		err := h.battleService.Pass(ctx, battleID, player.ID)
		if err != nil {
			return c.Send(h.renderer.ErrorMessage(err.Error()), tele.ModeHTML)
		}
		return h.showBattleState(c, battle, player)

	case "battle_leader":
		err := h.battleService.UseLeaderAbility(ctx, battleID, player.ID)
		if err != nil {
			return c.Send(h.renderer.ErrorMessage(err.Error()), tele.ModeHTML)
		}
		return h.showBattleState(c, battle, player)

	case "battle_view_board":
		return h.showBattleState(c, battle, player)

	case "battle_view_hand":
		var hand []models.CardInstance
		if battle.Player1ID == player.ID {
			hand = battle.Player1Hand
		} else {
			hand = battle.Player2Hand
		}
		msg := "ðŸ¤² <b>Your Hand:</b>\n\n"
		for i, card := range hand {
			msg += fmt.Sprintf("%d. %s (Power: %d)\n", i+1, card.InstanceID[:8], card.Power)
		}
		return c.Edit(msg, h.keyboardBuilder.BattleAction(battleID, true, true, len(hand)), tele.ModeHTML)

	case "battle_forfeit":
		msg := "ðŸ³ï¸ <b>Are you sure you want to forfeit?</b>"
		return c.Edit(msg, h.keyboardBuilder.Confirm("forfeit", battleID), tele.ModeHTML)

	case "battle_back":
		return h.showBattleState(c, battle, player)
	}

	return nil
}

func (h *Handlers) handleShopCallback(c tele.Context, data string) error {
	return c.Send(h.renderer.InfoMessage("Shop feature coming soon!"), tele.ModeHTML)
}

func (h *Handlers) handleMineCallback(c tele.Context, data string) error {
	ctx := context.Background()
	user := c.Sender()
	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	switch data {
	case "mine_unlock":
		err := h.progressionService.UnlockMine(ctx, player.ID)
		if err != nil {
			return c.Send(h.renderer.ErrorMessage(err.Error()), tele.ModeHTML)
		}
		return c.Send(h.renderer.SuccessMessage("Mine unlocked!"), tele.ModeHTML)

	case "mine_extract":
		gold, left, err := h.progressionService.ExtractMine(ctx, player.ID)
		if err != nil {
			return c.Send(h.renderer.ErrorMessage(err.Error()), tele.ModeHTML)
		}
		msg := h.renderer.MineExtractMessage(gold, left)
		return c.Send(msg, tele.ModeHTML)

	case "mine_buy_axe":
		err := h.playerRepo.SpendGold(ctx, player.ID, 100)
		if err != nil {
			return c.Send(h.renderer.ErrorMessage("Not enough gold"), tele.ModeHTML)
		}
		durability, err := h.progressionService.ClaimAxe(ctx, player.ID)
		if err != nil {
			return c.Send(h.renderer.ErrorMessage(err.Error()), tele.ModeHTML)
		}
		return c.Send(h.renderer.SuccessMessage(fmt.Sprintf("Axe purchased! Durability: %d", durability)), tele.ModeHTML)
	}

	return nil
}

func (h *Handlers) handleChestCallback(c tele.Context, data string) error {
	return c.Send(h.renderer.InfoMessage("Chest feature coming soon!"), tele.ModeHTML)
}

func (h *Handlers) handleLeaderCallback(c tele.Context, data string) error {
	ctx := context.Background()
	user := c.Sender()
	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	if strings.HasPrefix(data, "leader_select:") {
		leaderID := strings.TrimPrefix(data, "leader_select:")
		player.LeaderID = leaderID
		err := h.playerRepo.UpdatePlayer(ctx, player)
		if err != nil {
			return c.Send(h.renderer.ErrorMessage("Failed to update leader"), tele.ModeHTML)
		}
		return c.Send(h.renderer.SuccessMessage("Leader updated!"), tele.ModeHTML)
	}

	return nil
}

func (h *Handlers) handleDeckCallback(c tele.Context, data string) error {
	return c.Send(h.renderer.InfoMessage("Deck management coming soon!"), tele.ModeHTML)
}

func (h *Handlers) handleConfirmCallback(c tele.Context, data string) error {
	parts := strings.Split(data, ":")
	if len(parts) < 3 {
		return nil
	}

	action := parts[1]
	battleID := parts[2]

	ctx := context.Background()
	user := c.Sender()
	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	if action == "forfeit" {
		battle, err := h.battleService.GetBattle(ctx, battleID)
		if err != nil || battle == nil {
			return c.Send(h.renderer.ErrorMessage("Battle not found"), tele.ModeHTML)
		}

		if battle.Player1ID == player.ID {
			battle.WinnerID = &battle.Player2ID
		} else {
			battle.WinnerID = &battle.Player1ID
		}
		battle.Phase = models.PhaseFinished
		now := time.Now()
		battle.FinishedAt = &now

		err = h.battleRepo.UpdateBattle(ctx, battle)
		if err != nil {
			return c.Send(h.renderer.ErrorMessage("Failed to forfeit"), tele.ModeHTML)
		}

		h.progressionService.RecordBattleResult(ctx, player.ID, false)

		return c.Send(h.renderer.ErrorMessage("You forfeited the battle!"), tele.ModeHTML)
	}

	return nil
}

func (h *Handlers) OnText(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()
	text := strings.ToLower(strings.TrimSpace(c.Text()))

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return nil
	}

	if text == strings.ToLower(h.cfg.DailyCoinWord) {
		reward, err := h.progressionService.ClaimDailyCoin(ctx, player.ID)
		if err != nil {
			if err == progression.ErrDailyCoinOnCooldown {
				return c.Send(h.renderer.InfoMessage("Daily coin already claimed! Come back later."), tele.ModeHTML)
			}
			return c.Send(h.renderer.ErrorMessage(err.Error()), tele.ModeHTML)
		}
		return c.Send(h.renderer.SuccessMessage(fmt.Sprintf("ðŸ’° Daily coin claimed! You received %d gold!", reward)), tele.ModeHTML)
	}

	if text == strings.ToLower(h.cfg.AxeWord) {
		durability, err := h.progressionService.ClaimAxe(ctx, player.ID)
		if err != nil {
			if err == progression.ErrAxeOnCooldown {
				return c.Send(h.renderer.InfoMessage("Axe on cooldown! Come back later."), tele.ModeHTML)
			}
			return c.Send(h.renderer.ErrorMessage(err.Error()), tele.ModeHTML)
		}
		return c.Send(h.renderer.SuccessMessage(fmt.Sprintf("ðŸª“ Axe claimed! Durability: %d", durability)), tele.ModeHTML)
	}

	if text == strings.ToLower(h.cfg.StartWord) {
		return h.OnStart(c)
	}

	return nil
}

func (h *Handlers) getAvailableChests(level models.PlayerLevel) []models.Chest {
	allChests := []models.Chest{
		{
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
		{
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
		{
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

	var available []models.Chest
	for _, chest := range allChests {
		if chest.LevelRequired <= level {
			available = append(available, chest)
		}
	}

	return available
}
