package keyboards

import (
	"fmt"
	"strconv"

	tele "gopkg.in/telebot.v3"

	"gwent-bot/internal/config"
	"gwent-bot/pkg/models"
)

type KeyboardBuilder struct {
	cfg *config.Config
}

func New(cfg *config.Config) *KeyboardBuilder {
	return &KeyboardBuilder{cfg: cfg}
}

func (k *KeyboardBuilder) MainMenu() *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{ResizeKeyboard: true}

	btnBattle := markup.Data("⚔️ Battle", "battle_menu")
	btnDeck := markup.Data("🃏 My Deck", "show_deck")
	btnCollection := markup.Data("📚 Collection", "show_collection")
	btnShop := markup.Data("🏪 Shop", "shop_menu")
	btnMine := markup.Data("⛏️ Mine", "mine_menu")
	btnChest := markup.Data("🎁 Chests", "chest_menu")
	btnProfile := markup.Data("👤 Profile", "show_profile")
	btnLeader := markup.Data("💎 Leader", "leader_menu")
	btnHelp := markup.Data("❓ Help", "show_help")

	markup.Row(btnBattle)
	markup.Row(btnDeck, btnCollection)
	markup.Row(btnShop, btnMine)
	markup.Row(btnChest, btnProfile)
	markup.Row(btnLeader, btnHelp)

	return markup
}

func (k *KeyboardBuilder) BattleMenu() *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{ResizeKeyboard: true}

	btnPvE := markup.Data("🤖 vs Bot (Easy)", "battle_pve_easy")
	btnPvEMedium := markup.Data("🤖 vs Bot (Medium)", "battle_pve_medium")
	btnPvEHard := markup.Data("🤖 vs Bot (Hard)", "battle_pve_hard")
	btnPvP := markup.Data("👥 vs Player", "battle_pvp")
	btnFriend := markup.Data("🤝 vs Friend", "battle_friend")
	btnBack := markup.Data("🔙 Back", "main_menu")

	markup.Row(btnPvE)
	markup.Row(btnPvEMedium)
	markup.Row(btnPvEHard)
	markup.Row(btnPvP, btnFriend)
	markup.Row(btnBack)

	return markup
}

func (k *KeyboardBuilder) BattleAction(battleID string, isPlayerTurn bool, canPlayCard bool, handSize int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	if !isPlayerTurn {
		btnWait := markup.Data("⏳ Opponent's turn...", "battle_wait")
		markup.Row(btnWait)
		return markup
	}

	if canPlayCard && handSize > 0 {
		btnPlay := markup.Data("🃏 Play Card", "battle_play_card:"+battleID)
		markup.Row(btnPlay)
	}

	btnPass := markup.Data("⏭️ Pass", "battle_pass:"+battleID)
	btnLeader := markup.Data("💎 Leader Ability", "battle_leader:"+battleID)
	btnViewBoard := markup.Data("👁️ View Board", "battle_view_board:"+battleID)
	btnViewHand := markup.Data("🤲 View Hand", "battle_view_hand:"+battleID)
	btnForfeit := markup.Data("🏳️ Forfeit", "battle_forfeit:"+battleID)

	markup.Row(btnPass)
	markup.Row(btnLeader, btnViewBoard)
	markup.Row(btnViewHand, btnForfeit)

	return markup
}

func (k *KeyboardBuilder) CardSelection(hand []models.CardInstance, battleID string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	for i, card := range hand {
		if i >= 8 {
			break
		}
		btn := markup.Data(
			"Card "+card.InstanceID[:8]+" (P:"+string(rune(card.Power))+")",
			"battle_select_card:"+battleID+":"+card.InstanceID,
		)
		markup.Row(btn)
	}

	btnBack := markup.Data("🔙 Back", "battle_back:"+battleID)
	markup.Row(btnBack)

	return markup
}

func (k *KeyboardBuilder) RowSelection(cardInstanceID, battleID string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	btnMelee := markup.Data("⚔️ Melee", "battle_row:"+battleID+":"+cardInstanceID+":melee")
	btnRanged := markup.Data("🏹 Ranged", "battle_row:"+battleID+":"+cardInstanceID+":ranged")
	btnSiege := markup.Data("🏰 Siege", "battle_row:"+battleID+":"+cardInstanceID+":siege")
	btnBack := markup.Data("🔙 Back", "battle_select_card:"+battleID)

	markup.Row(btnMelee, btnRanged, btnSiege)
	markup.Row(btnBack)

	return markup
}

func (k *KeyboardBuilder) DeckManagement(deck []string, collection []models.Card, maxDeckSize int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	btnAdd := markup.Data("➕ Add Card", "deck_add")
	btnRemove := markup.Data("➖ Remove Card", "deck_remove")
	btnSave := markup.Data("💾 Save Deck", "deck_save")
	btnBack := markup.Data("🔙 Back", "main_menu")

	markup.Row(btnAdd, btnRemove)
	markup.Row(btnSave, btnBack)

	return markup
}

