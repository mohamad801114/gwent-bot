package handlers

import (
	"context"
	"fmt"
	"html"
	"log"
	"os"
	"strconv"
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
	db                 *database.DB
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
	db *database.DB,
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
		db:                 db,
	}
}

func (h *Handlers) RegisterRoutes() {
	// Public routes
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
	h.bot.Handle("/testcb", h.OnTestCallback)

	// Admin routes (with middleware)
	h.bot.Handle("/admin", h.AdminMiddleware(h.OnAdmin))
	h.bot.Handle("/admin_cards", h.AdminMiddleware(h.OnAdminCards))
	h.bot.Handle("/admin_shop", h.AdminMiddleware(h.OnAdminShop))
	h.bot.Handle("/admin_shop_add", h.AdminMiddleware(h.OnAdminShopAdd))
	h.bot.Handle("/admin_shop_remove", h.AdminMiddleware(h.OnAdminShopRemove))
	h.bot.Handle("/admin_rarity", h.AdminMiddleware(h.OnAdminRarity))
	h.bot.Handle("/admin_players", h.AdminMiddleware(h.OnAdminPlayers))
	h.bot.Handle("/admin_set_level", h.AdminMiddleware(h.OnAdminSetLevel))
	h.bot.Handle("/admin_add_exp", h.AdminMiddleware(h.OnAdminAddExp))
	h.bot.Handle("/admin_battles", h.AdminMiddleware(h.OnAdminBattles))

	h.bot.Handle("main_menu", h.OnMainMenu)
	h.bot.Handle("battle_menu", h.OnBattleMenu)
	h.bot.Handle("show_deck", h.OnShowDeck)
	h.bot.Handle("show_collection", h.OnShowCollection)
	h.bot.Handle("show_profile", h.OnShowProfile)
	h.bot.Handle("show_help", h.OnShowHelp)
	h.bot.Handle("shop_menu", h.OnShopMenu)
	h.bot.Handle("mine_menu", h.OnMineMenu)
	h.bot.Handle("chest_menu", h.OnChestMenu)
	h.bot.Handle("leader_menu", h.OnLeaderMenu)

	h.bot.Handle("battle_pve_easy", h.OnBattlePvEEasy)
	h.bot.Handle("battle_pve_medium", h.OnBattlePvEMedium)
	h.bot.Handle("battle_pve_hard", h.OnBattlePvEHard)
	h.bot.Handle("battle_pvp", h.OnBattlePvP)
	h.bot.Handle("battle_friend", h.OnBattleFriend)

	h.bot.Handle(tele.OnCallback, h.OnCallback)

	h.bot.Handle(tele.OnText, h.OnText)
}

// AdminMiddleware checks if the user is an admin
func (h *Handlers) AdminMiddleware(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		user := c.Sender()
		if !h.isAdmin(user.ID) {
			return c.Send(h.renderer.ErrorMessage("Admin access required"), tele.ModeHTML)
		}
		return next(c)
	}
}

func (h *Handlers) isAdmin(userID int64) bool {
	for _, adminID := range h.cfg.AdminUserIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}

// answerCallback answers the callback query to remove loading state
func (h *Handlers) answerCallback(c tele.Context) {
	if cb := c.Callback(); cb != nil {
		if err := c.Bot().Respond(cb, &tele.CallbackResponse{Text: ""}); err != nil {
			log.Printf("ERROR: Failed to answer callback query in answerCallback: %v", err)
		}
	}
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

	// Send photos for each card in deck
	for _, card := range deckCards {
		if card.ArtURL != "" {
			if _, err := os.Stat(card.ArtURL); err == nil {
				photo := &tele.Photo{
					File:    tele.FromDisk(card.ArtURL),
					Caption: h.renderer.CardCaption(&card),
				}
				c.Send(photo, tele.ModeHTML)
			}
		}
	}

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

	// Send photos for each owned card (limit to first 10 to avoid spam)
	count := 0
	for _, card := range ownedCards {
		if count >= 10 {
			break
		}
		if card.ArtURL != "" {
			if _, err := os.Stat(card.ArtURL); err == nil {
				photo := &tele.Photo{
					File:    tele.FromDisk(card.ArtURL),
					Caption: h.renderer.CardCaption(&card),
				}
				c.Send(photo, tele.ModeHTML)
				count++
			}
		}
	}

	return c.Send(msg, h.keyboardBuilder.Pagination(1, 1, "collection"), tele.ModeHTML)
}

