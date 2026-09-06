package tarot

import (
	_ "embed"
	"encoding/json"
	"math/rand"
	"strings"
	"sync"
	"time"
)

//go:embed cards.json
var cardsJSON []byte

const baseImageURL = "https://raw.githubusercontent.com/jeremytarling/python-tarot/refs/heads/master/webapp/static/"

// Card represents a Tarot card entity.
type Card struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	Image        string `json:"image"`
	Desc         string `json:"desc"`
	Message      string `json:"message"`
	RDesc        string `json:"rdesc"`
	Sequence     int    `json:"sequence"`
	Qabalah      string `json:"qabalah"`
	HebrewLetter string `json:"hebrew_letter"`
	Meditation   string `json:"meditation"`
	CardType     string `json:"cardtype"`
}

// Service provides tarot operations.
type Service struct {
	cards []Card
	mu    sync.Mutex
	rng   *rand.Rand
}

var (
	defaultService *Service
	once           sync.Once
)

// NewService creates a new Tarot service instance with embedded deck.
func NewService() *Service {
	var deck []Card
	_ = json.Unmarshal(cardsJSON, &deck)

	// Filter cards that have a message
	var validCards []Card
	for _, c := range deck {
		if strings.TrimSpace(c.Message) != "" {
			validCards = append(validCards, c)
		}
	}

	return &Service{
		cards: validCards,
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// DefaultService returns the singleton tarot service.
func DefaultService() *Service {
	once.Do(func() {
		defaultService = NewService()
	})
	return defaultService
}

// GetCard returns a randomly selected tarot card with full image URL.
func (s *Service) GetCard() Card {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.cards) == 0 {
		return Card{Name: "The Fool", Message: "Begin the journey", Image: baseImageURL + "images/00.jpeg"}
	}

	idx := s.rng.Intn(len(s.cards))
	card := s.cards[idx]
	if !strings.HasPrefix(card.Image, "http") {
		card.Image = baseImageURL + card.Image
	}
	return card
}
