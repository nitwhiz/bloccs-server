package server

import (
	"bloccs-server/pkg/bloccs"
	"bloccs-server/pkg/event"
	"math/rand"
	"time"
)

const EventDistributeBedrock = "room_distribute_bedrock"

type DistributeBedrockPayload struct {
	From   *bloccs.Game `json:"from"`
	To     *bloccs.Game `json:"to"`
	Amount int          `json:"amount"`
}

func (p *DistributeBedrockPayload) RLock() {
	p.From.RLock()
	p.To.RLock()
}

func (p *DistributeBedrockPayload) RUnlock() {
	p.From.RUnlock()
	p.To.RUnlock()
}

func (r *Room) cycleBedrockTargetMap() {
	gameCount := len(r.Games)

	if gameCount == 1 {
		return
	}

	var games []*bloccs.Game

	for _, game := range r.Games {
		games = append(games, game)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	rng.Shuffle(gameCount, func(i int, j int) {
		games[i], games[j] = games[j], games[i]
	})

	r.BedrockTargetMap = map[string]string{}

	for i := 0; i < gameCount; i++ {
		if i == 0 {
			r.BedrockTargetMap[games[gameCount-1].GetId()] = games[0].GetId()
		} else {
			r.BedrockTargetMap[games[i-1].GetId()] = games[i].GetId()
		}
	}

	r.eventBus.Publish(event.New(EventUpdateBedrockTargetMap, r, nil))
}

func (r *Room) startBedrockDistributor() {
	go func() {
		r.waitGroup.Add(1)
		defer r.waitGroup.Done()

		for {
			select {
			case <-r.ctx.Done():
				return
			case packet := <-r.bedrockChannel:
				r.mu.RLock()

				if targetGameId, ok := r.BedrockTargetMap[packet.Game.GetId()]; ok {
					if targetGame, ok := r.Games[targetGameId]; ok && targetGame.IsRunning() {
						r.eventBus.Publish(event.New(EventDistributeBedrock, r, &DistributeBedrockPayload{
							From:   packet.Game,
							To:     targetGame,
							Amount: packet.Amount,
						}))

						targetGame.GetField().IncreaseBedrock(packet.Amount)
					}
				}

				r.mu.RUnlock()

				break
			}
		}
	}()
}