func (h *Handlers) OnBattle(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		if c.Callback() != nil {
			return c.Edit(h.renderer.NotRegisteredMessage(), h.keyboardBuilder.MainMenu(), tele.ModeHTML)
		}
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	activeBattle, err := h.battleService.GetActiveBattle(ctx, player.ID)
	if err == nil && activeBattle != nil {
		return h.showBattleState(c, activeBattle, player)
	}

	msg := "⚔️ <b>Battle Menu</b>\n\nChoose your battle type:"
	if c.Callback() != nil {
		return c.Edit(msg, h.keyboardBuilder.BattleMenu(), tele.ModeHTML)
	}
	return c.Send(msg, h.keyboardBuilder.BattleMenu(), tele.ModeHTML)
}

func (h *Handlers) OnShop(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		if c.Callback() != nil {
			return c.Edit(h.renderer.NotRegisteredMessage(), h.keyboardBuilder.MainMenu(), tele.ModeHTML)
		}
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	msg := h.renderer.ShopMessage(player)
	if c.Callback() != nil {
		return c.Edit(msg, h.keyboardBuilder.ShopMenu(), tele.ModeHTML)
	}
	return c.Send(msg, h.keyboardBuilder.ShopMenu(), tele.ModeHTML)
}

func (h *Handlers) OnShopCards(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Edit(h.renderer.NotRegisteredMessage(), h.keyboardBuilder.MainMenu(), tele.ModeHTML)
	}

	// Get shop items from database
	shopItems, err := h.cardRepo.GetShopItems(ctx)
	if err != nil {
		return c.Edit(h.renderer.ErrorMessage("Failed to load shop"), h.keyboardBuilder.ShopMenu(), tele.ModeHTML)
	}

	// Get card details for each shop item
	var buyableCards []models.Card
	for _, item := range shopItems {
		if item.CardID != "" && (item.CostGold > 0 || item.CostScrap > 0) {
			card, err := h.cardService.GetCard(ctx, item.CardID)
			if err == nil && card != nil {
				buyableCards = append(buyableCards, *card)
			}
		}
	}

	msg := h.renderer.ShopCardsMessage(buyableCards, 1, 1)

	// Send photos for first few cards (non-blocking)
	count := 0
	for _, card := range buyableCards {
		if count >= 5 {
			break
		}
		if card.ArtURL != "" {
			if _, err := os.Stat(card.ArtURL); err == nil {
				photo := &tele.Photo{
					File:    tele.FromDisk(card.ArtURL),
					Caption: h.renderer.CardCaption(&card),
				}
				if err := c.Send(photo, tele.ModeHTML); err != nil {
					log.Printf("DEBUG OnShopCards: failed to send photo: %v", err)
				}
				count++
			}
		}
	}

	// Edit the callback message with shop cards
	err = c.Edit(msg, h.keyboardBuilder.ShopCards(buyableCards, 1, 1), tele.ModeHTML)
	if err != nil {
		log.Printf("DEBUG OnShopCards: failed to edit message: %v", err)
	}
	return err
}

func (h *Handlers) OnShopChests(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Edit(h.renderer.NotRegisteredMessage(), h.keyboardBuilder.MainMenu(), tele.ModeHTML)
	}

	// Get shop items for chests (items with no card_id)
	shopItems, err := h.cardRepo.GetShopItems(ctx)
	if err != nil {
		return c.Edit(h.renderer.ErrorMessage("Failed to load shop"), h.keyboardBuilder.ShopMenu(), tele.ModeHTML)
	}

	// Filter chest items (no card_id)
	var chestItems []*models.ShopItem
	for _, item := range shopItems {
		if item.CardID == "" && item.CostGold > 0 {
			chestItems = append(chestItems, item)
		}
	}

	// Convert to Chest models for rendering
	var chests []models.Chest
	for _, item := range chestItems {
		if item.LevelRequired <= int(player.Level) {
			chests = append(chests, models.Chest{
				ID:            item.ID,
				Name:          item.Name,
				Description:   item.Description,
				LevelRequired: item.LevelRequired,
				Gold:          item.CostGold,
				Cooldown:      24 * time.Hour,
			})
		}
	}

	msg := h.renderer.ChestMessage(chests, player.Level)
	err = c.Edit(msg, h.keyboardBuilder.ChestMenu(chests, player.Level), tele.ModeHTML)
	if err != nil {
		log.Printf("DEBUG OnShopChests: edit error: %v", err)
	}
	return err
}

