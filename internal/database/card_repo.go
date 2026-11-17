package database

import (
	"log"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"gwent-bot/pkg/models"
)

type CardRepo struct {
	db *DB
}

func NewCardRepo(db *DB) *CardRepo {
	return &CardRepo{db: db}
}

func (r *CardRepo) GetCard(ctx context.Context, cardID string) (*models.Card, error) {
	query := `SELECT id, name, description, faction, type, row, power, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap FROM cards WHERE id = $1`

	var c models.Card
	var abilitiesJSON []byte
	var row sql.NullString

	err := r.db.Pool.QueryRow(ctx, query, cardID).Scan(
		&c.ID, &c.Name, &c.Description, &c.Faction, &c.Type,
		&row, &c.Power, &c.IsHero,
		&abilitiesJSON, &c.ArtURL, &c.FlavorText, &c.CardGroup, &c.Rarity, &c.CostGold, &c.CostScrap,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if row.Valid {
		rowStr := row.String
		// Handle legacy "close" -> "melee" conversion
		if rowStr == "close" {
			rowStr = "melee"
		}
		c.Row = models.RowType(rowStr)
	}
	if len(abilitiesJSON) > 0 {
		json.Unmarshal(abilitiesJSON, &c.Abilities)
	}

	// Set Nation as alias for Faction
	c.Nation = c.Faction

	return &c, nil
}

func (r *CardRepo) GetCardsByNation(ctx context.Context, nation models.Nation) ([]*models.Card, error) {
	query := `SELECT id, name, description, faction, type, row, power, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap FROM cards WHERE faction = $1 ORDER BY rarity`

	rows, err := r.db.Pool.Query(ctx, query, nation)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		var c models.Card
		var abilitiesJSON []byte
		var row sql.NullString

		err := rows.Scan(
			&c.ID, &c.Name, &c.Description, &c.Faction, &c.Type,
			&row, &c.Power, &c.IsHero,
			&abilitiesJSON, &c.ArtURL, &c.FlavorText, &c.CardGroup, &c.Rarity, &c.CostGold, &c.CostScrap,
		)
		if err != nil {
			return nil, err
		}

		if row.Valid {
			c.Row = models.RowType(row.String)
		}
		if len(abilitiesJSON) > 0 {
			json.Unmarshal(abilitiesJSON, &c.Abilities)
		}
		c.Nation = c.Faction
		cards = append(cards, &c)
	}

	return cards, nil
}

func (r *CardRepo) GetCardsByFaction(ctx context.Context, faction models.Faction) ([]*models.Card, error) {
	return r.GetCardsByNation(ctx, models.Nation(faction))
}

func (r *CardRepo) GetCardsByRarity(ctx context.Context, rarity models.Rarity) ([]*models.Card, error) {
	query := `SELECT id, name, description, faction, type, row, power, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap FROM cards WHERE rarity = $1 ORDER BY faction`

	rows, err := r.db.Pool.Query(ctx, query, rarity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		var c models.Card
		var abilitiesJSON []byte
		var row sql.NullString

		err := rows.Scan(
			&c.ID, &c.Name, &c.Description, &c.Faction, &c.Type,
			&row, &c.Power, &c.IsHero,
			&abilitiesJSON, &c.ArtURL, &c.FlavorText, &c.CardGroup, &c.Rarity, &c.CostGold, &c.CostScrap,
		)
		if err != nil {
			return nil, err
		}

		if row.Valid {
			c.Row = models.RowType(row.String)
		}
		if len(abilitiesJSON) > 0 {
			json.Unmarshal(abilitiesJSON, &c.Abilities)
		}
		c.Nation = c.Faction
		cards = append(cards, &c)
	}

	return cards, nil
}

func (r *CardRepo) GetCardsByType(ctx context.Context, cardType models.CardType) ([]*models.Card, error) {
	query := `SELECT id, name, description, faction, type, row, power, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap FROM cards WHERE type = $1 ORDER BY faction, rarity`

	rows, err := r.db.Pool.Query(ctx, query, cardType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		var c models.Card
		var abilitiesJSON []byte
		var row sql.NullString

		err := rows.Scan(
			&c.ID, &c.Name, &c.Description, &c.Faction, &c.Type,
			&row, &c.Power, &c.IsHero,
			&abilitiesJSON, &c.ArtURL, &c.FlavorText, &c.CardGroup, &c.Rarity, &c.CostGold, &c.CostScrap,
		)
		if err != nil {
			return nil, err
		}

		if row.Valid {
			c.Row = models.RowType(row.String)
		}
		if len(abilitiesJSON) > 0 {
			json.Unmarshal(abilitiesJSON, &c.Abilities)
		}
		c.Nation = c.Faction
		cards = append(cards, &c)
	}

	return cards, nil
}

func (r *CardRepo) GetRandomCards(ctx context.Context, nations []models.Nation, rarityWeights map[models.Rarity]int, count int) ([]*models.Card, error) {
	// Build a weighted random query
	placeholders := make([]string, len(nations))
	args := make([]any, len(nations))
	for i, n := range nations {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = n
	}

	// For simplicity, just get random cards from the specified nations
	query := fmt.Sprintf(`
		SELECT id, name, description, faction, type, row, power, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap 
		FROM cards 
		WHERE faction IN (%s) 
		ORDER BY RANDOM() 
		LIMIT $%d`,
		joinPlaceholders(placeholders), len(nations)+1)

	args = append(args, count)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		var c models.Card
		var abilitiesJSON []byte
		var row sql.NullString

		err := rows.Scan(
			&c.ID, &c.Name, &c.Description, &c.Faction, &c.Type,
			&row, &c.Power, &c.IsHero,
			&abilitiesJSON, &c.ArtURL, &c.FlavorText, &c.CardGroup, &c.Rarity, &c.CostGold, &c.CostScrap,
		)
		if err != nil {
			return nil, err
		}

		if row.Valid {
			c.Row = models.RowType(row.String)
		}
		if len(abilitiesJSON) > 0 {
			json.Unmarshal(abilitiesJSON, &c.Abilities)
		}
		c.Nation = c.Faction
		cards = append(cards, &c)
	}

	return cards, nil
}

