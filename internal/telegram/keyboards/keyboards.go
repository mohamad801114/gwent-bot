package keyboards

import (
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

	btnBattle := markup.Data("âš”ï¸ Battle", "battle_menu")
	btnDeck := markup.Data("ðŸƒ My Deck", "show_deck")
	btnCollection := markup.Data("ðŸ“š Collection", "show_collection")
	btnShop := markup.Data("ðŸª Shop", "shop_menu")
	btnMine := markup.Data("â›ï¸ Mine", "mine_menu")
	btnChest := markup.Data("ðŸŽ Chests", "chest_menu")
	btnProfile := markup.Data("ðŸ‘¤ Profile", "show_profile")
	btnLeader := markup.Data("ðŸ‘‘ Leader", "leader_menu")
	btnHelp := markup.Data("â“ Help", "show_help")

	markup.Row(btnBattle)
	markup.Row(btnDeck, btnCollection)
	markup.Row(btnShop, btnMine)
	markup.Row(btnChest, btnProfile)
	markup.Row(btnLeader, btnHelp)

	return markup
}

func (k *KeyboardBuilder) BattleMenu() *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{ResizeKeyboard: true}

	btnPvE := markup.Data("ðŸ¤– vs Bot (Easy)", "battle_pve_easy")
	btnPvEMedium := markup.Data("ðŸ¤– vs Bot (Medium)", "battle_pve_medium")
	btnPvEHard := markup.Data("ðŸ¤– vs Bot (Hard)", "battle_pve_hard")
	btnPvP := markup.Data("ðŸ‘¥ vs Player", "battle_pvp")
	btnFriend := markup.Data("ðŸ¤ vs Friend", "battle_friend")
	btnBack := markup.Data("ðŸ”™ Back", "main_menu")

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
		btnWait := markup.Data("â³ Opponent's turn...", "battle_wait")
		markup.Row(btnWait)
		return markup
	}

	if canPlayCard && handSize > 0 {
		btnPlay := markup.Data("ðŸƒ Play Card", "battle_play_card:"+battleID)
		markup.Row(btnPlay)
	}

	btnPass := markup.Data("â­ï¸ Pass", "battle_pass:"+battleID)
	btnLeader := markup.Data("ðŸ‘‘ Leader Ability", "battle_leader:"+battleID)
	btnViewBoard := markup.Data("ðŸ‘ï¸ View Board", "battle_view_board:"+battleID)
	btnViewHand := markup.Data("ðŸ¤² View Hand", "battle_view_hand:"+battleID)
	btnForfeit := markup.Data("ðŸ³ï¸ Forfeit", "battle_forfeit:"+battleID)

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

	btnBack := markup.Data("ðŸ”™ Back", "battle_back:"+battleID)
	markup.Row(btnBack)

	return markup
}

func (k *KeyboardBuilder) RowSelection(cardInstanceID, battleID string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	btnMelee := markup.Data("âš”ï¸ Melee", "battle_row:"+battleID+":"+cardInstanceID+":melee")
	btnRanged := markup.Data("ðŸ¹ Ranged", "battle_row:"+battleID+":"+cardInstanceID+":ranged")
	btnSiege := markup.Data("ðŸ° Siege", "battle_row:"+battleID+":"+cardInstanceID+":siege")
	btnBack := markup.Data("ðŸ”™ Back", "battle_select_card:"+battleID)

	markup.Row(btnMelee, btnRanged, btnSiege)
	markup.Row(btnBack)

	return markup
}

func (k *KeyboardBuilder) DeckManagement(deck []string, collection []models.Card, maxDeckSize int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	btnAdd := markup.Data("âž• Add Card", "deck_add")
	btnRemove := markup.Data("âž– Remove Card", "deck_remove")
	btnSave := markup.Data("ðŸ’¾ Save Deck", "deck_save")
	btnBack := markup.Data("ðŸ”™ Back", "main_menu")

	markup.Row(btnAdd, btnRemove)
	markup.Row(btnSave, btnBack)

	return markup
}

