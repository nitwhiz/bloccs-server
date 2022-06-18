package server

import (
	"bloccs-server/pkg/bloccs"
	"bloccs-server/pkg/event"
	"log"
)

const EventPlayerJoin = "player_join"
const EventPlayerLeave = "player_leave"

// aka not-ingame events
var lobbyEvents = []event.Type{
	EventHello,
	EventHelloAck,
	EventPlayerJoin,
	EventPlayerLeave,
	bloccs.EventGameStart,
	EventUpdateBedrockTargetMap,
	EventRoomStart,
	EventRoomStop,
}

func isInGameEvent(eventType event.Type) bool {
	for _, ie := range lobbyEvents {
		if eventType == ie {
			return false
		}
	}

	return true
}

type PlayerJoinLeavePayload struct {
	Player *Player `json:"player"`
}

func (p *PlayerJoinLeavePayload) RLock() {
	p.Player.RLock()
}

func (p *PlayerJoinLeavePayload) RUnlock() {
	p.Player.RUnlock()
}

func (r *Room) sendEventToPlayer(e *event.Event, p *Player) error {
	if !r.HasGamesRunning() && isInGameEvent(e.Type) {
		return nil
	}

	bs, err := e.GetAsBytes()

	if err != nil {
		log.Printf("cannot get bytes of message")
		return err
	}

	if err := p.SendMessage(bs); err != nil {
		return err
	}

	return nil
}

func (r *Room) RegisterPlayer(p *Player, g *bloccs.Game) {
	defer r.mu.Unlock()
	r.mu.Lock()

	if _, ok := r.Players[p.GetId()]; ok {
		return
	}

	if _, ok := r.Games[g.GetId()]; ok {
		return
	}

	if len(r.Players) >= 7 {
		return
	}

	r.Players[p.GetId()] = p
	r.Games[g.GetId()] = g

	if r.HostPlayer == nil {
		// todo: promote random player to host player if this one leaves
		r.HostPlayer = p
	}

	r.eventHandlerIds[p.GetId()] = r.eventBus.Subscribe(event.All, func(event *event.Event) {
		if err := r.sendEventToPlayer(event, p); err != nil {
			log.Println("error passing event")

			r.RemoveGame(g)
		}
	})

	r.eventBus.Publish(event.New(EventPlayerJoin, r, &PlayerJoinLeavePayload{
		Player: p,
	}))

	r.cycleBedrockTargetMap()
}

func (r *Room) RemoveGame(g *bloccs.Game) {
	defer r.mu.Unlock()
	r.mu.Lock()

	id := g.GetId()

	if _, ok := r.Games[id]; ok {
		delete(r.Games, id)

		g.Stop()
	}

	if p, ok := r.Players[id]; ok {
		delete(r.Players, id)

		if hid, ok := r.eventHandlerIds[id]; ok {
			r.eventBus.Unsubscribe(hid)

			delete(r.eventHandlerIds, id)
		}

		r.eventBus.Publish(event.New(EventPlayerLeave, r, &PlayerJoinLeavePayload{
			Player: p,
		}))

		r.cycleBedrockTargetMap()
	}
}
