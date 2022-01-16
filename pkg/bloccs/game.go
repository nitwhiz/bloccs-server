package bloccs

import (
	"bloccs-server/pkg/event"
	"fmt"
	"sync"
	"time"
)

type UpdateHandler func()

type Game struct {
	ID              string
	Field           *Field
	EventBus        *event.Bus
	Over            bool
	stopChannel     chan bool
	globalWaitGroup *sync.WaitGroup
}

func NewGame(bus *event.Bus, id string) *Game {
	game := &Game{
		ID:              id,
		Field:           NewField(bus, 10, 20, id),
		EventBus:        bus,
		Over:            false,
		stopChannel:     make(chan bool),
		globalWaitGroup: &sync.WaitGroup{},
	}

	bus.AddChannel(fmt.Sprintf("update/%s", id))

	return game
}

func (g *Game) Start() {
	g.globalWaitGroup.Add(1)

	go func() {
		defer g.globalWaitGroup.Done()

		// 100 fps
		ticker := time.NewTicker(time.Millisecond * 10)

		defer ticker.Stop()

		for {
			select {
			case <-g.stopChannel:
				return
			case <-ticker.C:
				g.Update()
			}
		}
	}()
}

func (g *Game) Stop() {
	close(g.stopChannel)
	g.globalWaitGroup.Wait()

	g.EventBus.RemoveChannel(fmt.Sprintf("update/%s", g.ID))
}

// todo: should actually be in field

func (g *Game) PublishFieldUpdate() {
	g.EventBus.Publish(event.New(fmt.Sprintf("update/%s", g.ID), EventGameFieldUpdate, &event.Payload{
		"field": g.Field,
	}))
}

func (g *Game) PublishFallingPieceUpdate() {
	g.EventBus.Publish(event.New(fmt.Sprintf("update/%s", g.Field.ID), EventUpdateFallingPiece, &event.Payload{
		"falling_piece_data": g.Field.FallingPiece,
		"piece_display":      g.Field.FallingPiece.CurrentPiece.GetData(),
	}))
}

func (g *Game) Update() {
	if g.Over {
		return
	}

	dirty, gameOver := g.Field.Update()

	if dirty {
		g.PublishFieldUpdate()

		// todo: g.Field should not be manipulated from here; maintain dirty flags from outside?; diffing?
		g.Field.Dirty = false
	}

	if gameOver {
		g.EventBus.Publish(event.New(fmt.Sprintf("update/%s", g.ID), EventGameOver, nil))
	}

	g.Over = gameOver
}

// Command returns true if the command was understood
func (g *Game) Command(cmd string) bool {
	// todo: refactor command no not query g.Field twice

	switch cmd {
	case "L":
		g.Field.FallingPiece.Move(g.Field, -1, 0, 0)
		return true
	case "R":
		g.Field.FallingPiece.Move(g.Field, 1, 0, 0)
		return true
	case "D":
		g.Field.FallingPiece.Move(g.Field, 0, 1, 0)
		return true
	case "P":
		g.Field.FallingPiece.Punch(g.Field)
		return true
	case "X":
		g.Field.FallingPiece.Move(g.Field, 0, 0, 1)
		return true
	default:
		return false
	}
}