func (r *CardRepo) GetAllCards(ctx context.Context) ([]*models.Card, error) {
	query := `SELECT id, name, description, faction, type, row, power, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap FROM cards ORDER BY faction, rarity`

	log.Printf("DEBUG GetAllCards: executing query")
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		log.Printf("ERROR GetAllCards: query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		var c models.Card
		var abilitiesJSON []byte
		var row sql.NullString

		err := rows.Scan(
			&c.ID, &c.Name, &c.Description, &c.Faction, &c.Type,
			&row, &c.Power, &c.IsHero,
			&abilitiesJSON, &c.ArtURL, &c.FlavorText, &c.CardGroup, &c.Rarity, &c.CostGold, &c.CostScrap,
		)
		if err != nil {
			log.Printf("ERROR GetAllCards: scan failed: %v", err)
			return nil, err
		}

		if row.Valid {
			c.Row = models.RowType(row.String)
		}
		if len(abilitiesJSON) > 0 {
			json.Unmarshal(abilitiesJSON, &c.Abilities)
		}
		c.Nation = c.Faction
		cards = append(cards, &c)
	}

	log.Printf("DEBUG GetAllCards: returned %d cards", len(cards))
	if len(cards) > 0 {
		log.Printf("DEBUG GetAllCards: first card: %s (ID: %s, Faction: %s, Rarity: %s)", cards[0].Name, cards[0].ID, cards[0].Faction, cards[0].Rarity)
	}
	return cards, nil
}

func (r *CardRepo) GetFundamentalCards(ctx context.Context) ([]*models.Card, error) {
	query := `SELECT id, name, description, faction, type, row, power, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap FROM cards WHERE rarity IN ($1, $2) ORDER BY faction, rarity`

	rows, err := r.db.Pool.Query(ctx, query, models.RarityCommon, models.RarityRare)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		var c models.Card
		var abilitiesJSON []byte
		var row sql.NullString

		err := rows.Scan(
			&c.ID, &c.Name, &c.Description, &c.Faction, &c.Type,
			&row, &c.Power, &c.IsHero,
			&abilitiesJSON, &c.ArtURL, &c.FlavorText, &c.CardGroup, &c.Rarity, &c.CostGold, &c.CostScrap,
		)
		if err != nil {
			return nil, err
		}

		if row.Valid {
			c.Row = models.RowType(row.String)
		}
		if len(abilitiesJSON) > 0 {
			json.Unmarshal(abilitiesJSON, &c.Abilities)
		}
		c.Nation = c.Faction
		cards = append(cards, &c)
	}

	return cards, nil
}

// GetShopItems returns all available shop items
func (r *CardRepo) GetShopItems(ctx context.Context) ([]*models.ShopItem, error) {
	query := `SELECT id, name, description, card_id, cost_gold, cost_scrap, stock, level_required FROM shop_items WHERE stock != 0 ORDER BY level_required, cost_gold`

	log.Printf("DEBUG GetShopItems: executing query")
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		log.Printf("ERROR GetShopItems: query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []*models.ShopItem
	for rows.Next() {
		var item models.ShopItem
		var cardID sql.NullString

		err := rows.Scan(
			&item.ID, &item.Name, &item.Description, &cardID,
			&item.CostGold, &item.CostScrap, &item.Stock, &item.LevelRequired,
		)
		if err != nil {
			log.Printf("ERROR GetShopItems: scan failed: %v", err)
			return nil, err
		}

		if cardID.Valid {
			item.CardID = cardID.String
		}
		items = append(items, &item)
	}

	log.Printf("DEBUG GetShopItems: returned %d shop items", len(items))
	for i, item := range items {
		if i < 5 {
			log.Printf("DEBUG GetShopItems: item[%d]: ID=%s, Name=%s, CardID=%s, CostGold=%d, CostScrap=%d, Stock=%d, LevelReq=%d", 
				i, item.ID, item.Name, item.CardID, item.CostGold, item.CostScrap, item.Stock, item.LevelRequired)
		}
	}
	return items, nil
}

// GetShopItem returns a specific shop item by ID
func (r *CardRepo) GetShopItem(ctx context.Context, itemID string) (*models.ShopItem, error) {
	query := `SELECT id, name, description, card_id, cost_gold, cost_scrap, stock, level_required FROM shop_items WHERE id = $1`

	var item models.ShopItem
	var cardID sql.NullString

	err := r.db.Pool.QueryRow(ctx, query, itemID).Scan(
		&item.ID, &item.Name, &item.Description, &cardID,
		&item.CostGold, &item.CostScrap, &item.Stock, &item.LevelRequired,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if cardID.Valid {
		item.CardID = cardID.String
	}

	return &item, nil
}

func joinPlaceholders(placeholders []string) string {
	result := ""
	for i, p := range placeholders {
		if i > 0 {
			result += ","
		}
		result += p
	}
	return result
}