func (k *KeyboardBuilder) ShopMenu() *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{ResizeKeyboard: true}

	btnCards := markup.Data("ðŸƒ Buy Cards", "shop_cards")
	btnChests := markup.Data("ðŸŽ Buy Chests", "shop_chests")
	btnAxe := markup.Data("â›ï¸ Buy Axe", "shop_axe")
	btnBack := markup.Data("ðŸ”™ Back", "main_menu")

	markup.Row(btnCards, btnChests)
	markup.Row(btnAxe, btnBack)

	return markup
}

func (k *KeyboardBuilder) ShopCards(cards []models.Card, page, totalPages int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	for i, card := range cards {
		if i >= 5 {
			break
		}
		btn := markup.Data(
			card.Rarity.Emoji()+" "+card.Name+" ["+card.Nation.Emoji()+"] P:"+string(rune(card.Provision)),
			"shop_buy_card:"+card.ID,
		)
		markup.Row(btn)
	}

	navRow := []tele.Btn{}
	if page > 1 {
		navRow = append(navRow, markup.Data("â¬…ï¸ Prev", "shop_page:"+strconv.Itoa(page-1)))
	}
	if page < totalPages {
		navRow = append(navRow, markup.Data("Next âž¡ï¸", "shop_page:"+strconv.Itoa(page+1)))
	}
	if len(navRow) > 0 {
		markup.Row(navRow...)
	}

	btnBack := markup.Data("ðŸ”™ Back", "shop_menu")
	markup.Row(btnBack)

	return markup
}

func (k *KeyboardBuilder) MineMenu(unlocked bool, extractsLeft, axeDurability int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{ResizeKeyboard: true}

	if !unlocked {
		btnUnlock := markup.Data("ðŸ”“ Unlock Mine (500 Gold)", "mine_unlock")
		markup.Row(btnUnlock)
	} else {
		btnExtract := markup.Data("â›ï¸ Extract Gold", "mine_extract")
		btnBuyAxe := markup.Data("ðŸª“ Buy Axe (100 Gold)", "mine_buy_axe")
		markup.Row(btnExtract)
		markup.Row(btnBuyAxe)
	}

	btnBack := markup.Data("ðŸ”™ Back", "main_menu")
	markup.Row(btnBack)

	return markup
}

func (k *KeyboardBuilder) ChestMenu(chests []models.Chest, playerLevel models.PlayerLevel) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{ResizeKeyboard: true}

	for _, chest := range chests {
		if chest.LevelRequired <= playerLevel {
			btn := markup.Data(
				chest.Name+" ("+chest.Cooldown.String()+")",
				"chest_open:"+chest.ID,
			)
			markup.Row(btn)
		}
	}

	btnBack := markup.Data("ðŸ”™ Back", "main_menu")
	markup.Row(btnBack)

	return markup
}

func (k *KeyboardBuilder) LeaderSelection(leaders []models.Card, currentLeaderID string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	for _, leader := range leaders {
		btnText := leader.Name
		if leader.ID == currentLeaderID {
			btnText = "âœ… " + btnText
		}
		btn := markup.Data(btnText, "leader_select:"+leader.ID)
		markup.Row(btn)
	}

	btnBack := markup.Data("ðŸ”™ Back", "main_menu")
	markup.Row(btnBack)

	return markup
}

func (k *KeyboardBuilder) Confirm(action, data string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	btnYes := markup.Data("âœ… Yes", "confirm:"+action+":"+data)
	btnNo := markup.Data("âŒ No", "cancel:"+action)

	markup.Row(btnYes, btnNo)

	return markup
}

func (k *KeyboardBuilder) Pagination(currentPage, totalPages int, prefix string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	navRow := []tele.Btn{}
	if currentPage > 1 {
		navRow = append(navRow, markup.Data("â¬…ï¸ Prev", prefix+":"+strconv.Itoa(currentPage-1)))
	}
	if currentPage < totalPages {
		navRow = append(navRow, markup.Data("Next âž¡ï¸", prefix+":"+strconv.Itoa(currentPage+1)))
	}
	if len(navRow) > 0 {
		markup.Row(navRow...)
	}

	return markup
}