func (h *Handlers) OnMine(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		if c.Callback() != nil {
			return c.Edit(h.renderer.NotRegisteredMessage(), h.keyboardBuilder.MainMenu(), tele.ModeHTML)
		}
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	msg := h.renderer.MineMessage(player, h.cfg)
	if c.Callback() != nil {
		return c.Edit(msg, h.keyboardBuilder.MineMenu(player.MineUnlocked, h.cfg.MineMaxExtracts-player.MineExtractsToday, player.AxeDurability), tele.ModeHTML)
	}
	return c.Send(msg, h.keyboardBuilder.MineMenu(player.MineUnlocked, h.cfg.MineMaxExtracts-player.MineExtractsToday, player.AxeDurability), tele.ModeHTML)
}

func (h *Handlers) OnChest(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		if c.Callback() != nil {
			return c.Edit(h.renderer.NotRegisteredMessage(), h.keyboardBuilder.MainMenu(), tele.ModeHTML)
		}
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	// Get shop items for chests (items with no card_id)
	shopItems, err := h.cardRepo.GetShopItems(ctx)
	if err != nil {
		if c.Callback() != nil {
			return c.Edit(h.renderer.ErrorMessage("Failed to load chests"), h.keyboardBuilder.MainMenu(), tele.ModeHTML)
		}
		return c.Send(h.renderer.ErrorMessage("Failed to load chests"), tele.ModeHTML)
	}

	// Filter chest items (no card_id)
	var chestItems []*models.ShopItem
	for _, item := range shopItems {
		if item.CardID == "" && item.CostGold > 0 {
			chestItems = append(chestItems, item)
		}
	}

	// Convert to Chest models for rendering
	var chests []models.Chest
	for _, item := range chestItems {
		if item.LevelRequired <= int(player.Level) {
			chests = append(chests, models.Chest{
				ID:            item.ID,
				Name:          item.Name,
				Description:   item.Description,
				LevelRequired: item.LevelRequired,
				Gold:          item.CostGold,
				Cooldown:      24 * time.Hour,
			})
		}
	}

	msg := h.renderer.ChestMessage(chests, player.Level)
	if c.Callback() != nil {
		return c.Edit(msg, h.keyboardBuilder.ChestMenu(chests, player.Level), tele.ModeHTML)
	}
	return c.Send(msg, h.keyboardBuilder.ChestMenu(chests, player.Level), tele.ModeHTML)
}