func (k *KeyboardBuilder) ShopMenu() *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	row1 := markup.Row(
		markup.Data("🃏 Buy Cards", "shop_cards"),
		markup.Data("🎁 Buy Chests", "shop_chests"),
	)
	row2 := markup.Row(
		markup.Data("⛏️ Buy Axe", "shop_axe"),
		markup.Data("🔙 Back", "main_menu"),
	)
	markup.Inline(row1, row2)

	return markup
}

func (k *KeyboardBuilder) ShopCards(cards []models.Card, page, totalPages int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	if len(cards) == 0 {
		btnBack := markup.Data("🔙 Back", "shop_main")
		markup.Row(btnBack)
		return markup
	}

	// Current card (1-indexed page)
	currentCard := cards[page-1]

	// Buy button for current card
	btnBuy := markup.Data(
		fmt.Sprintf("💰 Buy %s (%d Gold)", currentCard.Name, currentCard.CostGold),
		"shop_buy_card:"+currentCard.ID,
	)
	markup.Row(btnBuy)

	// Navigation row: Previous, Next, Back
	navRow := []tele.Btn{}

	if page > 1 {
		navRow = append(navRow, markup.Data("⬅️ Prev", "shop_page:"+strconv.Itoa(page-1)))
	} else {
		// Disabled previous button
		navRow = append(navRow, markup.Data("⬅️ Prev", "shop_page_disabled"))
	}

	if page < totalPages {
		navRow = append(navRow, markup.Data("Next ➡️", "shop_page:"+strconv.Itoa(page+1)))
	} else {
		// Disabled next button
		navRow = append(navRow, markup.Data("Next ➡️", "shop_page_disabled"))
	}

	// Back button
	navRow = append(navRow, markup.Data("🔙 Back", "shop_main"))

	markup.Row(navRow...)

	return markup
}

func (k *KeyboardBuilder) MineMenu(unlocked bool, extractsLeft, axeDurability int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{ResizeKeyboard: true}

	if !unlocked {
		btnUnlock := markup.Data("💎 Unlock Mine (500 Gold)", "mine_unlock")
		markup.Row(btnUnlock)
	} else {
		btnExtract := markup.Data("⛏️ Extract Gold", "mine_extract")
		btnBuyAxe := markup.Data("🪓 Buy Axe (100 Gold)", "mine_buy_axe")
		markup.Row(btnExtract)
		markup.Row(btnBuyAxe)
	}

	btnBack := markup.Data("🔙 Back", "main_menu")
	markup.Row(btnBack)

	return markup
}

func (k *KeyboardBuilder) ChestMenu(chests []models.Chest, playerLevel models.PlayerLevel) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{ResizeKeyboard: true}

	for _, chest := range chests {
		if chest.LevelRequired <= int(playerLevel) {
			btn := markup.Data(
				chest.Name+" ("+chest.Cooldown.String()+")",
				"shop_buy_chest:"+chest.ID,
			)
			markup.Row(btn)
		}
	}

	btnBack := markup.Data("🔙 Back", "shop_main")
	markup.Row(btnBack)

	return markup
}

func (k *KeyboardBuilder) LeaderSelection(leaders []models.Card, currentLeaderID string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	for _, leader := range leaders {
		btnText := leader.Name
		if leader.ID == currentLeaderID {
			btnText = "✅ " + btnText
		}
		btn := markup.Data(btnText, "leader_select:"+leader.ID)
		markup.Row(btn)
	}

	btnBack := markup.Data("🔙 Back", "main_menu")
	markup.Row(btnBack)

	return markup
}

func (k *KeyboardBuilder) Confirm(action, data string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	btnYes := markup.Data("✅ Yes", "confirm:"+action+":"+data)
	btnNo := markup.Data("❌ No", "cancel:"+action)

	markup.Row(btnYes, btnNo)

	return markup
}

func (k *KeyboardBuilder) FriendBattleInvite(inviterID int64) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	btnAccept := markup.Data("✅ Accept Challenge", "friend_accept:"+strconv.FormatInt(inviterID, 10))
	btnDecline := markup.Data("❌ Decline", "friend_decline")

	markup.Row(btnAccept, btnDecline)

	return markup
}

func (k *KeyboardBuilder) Pagination(currentPage, totalPages int, prefix string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	navRow := []tele.Btn{}
	if currentPage > 1 {
		navRow = append(navRow, markup.Data("⬅️ Prev", prefix+":"+strconv.Itoa(currentPage-1)))
	}
	if currentPage < totalPages {
		navRow = append(navRow, markup.Data("Next ➡️", prefix+":"+strconv.Itoa(currentPage+1)))
	}
	if len(navRow) > 0 {
		markup.Row(navRow...)
	}

	return markup
}
