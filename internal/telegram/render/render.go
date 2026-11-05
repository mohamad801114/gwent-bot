package render

import (
	"fmt"
	"strings"
	"time"

	"gwent-bot/internal/config"
	"gwent-bot/pkg/models"
)

type Renderer struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Renderer {
	return &Renderer{cfg: cfg}
}

func (r *Renderer) WelcomeMessage(player *models.Player) string {
	return fmt.Sprintf(
		"🎮 <b>Welcome to Gwent Bot!</b>\n\n"+
			"Hello, %s! You are a <b>Level %d</b> %s player.\n"+
			"💰 Gold: %d | 💎 Scrap: %d | ✨ Powder: %d\n\n"+
			"Choose an action from the menu below:",
		player.FirstName,
		player.Level,
		player.NationEmoji(),
		player.Gold,
		player.Scrap,
		player.MeteoritePowder,
	)
}

func (r *Renderer) HelpMessage() string {
	return fmt.Sprintf(`<b>🃏 Gwent Bot Commands</b>

<b>Main Commands:</b>
/start - Register and get starter deck
/deck - View and manage your deck
/collection - View your card collection
/battle - Start a battle (PvE or PvP)
/shop - Buy cards, chests, and tools
/mine - Extract gold from your mine
/chest - Open reward chests
/leader - Select your leader ability
/profile - View your stats and progress
/help - Show this help message

<b>Game Mechanics:</b>
• <b>Battles:</b> 3 rounds, best of 3. Play cards to rows (Melee/Ranged/Siege). Highest total power wins round.
• <b>Leader:</b> Once per battle, use your leader's ability.
• <b>Pass:</b> When you pass, you're done for the round. Opponent plays alone until they pass.
• <b>Win:</b> Win 2 rounds to win the match.

<b>Progression:</b>
• <b>Daily Coin:</b> Say "<code>%s</code>" once per day for free gold
• <b>Axe:</b> Say "<code>%s</code>" to get a mining axe (1hr cooldown)
• <b>Mine:</b> Extract gold daily with axe (durability 10)
• <b>Chests:</b> Open chests for cards and gold (level-based cooldown)
• <b>XP:</b> Earn XP from battles to level up

<b>Keywords:</b>
• <b>Deploy:</b> Trigger when played
• <b>Deathwish:</b> Trigger when destroyed
• <b>Order:</b> Manual ability (once per turn)
• <b>Zeal:</b> Order can be used immediately
• <b>Vitality:</b> Boost self by X
• <b>Armor:</b> Protects health from damage`,
		r.cfg.DailyCoinWord,
		r.cfg.AxeWord,
	)
}