func (h *Handlers) OnLeader(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		if c.Callback() != nil {
			return c.Edit(h.renderer.NotRegisteredMessage(), h.keyboardBuilder.MainMenu(), tele.ModeHTML)
		}
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	leadersPtr, err := h.cardRepo.GetCardsByType(ctx, models.CardTypeLeader)
	if err != nil {
		if c.Callback() != nil {
			return c.Edit(h.renderer.ErrorMessage("Failed to load leaders"), h.keyboardBuilder.MainMenu(), tele.ModeHTML)
		}
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
	if c.Callback() != nil {
		return c.Edit(msg, h.keyboardBuilder.LeaderSelection(leaders, player.LeaderID), tele.ModeHTML)
	}
	return c.Send(msg, h.keyboardBuilder.LeaderSelection(leaders, player.LeaderID), tele.ModeHTML)
}

func (h *Handlers) OnMainMenu(c tele.Context) error {
	h.answerCallback(c)
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
	h.answerCallback(c)
	msg := "⚔️ <b>Battle Menu</b>\n\nChoose your battle type:"
	return c.Edit(msg, h.keyboardBuilder.BattleMenu(), tele.ModeHTML)
}

func (h *Handlers) OnShowDeck(c tele.Context) error {
	h.answerCallback(c)
	return h.OnDeck(c)
}

func (h *Handlers) OnShowCollection(c tele.Context) error {
	h.answerCallback(c)
	return h.OnCollection(c)
}

func (h *Handlers) OnShowProfile(c tele.Context) error {
	h.answerCallback(c)
	return h.OnProfile(c)
}

func (h *Handlers) OnShowHelp(c tele.Context) error {
	h.answerCallback(c)
	msg := h.renderer.HelpMessage()
	return c.Edit(msg, h.keyboardBuilder.MainMenu(), tele.ModeHTML)
}

func (h *Handlers) OnShopMenu(c tele.Context) error {
	h.answerCallback(c)
	return h.OnShop(c)
}

func (h *Handlers) OnMineMenu(c tele.Context) error {
	h.answerCallback(c)
	return h.OnMine(c)
}

func (h *Handlers) OnChestMenu(c tele.Context) error {
	h.answerCallback(c)
	return h.OnChest(c)
}

func (h *Handlers) OnLeaderMenu(c tele.Context) error {
	h.answerCallback(c)
	return h.OnLeader(c)
}

func (h *Handlers) OnBattlePvEEasy(c tele.Context) error {
	h.answerCallback(c)
	return h.createPvEBattle(c, "easy")
}

func (h *Handlers) OnBattlePvEMedium(c tele.Context) error {
	h.answerCallback(c)
	return h.createPvEBattle(c, "medium")
}

func (h *Handlers) OnBattlePvEHard(c tele.Context) error {
	h.answerCallback(c)
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
	h.answerCallback(c)
	return c.Send(h.renderer.InfoMessage("PvP matchmaking coming soon!"), tele.ModeHTML)
}

func (h *Handlers) OnBattleFriend(c tele.Context) error {
	h.answerCallback(c)
	return h.createFriendBattle(c)
}

func (h *Handlers) createFriendBattle(c tele.Context) error {
	ctx := context.Background()
	user := c.Sender()

	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	// Check if in group chat
	chat := c.Chat()
	if chat == nil || chat.Type != tele.ChatGroup && chat.Type != tele.ChatSuperGroup {
		return c.Send(h.renderer.ErrorMessage("Friend battles can only be started in group chats!"), tele.ModeHTML)
	}

	// Create a friend invite that other players in the group can accept
	msg := fmt.Sprintf(
		"⚔️ <b>%s</b> wants to play a friendly match!\n\n"+
			"Click the button below to accept the challenge.",
		player.FirstName,
	)

	return c.Send(msg, h.keyboardBuilder.FriendBattleInvite(player.ID), tele.ModeHTML)
}

// Admin command handlers
func (h *Handlers) OnAdmin(c tele.Context) error {
	msg := `🔧 <b>Admin Panel</b>

Available commands:
/admin_cards - Manage cards
/admin_shop - Manage shop items
/admin_rarity - Manage rarities
/admin_players - View all players
/admin_set_level <telegram_id> <level> - Set player level
/admin_add_exp <telegram_id> <amount> - Add experience
/admin_battles - View battles`

	return c.Send(msg, tele.ModeHTML)
}

func (h *Handlers) OnAdminCards(c tele.Context) error {
	ctx := context.Background()
	cards, err := h.cardRepo.GetAllCards(ctx)
	if err != nil {
		return c.Send(h.renderer.ErrorMessage("Failed to load cards"), tele.ModeHTML)
	}

	var sb strings.Builder
	sb.WriteString("🃏 <b>All Cards</b>\n\n")
	for _, card := range cards {
		sb.WriteString(fmt.Sprintf("%s %s [%s] P:%d (ID: %s)\n",
			card.Rarity.Emoji(),
			html.EscapeString(card.Name),
			card.Nation.Emoji(),
			card.Provision,
			html.EscapeString(card.ID),
		))
	}

	return c.Send(sb.String(), tele.ModeHTML)
}

func (h *Handlers) OnAdminShop(c tele.Context) error {
	ctx := context.Background()
	shopItems, err := h.cardRepo.GetShopItems(ctx)
	if err != nil {
		return c.Send(h.renderer.ErrorMessage("Failed to load shop items"), tele.ModeHTML)
	}

	var sb strings.Builder
	sb.WriteString("🏪 <b>Shop Items</b>\n\n")
	if len(shopItems) == 0 {
		sb.WriteString("No shop items configured.\n\n")
	} else {
		for _, item := range shopItems {
			cardName := "N/A"
			if item.CardID != "" {
				card, err := h.cardService.GetCard(ctx, item.CardID)
				if err == nil && card != nil {
					cardName = card.Name
				}
			}
			sb.WriteString(fmt.Sprintf("📦 %s (ID: %s)\n", html.EscapeString(item.Name), html.EscapeString(item.ID)))
			sb.WriteString(fmt.Sprintf("   Card: %s (ID: %s)\n", html.EscapeString(cardName), html.EscapeString(item.CardID)))
			sb.WriteString(fmt.Sprintf("   Cost: %d Gold / %d Scrap | Stock: %d | Level: %d\n\n",
				item.CostGold, item.CostScrap, item.Stock, item.LevelRequired))
		}
	}
	sb.WriteString("\nUse /admin_shop_add <card_id> <cost_gold> <cost_scrap> [level] to add item\n")
	sb.WriteString("Use /admin_shop_remove <item_id> to remove item")

	return c.Send(sb.String(), tele.ModeHTML)
}

func (h *Handlers) OnAdminShopAdd(c tele.Context) error {
	if !h.isAdmin(c.Sender().ID) {
		return c.Send(h.renderer.ErrorMessage("Admin access required"), tele.ModeHTML)
	}

	args := strings.Fields(c.Text())
	if len(args) < 4 {
		return c.Send(h.renderer.ErrorMessage("Usage: /admin_shop_add <card_id> <cost_gold> <cost_scrap> [level_required]"), tele.ModeHTML)
	}

	cardID := args[1]
	costGold, err := strconv.Atoi(args[2])
	if err != nil {
		return c.Send(h.renderer.ErrorMessage("Invalid cost_gold"), tele.ModeHTML)
	}
	costScrap, err := strconv.Atoi(args[3])
	if err != nil {
		return c.Send(h.renderer.ErrorMessage("Invalid cost_scrap"), tele.ModeHTML)
	}
	levelRequired := 1
	if len(args) >= 5 {
		levelRequired, _ = strconv.Atoi(args[4])
	}

	ctx := context.Background()
	// Verify card exists
	card, err := h.cardService.GetCard(ctx, cardID)
	if err != nil || card == nil {
		return c.Send(h.renderer.ErrorMessage("Card not found"), tele.ModeHTML)
	}

	itemID := "shop_card_" + cardID
	_, err = h.db.Pool.Exec(ctx,
		`INSERT INTO shop_items (id, name, description, card_id, cost_gold, cost_scrap, stock, level_required)
		 VALUES ($1, $2, $3, $4, $5, $6, -1, $7)
		 ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			card_id = EXCLUDED.card_id,
			cost_gold = EXCLUDED.cost_gold,
			cost_scrap = EXCLUDED.cost_scrap,
			level_required = EXCLUDED.level_required`,
		itemID, card.Name, "Buy "+card.Name, cardID, costGold, costScrap, levelRequired)
	if err != nil {
		log.Printf("Error adding shop item: %v", err)
		return c.Send(h.renderer.ErrorMessage("Failed to add shop item"), tele.ModeHTML)
	}

	return c.Send(h.renderer.SuccessMessage(fmt.Sprintf("Added %s to shop (Gold: %d, Scrap: %d)", card.Name, costGold, costScrap)), tele.ModeHTML)
}

func (h *Handlers) OnAdminShopRemove(c tele.Context) error {
	if !h.isAdmin(c.Sender().ID) {
		return c.Send(h.renderer.ErrorMessage("Admin access required"), tele.ModeHTML)
	}

	args := strings.Fields(c.Text())
	if len(args) < 2 {
		return c.Send(h.renderer.ErrorMessage("Usage: /admin_shop_remove <item_id>"), tele.ModeHTML)
	}

	itemID := args[1]
	ctx := context.Background()

	_, err := h.db.Pool.Exec(ctx, `DELETE FROM shop_items WHERE id = $1`, itemID)
	if err != nil {
		log.Printf("Error removing shop item: %v", err)
		return c.Send(h.renderer.ErrorMessage("Failed to remove shop item"), tele.ModeHTML)
	}

	return c.Send(h.renderer.SuccessMessage("Shop item removed"), tele.ModeHTML)
}

func (h *Handlers) OnAdminRarity(c tele.Context) error {
	msg := `🎨 <b>Rarity Management</b>

Rarities available:
• Common (⚪)
• Rare (🔵)
• Epic (🟣)
• Legendary (🟡)

Use /admin_cards to view cards by rarity.`

	return c.Send(msg, tele.ModeHTML)
}

func (h *Handlers) OnAdminPlayers(c tele.Context) error {
	ctx := context.Background()
	players, err := h.playerRepo.GetAllPlayers(ctx)
	if err != nil {
		return c.Send(h.renderer.ErrorMessage("Failed to load players"), tele.ModeHTML)
	}

	var sb strings.Builder
	sb.WriteString("👥 <b>All Players (Top 100)</b>\n\n")
	if len(players) == 0 {
		sb.WriteString("No players found.\n\n")
	} else {
		for _, p := range players {
			sb.WriteString(fmt.Sprintf("%d. %s (@%s) - ID: %d\n", p.ID, html.EscapeString(p.FirstName), html.EscapeString(p.Username), p.TelegramID))
			sb.WriteString(fmt.Sprintf("   Level: %d | Exp: %d | Gold: %d | Scrap: %d\n", p.Level, p.Experience, p.Gold, p.Scrap))
			sb.WriteString(fmt.Sprintf("   Games: %dW/%dL | Cards: %d | Deck: %d\n\n",
				p.GamesWon, p.GamesLost, len(p.Collection), len(p.Deck)))
		}
	}
	sb.WriteString("\nUse /admin_set_level <telegram_id> <level> to set player level")
	sb.WriteString("\nUse /admin_add_exp <telegram_id> <amount> to add experience")

	return c.Send(sb.String(), tele.ModeHTML)
}

func (h *Handlers) OnAdminSetLevel(c tele.Context) error {
	if !h.isAdmin(c.Sender().ID) {
		return c.Send(h.renderer.ErrorMessage("Admin access required"), tele.ModeHTML)
	}

	args := strings.Fields(c.Text())
	if len(args) < 3 {
		return c.Send(h.renderer.ErrorMessage("Usage: /admin_set_level <telegram_id> <level>"), tele.ModeHTML)
	}

	telegramID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return c.Send(h.renderer.ErrorMessage("Invalid telegram_id"), tele.ModeHTML)
	}

	level, err := strconv.Atoi(args[2])
	if err != nil {
		return c.Send(h.renderer.ErrorMessage("Invalid level"), tele.ModeHTML)
	}

	ctx := context.Background()
	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, telegramID)
	if err != nil || player == nil {
		return c.Send(h.renderer.ErrorMessage("Player not found"), tele.ModeHTML)
	}

	player.Level = models.PlayerLevel(level)
	err = h.playerRepo.UpdatePlayer(ctx, player)
	if err != nil {
		log.Printf("Error updating player level: %v", err)
		return c.Send(h.renderer.ErrorMessage("Failed to update player level"), tele.ModeHTML)
	}

	return c.Send(h.renderer.SuccessMessage(fmt.Sprintf("Set %s's level to %d", player.FirstName, level)), tele.ModeHTML)
}

