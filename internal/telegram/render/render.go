package render

import (
	"fmt"
	"os"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"

	"gwent-bot/internal/config"
	"gwent-bot/pkg/models"
)

type Renderer struct {
	cfg *config.Config


func New(cfg *config.Config) *Renderer {
	return &Renderer{cfg: cfg


func (r *Renderer) WelcomeMessage(player *models.Player) string {
	return fmt.Sprintf(
		"ðŸŽ® <b>Welcome to Gwent Bot!</b>\n\n"+
			"Hello, %s! You are a <b>Level %d</b> %s player.\n"+
			"ðŸ’° Gold: %d | ðŸ’Ž Scrap: %d | âœ¨ Powder: %d\n\n"+
			"Choose an action from the menu below:",
		player.FirstName,
		player.Level,
		player.NationEmoji(),
		player.Gold,
		player.Scrap,
		player.MeteoritePowder,
	)


func (r *Renderer) HelpMessage() string {
	return fmt.Sprintf(`<b>ðŸƒ Gwent Bot Commands</b>

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
â€¢ <b>Battles:</b> 3 rounds, best of 3. Play cards to rows (Melee/Ranged/Siege). Highest total power wins round.
â€¢ <b>Leader:</b> Once per battle, use your leader's ability.
â€¢ <b>Pass:</b> When you pass, you're done for the round. Opponent plays alone until they pass.
â€¢ <b>Win:</b> Win 2 rounds to win the match.

<b>Progression:</b>
â€¢ <b>Daily Coin:</b> Say "<code>%s</code>" once per day for free gold
â€¢ <b>Axe:</b> Say "<code>%s</code>" to get a mining axe (1hr cooldown)
â€¢ <b>Mine:</b> Extract gold daily with axe (durability 10)
â€¢ <b>Chests:</b> Open chests for cards and gold (level-based cooldown)
â€¢ <b>XP:</b> Earn XP from battles to level up

<b>Keywords:</b>
â€¢ <b>Deploy:</b> Trigger when played
â€¢ <b>Deathwish:</b> Trigger when destroyed
â€¢ <b>Order:</b> Manual ability (once per turn)
â€¢ <b>Zeal:</b> Order can be used immediately
â€¢ <b>Vitality:</b> Boost self by X
`,
		r.cfg.DailyCoinWord,
		r.cfg.AxeWord,
	)


func ternary(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	
	return falseVal


func (r *Renderer) ProfileMessage(player *models.Player) string {
	winRate := 0.0
	if player.GamesPlayed > 0 {
		winRate = float64(player.GamesWon) / float64(player.GamesPlayed) * 100
	

	dailyCoinStatus := "âœ… Available"
	if player.DailyCoinClaimedAt != nil {
		nextClaim := player.DailyCoinClaimedAt.Add(r.cfg.DailyCoinCooldown)
		if time.Now().Before(nextClaim) {
			dailyCoinStatus = fmt.Sprintf("â³ %s", nextClaim.Format("15:04"))
		
	

	axeStatus := "âœ… Available"
	if player.AxeClaimedAt != nil {
		nextAxe := player.AxeClaimedAt.Add(r.cfg.AxeCooldown)
		if time.Now().Before(nextAxe) {
			axeStatus = fmt.Sprintf("â³ %s", nextAxe.Format("15:04"))
		
	

	mineStatus := "ðŸ”’ Locked"
	if player.MineUnlocked {
		mineStatus = fmt.Sprintf("â›ï¸ %d/%d extracts", player.MineExtractsToday, r.cfg.MineMaxExtracts)
	

	return fmt.Sprintf(
		"<b>ðŸ‘¤ %s's Profile</b>\n\n"+
			"<b>Level:</b> %d (XP: %d)\n"+
			"<b>Nation:</b> %s\n\n"+
			"<b>ðŸ’° Resources:</b>\n"+
			"Gold: %d | Scrap: %d | Powder: %d\n\n"+
			"<b>âš”ï¸ Battle Stats:</b>\n"+
			"Games: %d | Wins: %d | Losses: %d | Win Rate: %.1f%%\n"+
			"Highest Rank: %d\n\n"+
			"<b>ðŸŽ Daily Rewards:</b>\n"+
			"Daily Coin: %s\n"+
			"Axe: %s\n"+
			"Mine: %s\n\n"+
			"<b>ðŸƒ Collection:</b> %d cards\n"+
			"<b>ðŸƒ Deck:</b> %d/%d cards",
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


func (r *Renderer) DeckMessage(player *models.Player, cards []models.Card) string {
	if len(cards) == 0 {
		return "ðŸƒ <b>Your Deck</b>\n\nYour deck is empty! Add cards from your collection."
	

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ðŸƒ <b>Your Deck</b> (%d/25 cards)\n\n", len(cards)))

	for i, card := range cards {
		
		sb.WriteString(fmt.Sprintf("%d. %s %s [%s] P:%d\n",
			i+1,
			card.Rarity.Emoji(),
			card.Name,
			card.Nation.Emoji(), card.Power))
	

	

	
		
	

	return sb.String()


func (r *Renderer) CollectionMessage(player *models.Player, cards []models.Card, page, totalPages int) string {
	if len(cards) == 0 {
		return "ðŸ“š <b>Your Collection</b>\n\nEmpty! Win battles, open chests, or buy from shop."
	

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ðŸ“š <b>Collection</b> (Page %d/%d)\n\n", page, totalPages))

	byNation := make(map[models.Nation][]models.Card)
	for _, card := range cards {
		byNation[card.Nation] = append(byNation[card.Nation], card)
	

	nationOrder := []models.Nation{
		models.NationNorthernRealms,
		models.NationNilfgaard,
		models.NationScoiatael,
		models.NationMonsters,
		models.NationSyndicate,
		models.NationNeutral,
	

	for _, nation := range nationOrder {
		nationCards := byNation[nation]
		if len(nationCards) == 0 {
			continue
		
		sb.WriteString(fmt.Sprintf("<b>%s %s</b>\n", nation.Emoji(), nation))
		for _, card := range nationCards {
			sb.WriteString(fmt.Sprintf("  %s %s P:%d\n",
				card.Rarity.Emoji(),
				card.Name,
				
			))
		
		sb.WriteString("\n")
	

	return sb.String()


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
	 else {
		myBoard = &b.Player2Board
		oppBoard = &b.Player1Board
		myHand = b.Player2Hand
		oppHand = b.Player1Hand
		myLeaderUsed = b.Player2LeaderUsed
		oppLeaderUsed = b.Player1LeaderUsed
		myPassed = b.Player2Passed
		oppPassed = b.Player1Passed
		myTurn = b.CurrentTurn == b.Player2ID
	

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("âš”ï¸ <b>Battle</b> | Round %d/3 | %s\n\n",
		b.CurrentRound,
		b.PhaseDisplay(),
	))

	sb.WriteString(fmt.Sprintf("ðŸ‘¤ <b>%s</b> (Power: %d)%s\n",
		opponentName,
		oppBoard.TotalPower,
		ternary(oppPassed, " â­ï¸ PASSED", ""),
	))
	sb.WriteString(r.renderBoard(oppBoard, false))
	sb.WriteString(fmt.Sprintf("Hand: %d cards%s\n\n",
		len(oppHand),
		ternary(oppLeaderUsed, " | Leader: âœ… Used", " | Leader: â­ Ready"),
	))

	if b.Weather.Active != "clear" {
		sb.WriteString(fmt.Sprintf("ðŸŒ¤ï¸ <b>Weather:</b> %s (%d rounds)\n\n",
			b.Weather.Display(),
			b.Weather.Duration,
		))
	

	sb.WriteString(fmt.Sprintf("ðŸ‘¤ <b>%s</b> (Power: %d)%s\n",
		playerName,
		myBoard.TotalPower,
		ternary(myPassed, " â­ï¸ PASSED", ""),
	))
	sb.WriteString(r.renderBoard(myBoard, true))
	sb.WriteString(fmt.Sprintf("Hand: %d cards%s\n",
		len(myHand),
		ternary(myLeaderUsed, " | Leader: âœ… Used", " | Leader: â­ Ready"),
	))

	if myTurn {
		sb.WriteString("\nðŸŽ¯ <b>YOUR TURN</b> - Play a card or Pass!")
	 else {
		sb.WriteString("\nâ³ Opponent's turn...")
	

	return sb.String()


func (r *Renderer) renderBoard(board *models.BoardState, showDetails bool) string {
	var sb strings.Builder

	rows := []struct {
		name  string
		emoji string
		cards models.CardRowDB
	{
		{"Siege", "ðŸ°", board.Siege,
		{"Ranged", "ðŸ¹", board.Ranged,
		{"Melee", "âš”ï¸", board.Melee,
	

	for _, row := range rows {
		if len(row.cards) == 0 {
			sb.WriteString(fmt.Sprintf("%s %s: â€”\n", row.emoji, row.name))
			continue
		
		sb.WriteString(fmt.Sprintf("%s <b>%s:</b> ", row.emoji, row.name))
		for i, card := range row.cards {
			if i > 0 {
				sb.WriteString(" | ")
			
			powerStr := fmt.Sprintf("%d", card.Power)
			// Show base power if different (damaged/boosted)
			if card.Power != card.BasePower {
				if card.Power > card.BasePower {
					powerStr = fmt.Sprintf("%dðŸŸ¢(%d)", card.Power, card.BasePower)
				 else {
					powerStr = fmt.Sprintf("%dðŸ”´(%d)", card.Power, card.BasePower)
				
			
			for _, status := range card.Statuses {
				powerStr += " " + status.Emoji()
			
			sb.WriteString(fmt.Sprintf("[%s]", powerStr))
		
		sb.WriteString("\n")
	

	return sb.String()


func (r *Renderer) BattleResultMessage(b *models.Battle, playerID int64, won bool, rewards string) string {
	result := "ðŸ† <b>VICTORY!</b>"
	if !won {
		result = "ðŸ’€ <b>DEFEAT</b>"
	

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


func (r *Renderer) ShopMessage(player *models.Player) string {
	return fmt.Sprintf(
		"ðŸª <b>Shop</b>\n\n"+
			"ðŸ’° Gold: %d | ðŸ’Ž Scrap: %d\n\n"+
			"What would you like to buy?",
		player.Gold,
		player.Scrap,
	)


func (r *Renderer) ShopCardsMessage(cards []models.Card, page, totalPages int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ðŸƒ <b>Buy Cards</b> (Page %d/%d)\n\n", page, totalPages))

	for _, card := range cards {
		cost := ""
		if card.CostGold > 0 {
			cost += fmt.Sprintf(" ðŸ’°%d", card.CostGold)
		
		if card.CostScrap > 0 {
			cost += fmt.Sprintf(" ðŸ’Ž%d", card.CostScrap)
		
		sb.WriteString(fmt.Sprintf("%s %s [%s] P:%d%s\n",
			card.Rarity.Emoji(),
			card.Name,
			card.Nation.Emoji(),
			
			cost,
		))
	

	return sb.String()


func (r *Renderer) MineMessage(player *models.Player, cfg *config.Config) string {
	if !player.MineUnlocked {
		return fmt.Sprintf(
			"â›ï¸ <b>Mine</b> ðŸ”’\n\n"+
				"Your mine is locked! Unlock it for <b>%d Gold</b> to start extracting gold daily.\n\n"+
				"Once unlocked:\n"+
				"â€¢ Extract up to %d times/day\n"+
				"â€¢ %d Gold per extract\n"+
				"â€¢ Requires Axe (durability %d)\n"+
				"â€¢ Say '%s' to get an axe",
			cfg.MineGoldPerExtract*10,
			cfg.MineMaxExtracts,
			cfg.MineGoldPerExtract,
			cfg.AxeDurability,
			cfg.AxeWord,
		)
	

	axeStatus := "âŒ No axe"
	if player.AxeDurability > 0 {
		axeStatus = fmt.Sprintf("ðŸª“ Durability: %d/%d", player.AxeDurability, cfg.AxeDurability)
	

	extractsLeft := cfg.MineMaxExtracts - player.MineExtractsToday

	return fmt.Sprintf(
		"â›ï¸ <b>Mine</b> âœ… Unlocked\n\n"+
			"ðŸª“ %s\n"+
			"â›ï¸ Extracts today: %d/%d (%d left)\n"+
			"ðŸ’° Gold per extract: %d\n\n"+
			"Say '%s' to get a new axe (1hr cooldown)",
		axeStatus,
		player.MineExtractsToday,
		cfg.MineMaxExtracts,
		extractsLeft,
		cfg.MineGoldPerExtract,
		cfg.AxeWord,
	)


func (r *Renderer) MineExtractMessage(goldEarned, extractsLeft int) string {
	return fmt.Sprintf(
		"â›ï¸ <b>Extraction Successful!</b>\n\n"+
			"ðŸ’° You extracted <b>%d Gold</b>!\n"+
			"â›ï¸ Extracts remaining today: %d",
		goldEarned,
		extractsLeft,
	)


func (r *Renderer) ChestMessage(chests []models.Chest, playerLevel models.PlayerLevel) string {
	var sb strings.Builder
	sb.WriteString("ðŸŽ <b>Reward Chests</b>\n\n")

	for _, chest := range chests {
		status := "ðŸ”’"
		if chest.LevelRequired <= int(playerLevel) {
			status = "ðŸ”“"
		
		sb.WriteString(fmt.Sprintf("%s <b>%s</b> (Level %d+)\n", status, chest.Name, chest.LevelRequired))
		sb.WriteString(fmt.Sprintf("   %s\n", chest.Description))
		sb.WriteString(fmt.Sprintf("   ðŸ’° %d Gold | ðŸŽ %d cards | â³ %s\n\n",
			chest.Gold,
			len(chest.Cards),
			chest.Cooldown.String(),
		))
	

	return sb.String()


func (r *Renderer) ChestOpenMessage(chestName string, cards []models.Card, gold int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ðŸŽ <b>%s Opened!</b>\n\n", chestName))

	if gold > 0 {
		sb.WriteString(fmt.Sprintf("ðŸ’° Gold: +%d\n", gold))
	

	if len(cards) > 0 {
		sb.WriteString("ðŸƒ Cards:\n")
		for _, card := range cards {
			sb.WriteString(fmt.Sprintf("  %s %s [%s] P:%d\n",
				card.Rarity.Emoji(),
				card.Name,
				card.Nation.Emoji(), card.Power))
		
	

	return sb.String()


func (r *Renderer) LeaderMessage(leaders []models.Card, currentLeaderID string) string {
	var sb strings.Builder
	sb.WriteString("ðŸ‘‘ <b>Select Leader Ability</b>\n\n")

	for _, leader := range leaders {
		marker := "  "
		if leader.ID == currentLeaderID {
			marker = "âœ… "
		
		sb.WriteString(fmt.Sprintf("%s%s %s\n", marker, leader.Rarity.Emoji(), leader.Name))
		sb.WriteString(fmt.Sprintf("   %s\n\n", leader.Description))
	

	return sb.String()


func (r *Renderer) ErrorMessage(err string) string {
	return fmt.Sprintf("âŒ <b>Error:</b> %s", err)


func (r *Renderer) SuccessMessage(msg string) string {
	return fmt.Sprintf("âœ… %s", msg)


func (r *Renderer) InfoMessage(msg string) string {
	return fmt.Sprintf("â„¹ï¸ %s", msg)


func (r *Renderer) NotRegisteredMessage() string {
	return "âŒ You're not registered! Use /start to begin your journey."


// GetCardPhoto returns a telebot Photo for the card's art
func (r *Renderer) GetCardPhoto(card *models.Card) (*tele.Photo, error) {
	if card == nil || card.ArtURL == "" {
		return nil, nil
	

	// Check if file exists
	if _, err := os.Stat(card.ArtURL); os.IsNotExist(err) {
		return nil, nil
	

	// Read the file and create a photo
	photo := &tele.Photo{
		File:    tele.FromDisk(card.ArtURL),
		Caption: r.CardCaption(card),
	

	return photo, nil


// CardCaption generates the caption for a card photo
func (r *Renderer) CardCaption(card *models.Card) string {
	cost := ""
	if card.CostGold > 0 {
		cost += fmt.Sprintf(" ðŸ’°%d", card.CostGold)
	
	if card.CostScrap > 0 {
		cost += fmt.Sprintf(" ðŸ’Ž%d", card.CostScrap)
	

	return fmt.Sprintf(
		"<b>%s %s</b> [%s] P:%d%s\n\n%s\n\n%s",
		card.Rarity.Emoji(),
		card.Name,
		card.Nation.Emoji(),
		
		cost,
		card.Description,
		card.FlavorText,
	)


// SendCardPhoto sends a card photo with caption to the chat
func (r *Renderer) SendCardPhoto(c tele.Context, card *models.Card) error {
	photo, err := r.GetCardPhoto(card)
	if err != nil || photo == nil {
		// Fallback to text message if no photo
		msg := r.CardCaption(card)
		return c.Send(msg, tele.ModeHTML)
	
	return c.Send(photo, tele.ModeHTML)


// BattleStateMessageWithPhoto returns battle state message and optionally a card photo
func (r *Renderer) BattleStateMessageWithPhoto(b *models.Battle, playerID int64, playerName string, opponentName string) (string, *tele.Photo) {
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
	 else {
		myBoard = &b.Player2Board
		oppBoard = &b.Player1Board
		myHand = b.Player2Hand
		oppHand = b.Player1Hand
		myLeaderUsed = b.Player2LeaderUsed
		oppLeaderUsed = b.Player1LeaderUsed
		myPassed = b.Player2Passed
		oppPassed = b.Player1Passed
		myTurn = b.CurrentTurn == b.Player2ID
	

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("âš”ï¸ <b>Battle</b> | Round %d/3 | %s\n\n",
		b.CurrentRound,
		b.PhaseDisplay(),
	))

	sb.WriteString(fmt.Sprintf("ðŸ‘¤ <b>%s</b> (Power: %d)%s\n",
		opponentName,
		oppBoard.TotalPower,
		ternary(oppPassed, " â­ï¸ PASSED", ""),
	))
	sb.WriteString(r.renderBoard(oppBoard, false))
	sb.WriteString(fmt.Sprintf("Hand: %d cards%s\n\n",
		len(oppHand),
		ternary(oppLeaderUsed, " | Leader: âœ… Used", " | Leader: â­ Ready"),
	))

	if b.Weather.Active != "clear" {
		sb.WriteString(fmt.Sprintf("ðŸŒ¤ï¸ <b>Weather:</b> %s (%d rounds)\n\n",
			b.Weather.Display(),
			b.Weather.Duration,
		))
	

	sb.WriteString(fmt.Sprintf("ðŸ‘¤ <b>%s</b> (Power: %d)%s\n",
		playerName,
		myBoard.TotalPower,
		ternary(myPassed, " â­ï¸ PASSED", ""),
	))
	sb.WriteString(r.renderBoard(myBoard, true))
	sb.WriteString(fmt.Sprintf("Hand: %d cards%s\n",
		len(myHand),
		ternary(myLeaderUsed, " | Leader: âœ… Used", " | Leader: â­ Ready"),
	))

	if myTurn {
		sb.WriteString("\nðŸŽ¯ <b>YOUR TURN</b> - Play a card or Pass!")
	 else {
		sb.WriteString("\nâ³ Opponent's turn...")
	

	// Return the message and optionally a photo of the last played card
	// This is just a placeholder - the actual card photo would be sent separately
	return sb.String(), nil


// GetCardPhotoByID returns a telebot Photo for a card by ID
func (r *Renderer) GetCardPhotoByID(cardID string) (*tele.Photo, error) {
	// This would need access to card service - kept for reference
	return nil, nil

