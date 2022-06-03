package bloccs

import (
	"bloccs-server/pkg/event"
	"sync"
	"time"
)

const EventGameFallingPieceUpdate = "game_falling_piece_update"
const EventGameNextPieceUpdate = "game_next_piece_update"
const EventGameHoldPieceUpdate = "game_hold_piece_update"

type FallingPiece struct {
	GameId    string `json:"gameId"`
	Piece     *Piece `json:"piece"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Rotation  int    `json:"rotation"`
	FallTimer int    `json:"fallTimer"`
	eventBus  *event.Bus
	mu        *sync.RWMutex
}

func NewFallingPiece(eventBus *event.Bus, gameId string) *FallingPiece {
	return &FallingPiece{
		GameId:    gameId,
		Piece:     nil,
		X:         0,
		Y:         0,
		FallTimer: 0,
		eventBus:  eventBus,
		mu:        &sync.RWMutex{},
	}
}

func (f *FallingPiece) GetId() string {
	return f.GameId
}

func (f *FallingPiece) publishUpdate() {
	f.eventBus.Publish(event.New(EventGameFallingPieceUpdate, f, nil))
}

func (g *Game) canMoveFallingPiece(dr int, dx int, dy int) bool {
	if g.fallingPiece.Piece != nil && g.field.CanPutPiece(
		g.fallingPiece.Piece,
		g.fallingPiece.Rotation+dr,
		g.fallingPiece.X+dx,
		g.fallingPiece.Y+dy,
	) {
		return true
	}

	return false
}

func (g *Game) moveFallingPiece(dr int, dx int, dy int) {
	if g.canMoveFallingPiece(dr, dx, dy) {
		g.fallingPiece.Rotation += dr
		g.fallingPiece.X += dx
		g.fallingPiece.Y += dy

		g.fallingPiece.publishUpdate()
	}
}

func (g *Game) getFallTimer() int {
	return int(1000.0 / g.fallingPieceSpeed)
}

func (g *Game) initFallingPiece() {
	g.fallingPiece.X = g.field.CenterX
	g.fallingPiece.Y = 0
	g.fallingPiece.Rotation = 0

	g.fallingPiece.FallTimer = g.getFallTimer()

	g.fallingPiece.publishUpdate()
}

func (g *Game) setFallingPiece(p *Piece) {
	g.fallingPiece.Piece = p

	g.initFallingPiece()
}

func (g *Game) nextFallingPiece() {
	g.setFallingPiece(g.nextPiece)

	g.nextPiece = g.rbg.NextPiece()

	g.fallingPiece.publishUpdate()

	g.eventBus.Publish(event.New(EventGameNextPieceUpdate, g, g.nextPiece))
}

func (g *Game) lockFallingPiece() int {
	if g.fallingPiece.Piece != nil {
		g.field.PutPiece(
			g.fallingPiece.Piece,
			g.fallingPiece.Rotation,
			g.fallingPiece.X,
			g.fallingPiece.Y,
		)
	}

	cleared := g.field.ClearFullRows()

	g.nextFallingPiece()

	g.holdLock = false

	return cleared
}

func (g *Game) updateFallingPiece(delta int) (int, bool) {
	clearedLines := 0
	gameOver := false

	if g.fallingPiece.Piece == nil {
		g.nextFallingPiece()
		g.holdLock = false
	} else {
		g.fallingPiece.FallTimer -= delta / int(time.Millisecond)

		if g.fallingPiece.FallTimer <= 0 {
			g.fallingPiece.FallTimer = g.getFallTimer()

			if g.canMoveFallingPiece(0, 0, 1) {
				g.moveFallingPiece(0, 0, 1)
			} else {
				clearedLines = g.lockFallingPiece()

				if !g.canMoveFallingPiece(0, 0, 1) &&
					!g.canMoveFallingPiece(0, 1, 0) &&
					!g.canMoveFallingPiece(0, -1, 0) {
					gameOver = true
				}
			}
		}
	}

	return clearedLines, gameOver
}

func (g *Game) holdFallingPiece() {
	if g.fallingPiece.Piece == nil {
		return
	}

	if g.holdLock {
		return
	}

	g.holdLock = true

	if g.holdPiece != nil {
		g.fallingPiece.Piece, g.holdPiece = g.holdPiece, g.fallingPiece.Piece

		g.initFallingPiece()
	} else {
		g.holdPiece = g.fallingPiece.Piece

		g.nextFallingPiece()
	}

	g.eventBus.Publish(event.New(EventGameHoldPieceUpdate, g, g.holdPiece))
}

func (g *Game) hardLockFallingPiece() {
	if g.fallingPiece.Piece == nil {
		return
	}

	dy := 0

	for dy < g.field.Height {
		if !g.canMoveFallingPiece(0, 0, dy) {
			break
		}

		dy++
	}

	g.fallingPiece.Y += dy - 1
	g.fallingPiece.FallTimer = 0
}