func (h *Handlers) OnAdminAddExp(c tele.Context) error {
	if !h.isAdmin(c.Sender().ID) {
		return c.Send(h.renderer.ErrorMessage("Admin access required"), tele.ModeHTML)
	}

	args := strings.Fields(c.Text())
	if len(args) < 3 {
		return c.Send(h.renderer.ErrorMessage("Usage: /admin_add_exp <telegram_id> <amount>"), tele.ModeHTML)
	}

	telegramID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return c.Send(h.renderer.ErrorMessage("Invalid telegram_id"), tele.ModeHTML)
	}

	amount, err := strconv.Atoi(args[2])
	if err != nil {
		return c.Send(h.renderer.ErrorMessage("Invalid amount"), tele.ModeHTML)
	}

	ctx := context.Background()
	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, telegramID)
	if err != nil || player == nil {
		return c.Send(h.renderer.ErrorMessage("Player not found"), tele.ModeHTML)
	}

	player.Experience += amount
	// Auto-level up based on experience (simple formula: level = 1 + exp/1000)
	newLevel := models.PlayerLevel(1 + player.Experience/1000)
	if newLevel > player.Level {
		player.Level = newLevel
	}

	err = h.playerRepo.UpdatePlayer(ctx, player)
	if err != nil {
		log.Printf("Error adding exp: %v", err)
		return c.Send(h.renderer.ErrorMessage("Failed to add experience"), tele.ModeHTML)
	}

	return c.Send(h.renderer.SuccessMessage(fmt.Sprintf("Added %d exp to %s (Level: %d)", amount, player.FirstName, player.Level)), tele.ModeHTML)
}

