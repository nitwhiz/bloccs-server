package bloccs

import (
	"bloccs-server/pkg/event"
	"sync"
)

const EventScoreUpdate = "score_update"

type Score struct {
	GameId   string `json:"gameId"`
	Score    int    `json:"score"`
	Lines    int    `json:"lines"`
	mu       *sync.RWMutex
	eventBus *event.Bus
}

func NewScore(bus *event.Bus, gameId string) *Score {
	return &Score{
		GameId:   gameId,
		Score:    0,
		Lines:    0,
		mu:       &sync.RWMutex{},
		eventBus: bus,
	}
}

func (s *Score) RLock() {
	s.mu.RLock()
}

func (s *Score) RUnlock() {
	s.mu.RUnlock()
}

func (s *Score) GetId() string {
	defer s.mu.RUnlock()
	s.mu.RLock()

	return s.GameId
}

func (s *Score) Reset() {
	defer s.mu.Unlock()
	s.mu.Lock()

	s.Score = 0
	s.Lines = 0

	if s.eventBus != nil {
		s.eventBus.Publish(event.New(EventScoreUpdate, s, nil))
	}
}

func (s *Score) AddLines(l int) {
	defer s.mu.Unlock()
	s.mu.Lock()

	if l == 0 {
		return
	}

	switch l {
	case 1:
		s.Score += 60
		break
	case 2:
		s.Score += 150
		break
	case 3:
		s.Score += 420
		break
	case 4:
		s.Score += 2500
		break
	default:
		break
	}

	s.Lines += l

	if s.eventBus != nil {
		s.eventBus.Publish(event.New(EventScoreUpdate, s, nil))
	}
}
