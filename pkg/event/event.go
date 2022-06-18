package event

import (
	"encoding/json"
)

type RLocker interface {
	RLock()
	RUnlock()
}

type Source interface {
	GetId() string
	RLocker
}

type Type string

type Payload interface {
	RLocker
}

type Event struct {
	Source  Source  `json:"source"`
	Type    Type    `json:"type"`
	Payload Payload `json:"payload"`
}

func New(eventType Type, source Source, payload Payload) *Event {
	return &Event{
		Source:  source,
		Type:    eventType,
		Payload: payload,
	}
}

func (e *Event) GetAsBytes() ([]byte, error) {
	if e.Source != nil {
		defer e.Source.RUnlock()
		e.Source.RLock()
	}

	if e.Payload != nil {
		defer e.Payload.RUnlock()
		e.Payload.RLock()
	}

	bs, err := json.Marshal(e)

	if err != nil {
		return nil, err
	}

	return bs, nil
}