func (h *Handlers) OnAdminBattles(c tele.Context) error {
	return c.Send(h.renderer.InfoMessage("Battle management coming soon!"), tele.ModeHTML)
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
	cb := c.Callback()
	if cb != nil {
		// Answer callback query to remove loading state
		if err := c.Bot().Respond(cb, &tele.CallbackResponse{
			Text: "",
		}); err != nil {
			log.Printf("ERROR: Failed to answer callback query: %v", err)
		}
	}

	data := cb.Data

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

	if strings.HasPrefix(data, "friend_") {
		return h.handleFriendCallback(c, data)
	}

	if strings.HasPrefix(data, "test_") {
		return h.handleTestCallback(c, data)
	}

	return nil
}

func (h *Handlers) handleTestCallback(c tele.Context, data string) error {
	if data == "test_click" {
		return c.Send("✅ Callback works! Button click detected.", tele.ModeHTML)
	}
	return nil
}

// DEBUG: Test callback query
func (h *Handlers) OnTestCallback(c tele.Context) error {
	markup := &tele.ReplyMarkup{}
	row := markup.Row(markup.Data("🔘 Test Click", "test_click"))
	markup.Inline(row)
	return c.Send("🔧 <b>Callback Test</b>\n\nClick the button below to test if callbacks work:", markup, tele.ModeHTML)
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
	log.Printf("DEBUG handleShopCallback: data=%s", data)
	ctx := context.Background()
	user := c.Sender()
	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		log.Printf("DEBUG handleShopCallback: player not found, err=%v", err)
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	log.Printf("DEBUG handleShopCallback: player found, id=%d", player.ID)

	// Parse callback data
	if strings.HasPrefix(data, "shop_buy_card:") {
		cardID := strings.TrimPrefix(data, "shop_buy_card:")
		log.Printf("DEBUG handleShopCallback: buying card %s", cardID)
		return h.buyCard(c, ctx, player, cardID)
	}

	if strings.HasPrefix(data, "shop_buy_chest:") {
		chestID := strings.TrimPrefix(data, "shop_buy_chest:")
		log.Printf("DEBUG handleShopCallback: buying chest %s", chestID)
		return h.buyChest(c, ctx, player, chestID)
	}

	if data == "shop_axe" {
		log.Printf("DEBUG handleShopCallback: buying axe")
		return h.buyAxe(c, ctx, player)
	}

	if data == "shop_cards" {
		log.Printf("DEBUG handleShopCallback: showing shop cards")
		return h.OnShopCards(c)
	}

	if data == "shop_chests" {
		log.Printf("DEBUG handleShopCallback: showing shop chests")
		return h.OnShopChests(c)
	}

	if data == "shop_main" {
		log.Printf("DEBUG handleShopCallback: back to shop main")
		return h.OnShop(c)
	}

	log.Printf("DEBUG handleShopCallback: unknown data %s", data)
	return nil
}

