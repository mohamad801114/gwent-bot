package database

import (
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
	query := `SELECT id, name, description, faction, type, row, power, armor, provision, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap FROM cards WHERE id = $1`

	var c models.Card
	var abilitiesJSON []byte
	var row sql.NullString

	err := r.db.Pool.QueryRow(ctx, query, cardID).Scan(
		&c.ID, &c.Name, &c.Description, &c.Faction, &c.Type,
		&row, &c.Power, &c.Armor, &c.Provision, &c.IsHero,
		&abilitiesJSON, &c.ArtURL, &c.FlavorText, &c.CardGroup, &c.Rarity, &c.CostGold, &c.CostScrap,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if row.Valid {
		c.Row = models.RowType(row.String)
	}
	if len(abilitiesJSON) > 0 {
		json.Unmarshal(abilitiesJSON, &c.Abilities)
	}

	// Set Nation as alias for Faction
	c.Nation = c.Faction

	return &c, nil
}

func (r *CardRepo) GetCardsByNation(ctx context.Context, nation models.Nation) ([]*models.Card, error) {
	query := `SELECT id, name, description, faction, type, row, power, armor, provision, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap FROM cards WHERE faction = $1 ORDER BY rarity, provision`

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
			&row, &c.Power, &c.Armor, &c.Provision, &c.IsHero,
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
	query := `SELECT id, name, description, faction, type, row, power, armor, provision, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap FROM cards WHERE rarity = $1 ORDER BY faction, provision`

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
			&row, &c.Power, &c.Armor, &c.Provision, &c.IsHero,
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
	query := `SELECT id, name, description, faction, type, row, power, armor, provision, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap FROM cards WHERE type = $1 ORDER BY faction, rarity, provision`

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
			&row, &c.Power, &c.Armor, &c.Provision, &c.IsHero,
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
		SELECT id, name, description, faction, type, row, power, armor, provision, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap 
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
			&row, &c.Power, &c.Armor, &c.Provision, &c.IsHero,
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
	query := `SELECT id, name, description, faction, type, row, power, armor, provision, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap FROM cards ORDER BY faction, rarity, provision`

	rows, err := r.db.Pool.Query(ctx, query)
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
			&row, &c.Power, &c.Armor, &c.Provision, &c.IsHero,
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

func (r *CardRepo) GetFundamentalCards(ctx context.Context) ([]*models.Card, error) {
	query := `SELECT id, name, description, faction, type, row, power, armor, provision, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap FROM cards WHERE rarity IN ($1, $2) AND provision <= 8 ORDER BY faction, rarity, provision`

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
			&row, &c.Power, &c.Armor, &c.Provision, &c.IsHero,
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