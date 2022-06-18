package bloccs

import (
	"bloccs-server/pkg/event"
	"bloccs-server/pkg/rng"
	"context"
	"github.com/google/uuid"
	"hash/fnv"
	"log"
	"sync"
	"time"
)

const EventGameStart = "game_start"
const EventGameOver = "game_over"

type GameSettings struct {
	// FallingPieceSpeed falling piece speed in blocks per second
	FallingPieceSpeed float64 `json:"-"`
	Seed              string  `json:"-"`
	FieldWidth        int     `json:"fieldWidth"`
	FieldHeight       int     `json:"fieldHeight"`
}

type BedrockPacket struct {
	Game   *Game
	Amount int
}

type Game struct {
	Id             string       `json:"id"`
	Settings       GameSettings `json:"settings"`
	running        bool
	gameOver       bool
	bedrockChannel chan *BedrockPacket
	commandChannel chan string
	fallingPiece   *FallingPiece
	holdPiece      *Piece
	holdLock       bool
	nextPiece      *Piece
	field          *Field
	eventBus       *event.Bus
	score          *Score
	rbg            *rng.RPG[Piece]
	lastUpdate     *time.Time
	mu             *sync.RWMutex
	ctx            context.Context
	stopFunc       context.CancelFunc
}

func (g *Game) RLock() {
	g.mu.RLock()
}

func (g *Game) RUnlock() {
	g.mu.RUnlock()
}

func (g *Game) GetId() string {
	defer g.mu.RUnlock()
	g.mu.RLock()

	return g.Id
}

func (g *Game) GetField() *Field {
	defer g.mu.RUnlock()
	g.mu.RLock()

	return g.field
}

func NewGame(bus *event.Bus, bedrockChannel chan *BedrockPacket, settings GameSettings) *Game {
	h := fnv.New32a()

	_, _ = h.Write([]byte(settings.Seed))

	rpg := rng.NewRBG(int64(h.Sum32()), func() []*Piece {
		var b []*Piece

		for _, p := range AllPieces {
			b = append(b, p)
		}

		return b
	})

	gameId := uuid.NewString()

	game := &Game{
		Id:             gameId,
		Settings:       settings,
		running:        false,
		gameOver:       false,
		commandChannel: make(chan string),
		bedrockChannel: bedrockChannel,
		fallingPiece:   NewFallingPiece(bus, gameId),
		holdPiece:      nil,
		holdLock:       false,
		nextPiece:      nil,
		field:          NewField(bus, gameId, settings.FieldWidth, settings.FieldHeight),
		eventBus:       bus,
		score:          NewScore(bus, gameId),
		rbg:            rpg,
		lastUpdate:     nil,
		mu:             &sync.RWMutex{},
		ctx:            nil,
		stopFunc:       nil,
	}

	game.nextFallingPiece()

	return game
}

func (g *Game) IsRunning() bool {
	defer g.mu.RUnlock()
	g.mu.RLock()

	return g.running
}

func (g *Game) IsGameOver() bool {
	defer g.mu.RUnlock()
	g.mu.RLock()

	return g.gameOver
}

func (g *Game) GetCommandChannel() chan string {
	defer g.mu.RUnlock()
	g.mu.RLock()

	return g.commandChannel
}

func (g *Game) Start() {
	defer g.mu.Unlock()
	g.mu.Lock()

	g.reset()

	g.rbg.NextBag()

	now := time.Now()
	g.lastUpdate = &now

	go g.loop()

	g.running = true

	if g.eventBus != nil {
		g.eventBus.Publish(event.New(EventGameStart, g, nil))
	}
}

func (g *Game) reset() {
	g.running = false
	g.gameOver = false
	g.fallingPiece.Reset(g)
	g.holdPiece = nil
	g.holdLock = false
	g.nextPiece = nil
	g.field.Reset()
	g.score.Reset()
	g.lastUpdate = nil
	g.ctx, g.stopFunc = context.WithCancel(context.Background())

	if g.eventBus != nil {
		g.eventBus.Publish(event.New(EventGameHoldPieceUpdate, g, nil))
		g.eventBus.Publish(event.New(EventGameNextPieceUpdate, g, nil))
	}
}

func (g *Game) Stop() {
	defer g.mu.Unlock()
	g.mu.Lock()

	if !g.running {
		return
	}

	g.running = false

	if g.stopFunc != nil {
		g.stopFunc()
	} else {
		log.Println("no stop func for game, stopping may never finish")
	}
}

func (g *Game) Update() bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.gameOver {
		return true
	}

	now := time.Now()

	if g.lastUpdate != nil {
		clearedRows, gameOver := g.updateFallingPiece(int(time.Now().Sub(*g.lastUpdate)))

		g.score.AddLines(clearedRows)
		producedBedrock := g.field.DecreaseBedrock(clearedRows)

		if producedBedrock != 0 {
			g.bedrockChannel <- &BedrockPacket{
				Game:   g,
				Amount: producedBedrock,
			}
		}

		if gameOver {
			if g.eventBus != nil {
				g.eventBus.Publish(event.New(EventGameOver, g, nil))
			}

			g.gameOver = gameOver
		}
	}

	g.lastUpdate = &now

	return true
}

func (g *Game) loop() {
	defer g.Stop()

	// 100 tps
	ticker := time.NewTicker(time.Millisecond * 10)

	defer ticker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return
		case cmd := <-g.commandChannel:
			g.Command(cmd)

			if !g.Update() {
				return
			}

			break
		case <-ticker.C:
			if !g.Update() {
				return
			}

			break
		}
	}
}