func (h *Handlers) buyCard(c tele.Context, ctx context.Context, player *models.Player, cardID string) error {
	card, err := h.cardService.GetCard(ctx, cardID)
	if err != nil || card == nil {
		return c.Send(h.renderer.ErrorMessage("Card not found"), tele.ModeHTML)
	}

	// Check if player already has this card in collection
	for _, colCard := range player.Collection {
		if colCard.ID == cardID {
			return c.Send(h.renderer.ErrorMessage("You already own this card!"), tele.ModeHTML)
		}
	}

	// Check gold cost
	if card.CostGold > 0 {
		if player.Gold < card.CostGold {
			return c.Send(h.renderer.ErrorMessage(fmt.Sprintf("Not enough gold! Need %d, have %d", card.CostGold, player.Gold)), tele.ModeHTML)
		}
		err := h.playerRepo.SpendGold(ctx, player.ID, card.CostGold)
		if err != nil {
			return c.Send(h.renderer.ErrorMessage("Failed to spend gold"), tele.ModeHTML)
		}
		player.Gold -= card.CostGold
	}

	// Check scrap cost
	if card.CostScrap > 0 {
		if player.Scrap < card.CostScrap {
			return c.Send(h.renderer.ErrorMessage(fmt.Sprintf("Not enough scrap! Need %d, have %d", card.CostScrap, player.Scrap)), tele.ModeHTML)
		}
		err := h.playerRepo.SpendScrap(ctx, player.ID, card.CostScrap)
		if err != nil {
			return c.Send(h.renderer.ErrorMessage("Failed to spend scrap"), tele.ModeHTML)
		}
		player.Scrap -= card.CostScrap
	}

	// Add card to collection
	err = h.playerRepo.AddCardToCollection(ctx, player.ID, cardID)
	if err != nil {
		return c.Send(h.renderer.ErrorMessage("Failed to add card to collection"), tele.ModeHTML)
	}

	// Send card photo
	if card.ArtURL != "" {
		if _, err := os.Stat(card.ArtURL); err == nil {
			photo := &tele.Photo{
				File:    tele.FromDisk(card.ArtURL),
				Caption: fmt.Sprintf("✅ <b>Purchased!</b>\n\n%s", h.renderer.CardCaption(card)),
			}
			c.Send(photo, tele.ModeHTML)
		}
	}

	return c.Send(h.renderer.SuccessMessage(fmt.Sprintf("Successfully purchased %s!", card.Name)), tele.ModeHTML)
}

