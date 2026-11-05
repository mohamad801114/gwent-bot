package cards

import (
	"context"
	"math/rand"
	"time"

	"gwent-bot/internal/database"
	"gwent-bot/pkg/models"
)

type Service struct {
	repo *database.CardRepo
}

func NewService(repo *database.CardRepo) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetCard(ctx context.Context, cardID string) (*models.Card, error) {
	return s.repo.GetCard(ctx, cardID)
}

func (s *Service) GetCardsByNation(ctx context.Context, nation models.Nation) ([]*models.Card, error) {
	return s.repo.GetCardsByNation(ctx, nation)
}

func (s *Service) GetCardsByRarity(ctx context.Context, rarity models.Rarity) ([]*models.Card, error) {
	return s.repo.GetCardsByRarity(ctx, rarity)
}

func (s *Service) GetAllCards(ctx context.Context) ([]*models.Card, error) {
	return s.repo.GetAllCards(ctx)
}

func (s *Service) GetFundamentalCards(ctx context.Context) ([]*models.Card, error) {
	return s.repo.GetFundamentalCards(ctx)
}

func (s *Service) GetRandomCards(ctx context.Context, nations []models.Nation, rarityWeights map[models.Rarity]int, count int) ([]*models.Card, error) {
	return s.repo.GetRandomCards(ctx, nations, rarityWeights, count)
}

func (s *Service) CreateCardInstance(card *models.Card, ownerID int64) *models.CardInstance {
	return &models.CardInstance{
		CardID:     card.ID,
		InstanceID: generateInstanceID(),
		Power:      card.Power,
		Armor:      card.Armor,
		Row:        card.Row,
		OwnerID:    ownerID,
		Statuses:   []models.Status{},
	}
}

func (s *Service) CreateCardInstances(card *models.Card, ownerID int64, count int) []*models.CardInstance {
	instances := make([]*models.CardInstance, count)
	for i := 0; i < count; i++ {
		instances[i] = s.CreateCardInstance(card, ownerID)
	}
	return instances
}

func (s *Service) GetStarterDeck(ctx context.Context, nation models.Nation) ([]*models.Card, error) {
	cards, err := s.repo.GetCardsByNation(ctx, nation)
	if err != nil {
		return nil, err
	}

	var starterCards []*models.Card
	for _, card := range cards {
		if (card.Rarity == models.RarityCommon || card.Rarity == models.RarityRare) && card.Provision <= 8 {
			starterCards = append(starterCards, card)
		}
	}

	if len(starterCards) > 25 {
		rand.Shuffle(len(starterCards), func(i, j int) {
			starterCards[i], starterCards[j] = starterCards[j], starterCards[i]
		})
		starterCards = starterCards[:25]
	}

	return starterCards, nil
}

func generateInstanceID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