func ternary(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

func (r *Renderer) ProfileMessage(player *models.Player) string {
	winRate := 0.0
	if player.GamesPlayed > 0 {
		winRate = float64(player.GamesWon) / float64(player.GamesPlayed) * 100
	}

	dailyCoinStatus := "✅ Available"
	if player.DailyCoinClaimedAt != nil {
		nextClaim := player.DailyCoinClaimedAt.Add(r.cfg.DailyCoinCooldown)
		if time.Now().Before(nextClaim) {
			dailyCoinStatus = fmt.Sprintf("⏳ %s", nextClaim.Format("15:04"))
		}
	}

	axeStatus := "✅ Available"
	if player.AxeClaimedAt != nil {
		nextAxe := player.AxeClaimedAt.Add(r.cfg.AxeCooldown)
		if time.Now().Before(nextAxe) {
			axeStatus = fmt.Sprintf("⏳ %s", nextAxe.Format("15:04"))
		}
	}

	mineStatus := "🔒 Locked"
	if player.MineUnlocked {
		mineStatus = fmt.Sprintf("⛏️ %d/%d extracts", player.MineExtractsToday, r.cfg.MineMaxExtracts)
	}

	return fmt.Sprintf(
		"<b>👤 %s's Profile</b>\n\n"+
			"<b>Level:</b> %d (XP: %d)\n"+
			"<b>Nation:</b> %s\n\n"+
			"<b>💰 Resources:</b>\n"+
			"Gold: %d | Scrap: %d | Powder: %d\n\n"+
			"<b>⚔️ Battle Stats:</b>\n"+
			"Games: %d | Wins: %d | Losses: %d | Win Rate: %.1f%%\n"+
			"Highest Rank: %d\n\n"+
			"<b>🎁 Daily Rewards:</b>\n"+
			"Daily Coin: %s\n"+
			"Axe: %s\n"+
			"Mine: %s\n\n"+
			"<b>🃏 Collection:</b> %d cards\n"+
			"<b>🃏 Deck:</b> %d/%d cards",
		player.FirstName,
		player.Level,
		player.Experience,
		player.NationEmoji(),
		player.Gold,
		player.Scrap,
		player.MeteoritePowder,
		player.GamesPlayed,
		player.GamesWon,
		player.GamesLost,
		winRate,
		player.HighestRank,
		dailyCoinStatus,
		axeStatus,
		mineStatus,
		len(player.Collection),
		len(player.Deck),
		25,
	)
}

func (r *Renderer) DeckMessage(player *models.Player, cards []models.Card) string {
	if len(cards) == 0 {
		return "🃏 <b>Your Deck</b>\n\nYour deck is empty! Add cards from your collection."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🃏 <b>Your Deck</b> (%d/25 cards)\n\n", len(cards)))

	provisionTotal := 0
	for i, card := range cards {
		provisionTotal += card.Provision
		sb.WriteString(fmt.Sprintf("%d. %s %s [%s] P:%d\n",
			i+1,
			card.Rarity.Emoji(),
			card.Name,
			card.Nation.Emoji(),
			card.Provision,
		))
	}

	sb.WriteString(fmt.Sprintf("\n<b>Total Provisions:</b> %d/165", provisionTotal))

	if provisionTotal > 165 {
		sb.WriteString(" ⚠️ <b>OVER LIMIT!</b>")
	}

	return sb.String()
}

func (r *Renderer) CollectionMessage(player *models.Player, cards []models.Card, page, totalPages int) string {
	if len(cards) == 0 {
		return "📚 <b>Your Collection</b>\n\nEmpty! Win battles, open chests, or buy from shop."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📚 <b>Collection</b> (Page %d/%d)\n\n", page, totalPages))

	byNation := make(map[models.Nation][]models.Card)
	for _, card := range cards {
		byNation[card.Nation] = append(byNation[card.Nation], card)
	}

	nationOrder := []models.Nation{
		models.NationNorthernRealms,
		models.NationNilfgaard,
		models.NationScoiatael,
		models.NationMonsters,
		models.NationSyndicate,
		models.NationNeutral,
	}

	for _, nation := range nationOrder {
		nationCards := byNation[nation]
		if len(nationCards) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("<b>%s %s</b>\n", nation.Emoji(), nation))
		for _, card := range nationCards {
			sb.WriteString(fmt.Sprintf("  %s %s P:%d\n",
				card.Rarity.Emoji(),
				card.Name,
				card.Provision,
			))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (r *Renderer) BattleStateMessage(b *models.Battle, playerID int64, playerName string, opponentName string) string {
	isPlayer1 := b.Player1ID == playerID

	var myBoard, oppBoard *models.BoardState
	var myHand, oppHand []models.CardInstance
	var myLeaderUsed, oppLeaderUsed bool
	var myPassed, oppPassed bool
	var myTurn bool

	if isPlayer1 {
		myBoard = &b.Player1Board
		oppBoard = &b.Player2Board
		myHand = b.Player1Hand
		oppHand = b.Player2Hand
		myLeaderUsed = b.Player1LeaderUsed
		oppLeaderUsed = b.Player2LeaderUsed
		myPassed = b.Player1Passed
		oppPassed = b.Player2Passed
		myTurn = b.CurrentTurn == b.Player1ID
	} else {
		myBoard = &b.Player2Board
		oppBoard = &b.Player1Board
		myHand = b.Player2Hand
		oppHand = b.Player1Hand
		myLeaderUsed = b.Player2LeaderUsed
		oppLeaderUsed = b.Player1LeaderUsed
		myPassed = b.Player2Passed
		oppPassed = b.Player1Passed
		myTurn = b.CurrentTurn == b.Player2ID
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("⚔️ <b>Battle</b> | Round %d/3 | %s\n\n",
		b.CurrentRound,
		b.PhaseDisplay(),
	))

	sb.WriteString(fmt.Sprintf("👤 <b>%s</b> (Power: %d)%s\n",
		opponentName,
		oppBoard.TotalPower,
		ternary(oppPassed, " ⏭️ PASSED", ""),
	))
	sb.WriteString(r.renderBoard(oppBoard, false))
	sb.WriteString(fmt.Sprintf("Hand: %d cards%s\n\n",
		len(oppHand),
		ternary(oppLeaderUsed, " | Leader: ✅ Used", " | Leader: ⭐ Ready"),
	))

	if b.Weather.Active != "clear" {
		sb.WriteString(fmt.Sprintf("🌤️ <b>Weather:</b> %s (%d rounds)\n\n",
			b.Weather.Display(),
			b.Weather.Duration,
		))
	}

	sb.WriteString(fmt.Sprintf("👤 <b>%s</b> (Power: %d)%s\n",
		playerName,
		myBoard.TotalPower,
		ternary(myPassed, " ⏭️ PASSED", ""),
	))
	sb.WriteString(r.renderBoard(myBoard, true))
	sb.WriteString(fmt.Sprintf("Hand: %d cards%s\n",
		len(myHand),
		ternary(myLeaderUsed, " | Leader: ✅ Used", " | Leader: ⭐ Ready"),
	))

	if myTurn {
		sb.WriteString("\n🎯 <b>YOUR TURN</b> - Play a card or Pass!")
	} else {
		sb.WriteString("\n⏳ Opponent's turn...")
	}

	return sb.String()
}

func (r *Renderer) renderBoard(board *models.BoardState, showDetails bool) string {
	var sb strings.Builder

	rows := []struct {
		name  string
		emoji string
		cards models.CardRowDB
	}{
		{"Siege", "🏰", board.Siege},
		{"Ranged", "🏹", board.Ranged},
		{"Melee", "⚔️", board.Melee},
	}

	for _, row := range rows {
		if len(row.cards) == 0 {
			sb.WriteString(fmt.Sprintf("%s %s: —\n", row.emoji, row.name))
			continue
		}
		sb.WriteString(fmt.Sprintf("%s <b>%s:</b> ", row.emoji, row.name))
		for i, card := range row.cards {
			if i > 0 {
				sb.WriteString(" | ")
			}
			powerStr := fmt.Sprintf("%d", card.Power)
			if card.Boosted > 0 {
				powerStr = fmt.Sprintf("%d🟢", card.Power)
			}
			if card.Damaged > 0 {
				powerStr = fmt.Sprintf("%d🔴", card.Power)
			}
			if card.Armor > 0 {
				powerStr += fmt.Sprintf("🛡️%d", card.Armor)
			}
			for _, status := range card.Statuses {
				powerStr += " " + status.Emoji()
			}
			sb.WriteString(fmt.Sprintf("[%s]", powerStr))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (r *Renderer) BattleResultMessage(b *models.Battle, playerID int64, won bool, rewards string) string {
	result := "🏆 <b>VICTORY!</b>"
	if !won {
		result = "💀 <b>DEFEAT</b>"
	}

	return fmt.Sprintf(
		"%s\n\n"+
			"Rounds played: %d\n"+
			"Final score: %d - %d\n\n"+
			"%s",
		result,
		b.CurrentRound,
		b.Player1Board.TotalPower,
		b.Player2Board.TotalPower,
		rewards,
	)
}

func (r *Renderer) ShopMessage(player *models.Player) string {
	return fmt.Sprintf(
		"🏪 <b>Shop</b>\n\n"+
			"💰 Gold: %d | 💎 Scrap: %d\n\n"+
			"What would you like to buy?",
		player.Gold,
		player.Scrap,
	)
}

func (r *Renderer) ShopCardsMessage(cards []models.Card, page, totalPages int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🃏 <b>Buy Cards</b> (Page %d/%d)\n\n", page, totalPages))

	for _, card := range cards {
		cost := ""
		if card.CostGold > 0 {
			cost += fmt.Sprintf(" 💰%d", card.CostGold)
		}
		if card.CostScrap > 0 {
			cost += fmt.Sprintf(" 💎%d", card.CostScrap)
		}
		sb.WriteString(fmt.Sprintf("%s %s [%s] P:%d%s\n",
			card.Rarity.Emoji(),
			card.Name,
			card.Nation.Emoji(),
			card.Provision,
			cost,
		))
	}

	return sb.String()
}

func (r *Renderer) MineMessage(player *models.Player, cfg *config.Config) string {
	if !player.MineUnlocked {
		return fmt.Sprintf(
			"⛏️ <b>Mine</b> 🔒\n\n"+
				"Your mine is locked! Unlock it for <b>%d Gold</b> to start extracting gold daily.\n\n"+
				"Once unlocked:\n"+
				"• Extract up to %d times/day\n"+
				"• %d Gold per extract\n"+
				"• Requires Axe (durability %d)\n"+
				"• Say '%s' to get an axe",
			cfg.MineGoldPerExtract*10,
			cfg.MineMaxExtracts,
			cfg.MineGoldPerExtract,
			cfg.AxeDurability,
			cfg.AxeWord,
		)
	}

	axeStatus := "❌ No axe"
	if player.AxeDurability > 0 {
		axeStatus = fmt.Sprintf("🪓 Durability: %d/%d", player.AxeDurability, cfg.AxeDurability)
	}

	extractsLeft := cfg.MineMaxExtracts - player.MineExtractsToday

	return fmt.Sprintf(
		"⛏️ <b>Mine</b> ✅ Unlocked\n\n"+
			"🪓 %s\n"+
			"⛏️ Extracts today: %d/%d (%d left)\n"+
			"💰 Gold per extract: %d\n\n"+
			"Say '%s' to get a new axe (1hr cooldown)",
		axeStatus,
		player.MineExtractsToday,
		cfg.MineMaxExtracts,
		extractsLeft,
		cfg.MineGoldPerExtract,
		cfg.AxeWord,
	)
}

func (r *Renderer) MineExtractMessage(goldEarned, extractsLeft int) string {
	return fmt.Sprintf(
		"⛏️ <b>Extraction Successful!</b>\n\n"+
			"💰 You extracted <b>%d Gold</b>!\n"+
			"⛏️ Extracts remaining today: %d",
		goldEarned,
		extractsLeft,
	)
}

func (r *Renderer) ChestMessage(chests []models.Chest, playerLevel models.PlayerLevel) string {
	var sb strings.Builder
	sb.WriteString("🎁 <b>Reward Chests</b>\n\n")

	for _, chest := range chests {
		status := "🔒"
		if chest.LevelRequired <= playerLevel {
			status = "🔓"
		}
		sb.WriteString(fmt.Sprintf("%s <b>%s</b> (Level %d+)\n", status, chest.Name, chest.LevelRequired))
		sb.WriteString(fmt.Sprintf("   %s\n", chest.Description))
		sb.WriteString(fmt.Sprintf("   💰 %d Gold | 🎁 %d cards | ⏳ %s\n\n",
			chest.Gold,
			len(chest.Cards),
			chest.Cooldown.String(),
		))
	}

	return sb.String()
}

func (r *Renderer) ChestOpenMessage(chestName string, cards []models.Card, gold int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🎁 <b>%s Opened!</b>\n\n", chestName))

	if gold > 0 {
		sb.WriteString(fmt.Sprintf("💰 Gold: +%d\n", gold))
	}

	if len(cards) > 0 {
		sb.WriteString("🃏 Cards:\n")
		for _, card := range cards {
			sb.WriteString(fmt.Sprintf("  %s %s [%s] P:%d\n",
				card.Rarity.Emoji(),
				card.Name,
				card.Nation.Emoji(),
				card.Provision,
			))
		}
	}

	return sb.String()
}

func (r *Renderer) LeaderMessage(leaders []models.Card, currentLeaderID string) string {
	var sb strings.Builder
	sb.WriteString("👑 <b>Select Leader Ability</b>\n\n")

	for _, leader := range leaders {
		marker := "  "
		if leader.ID == currentLeaderID {
			marker = "✅ "
		}
		sb.WriteString(fmt.Sprintf("%s%s %s\n", marker, leader.Rarity.Emoji(), leader.Name))
		sb.WriteString(fmt.Sprintf("   %s\n\n", leader.Description))
	}

	return sb.String()
}

func (r *Renderer) ErrorMessage(err string) string {
	return fmt.Sprintf("❌ <b>Error:</b> %s", err)
}

func (r *Renderer) SuccessMessage(msg string) string {
	return fmt.Sprintf("✅ %s", msg)
}

func (r *Renderer) InfoMessage(msg string) string {
	return fmt.Sprintf("ℹ️ %s", msg)
}

func (r *Renderer) NotRegisteredMessage() string {
	return "❌ You're not registered! Use /start to begin your journey."
}