func (h *Handlers) buyChest(c tele.Context, ctx context.Context, player *models.Player, chestID string) error {
	// Get shop item for this chest
	shopItem, err := h.cardRepo.GetShopItem(ctx, chestID)
	if err != nil {
		return c.Send(h.renderer.ErrorMessage("Failed to load chest"), tele.ModeHTML)
	}

	if shopItem == nil || shopItem.CardID != "" {
		return c.Send(h.renderer.ErrorMessage("Chest not found"), tele.ModeHTML)
	}

	if shopItem.LevelRequired > int(player.Level) {
		return c.Send(h.renderer.ErrorMessage(fmt.Sprintf("Chest requires level %d", shopItem.LevelRequired)), tele.ModeHTML)
	}

	cost := shopItem.CostGold
	if player.Gold < cost {
		return c.Send(h.renderer.ErrorMessage(fmt.Sprintf("Not enough gold! Need %d, have %d", cost, player.Gold)), tele.ModeHTML)
	}

	err = h.playerRepo.SpendGold(ctx, player.ID, cost)
	if err != nil {
		return c.Send(h.renderer.ErrorMessage("Failed to spend gold"), tele.ModeHTML)
	}

	// Open chest and give rewards (simplified)
	// In a real implementation, this would use the progression service
	return c.Send(h.renderer.SuccessMessage(fmt.Sprintf("Purchased %s for %d gold! Chest opening coming soon.", shopItem.Name, cost)), tele.ModeHTML)
}

func (h *Handlers) buyAxe(c tele.Context, ctx context.Context, player *models.Player) error {
	log.Printf("DEBUG buyAxe: started for player %d", player.ID)
	cost := 100
	if player.Gold < cost {
		log.Printf("DEBUG buyAxe: not enough gold")
		return c.Send(h.renderer.ErrorMessage(fmt.Sprintf("Not enough gold! Need %d, have %d", cost, player.Gold)), tele.ModeHTML)
	}

	err := h.playerRepo.SpendGold(ctx, player.ID, cost)
	if err != nil {
		log.Printf("DEBUG buyAxe: SpendGold error: %v", err)
		return c.Send(h.renderer.ErrorMessage("Failed to spend gold"), tele.ModeHTML)
	}

	durability, err := h.progressionService.ClaimAxe(ctx, player.ID)
	if err != nil {
		log.Printf("DEBUG buyAxe: ClaimAxe error: %v", err)
		return c.Send(h.renderer.ErrorMessage(err.Error()), tele.ModeHTML)
	}

	err = c.Send(h.renderer.SuccessMessage(fmt.Sprintf("Axe purchased! Durability: %d", durability)), tele.ModeHTML)
	if err != nil {
		log.Printf("DEBUG buyAxe: send error: %v", err)
	}
	log.Printf("DEBUG buyAxe: done")
	return err
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

func (h *Handlers) handleFriendCallback(c tele.Context, data string) error {
	ctx := context.Background()
	user := c.Sender()
	player, err := h.playerRepo.GetPlayerByTelegramID(ctx, user.ID)
	if err != nil || player == nil {
		return c.Send(h.renderer.NotRegisteredMessage(), tele.ModeHTML)
	}

	if strings.HasPrefix(data, "friend_accept:") {
		inviterIDStr := strings.TrimPrefix(data, "friend_accept:")
		inviterID, err := strconv.ParseInt(inviterIDStr, 10, 64)
		if err != nil {
			return c.Send(h.renderer.ErrorMessage("Invalid invite"), tele.ModeHTML)
		}

		// Check if inviter exists
		inviter, err := h.playerRepo.GetPlayerByID(ctx, inviterID)
		if err != nil || inviter == nil {
			return c.Send(h.renderer.ErrorMessage("Inviter not found"), tele.ModeHTML)
		}

		// Create PvP battle between the two players
		battle, err := h.battleService.CreatePvPBattle(ctx, inviterID, player.ID)
		if err != nil {
			return c.Send(h.renderer.ErrorMessage("Failed to create battle"), tele.ModeHTML)
		}

		// Send battle state to both players
		msg := h.renderer.BattleStateMessage(battle, player.ID, player.FirstName, inviter.FirstName)
		markup := h.keyboardBuilder.BattleAction(battle.ID, battle.CurrentTurn == player.ID, len(battle.Player1Hand) > 0, len(battle.Player1Hand))

		// Send to the player who accepted
		c.Send(msg, markup, tele.ModeHTML)

		// Send to the inviter (would need to use bot.Send with chat ID in real implementation)
		// For now, just notify the inviter
		h.bot.Send(&tele.User{ID: inviter.TelegramID}, fmt.Sprintf("⚔️ %s accepted your challenge! Battle started.", player.FirstName), tele.ModeHTML)

		return nil
	}

	if data == "friend_decline" {
		return c.Send(h.renderer.InfoMessage("Challenge declined."), tele.ModeHTML)
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
	playerLevel := int(level)
	for _, chest := range allChests {
		if chest.LevelRequired <= playerLevel {
			available = append(available, chest)
		}
	}

	return available
}
