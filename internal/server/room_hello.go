package server

import (
	"bloccs-server/pkg/event"
	"encoding/json"
	"errors"
	"strings"
)

const EventHello = "hello"
const EventHelloAck = "hello_ack"

type HelloResponseMessage struct {
	Name string `json:"name"`
}

type HelloAckPayload struct {
	Player *Player `json:"player"`
}

func (h *HelloAckPayload) RLock() {
	h.Player.RLock()
}

func (h *HelloAckPayload) RUnlock() {
	h.Player.RUnlock()
}

func listenForHello(p *Player) (*HelloResponseMessage, error) {
	msg, err := p.ReadMessage()

	if err != nil {
		return nil, err
	}

	helloResponse := HelloResponseMessage{}

	err = json.Unmarshal(msg, &helloResponse)

	if err != nil || helloResponse.Name == "" {
		return nil, errors.New("invalid hello response")
	}

	return &helloResponse, nil
}

func (r *Room) sendHello(p *Player) error {
	return r.sendEventToPlayer(event.New(EventHello, nil, nil), p)
}

func sanitizeName(name string) string {
	name = strings.TrimSpace(name)

	if len(name) > 8 {
		name = name[:8]
	}

	return strings.ToUpper(name)
}

func (r *Room) handshakeHello(p *Player) error {
	if err := r.sendHello(p); err != nil {
		return err
	}

	helloResponse, err := listenForHello(p)

	if err != nil {
		return err
	}

	p.Name = sanitizeName(helloResponse.Name)

	return r.sendEventToPlayer(event.New(EventHelloAck, r, &HelloAckPayload{Player: p}), p)
}
