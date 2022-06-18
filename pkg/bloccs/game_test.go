package bloccs

import (
	"bloccs-server/pkg/event"
	"testing"
)

func TestGame_GetId(t *testing.T) {
	game := NewGame(nil, make(chan *BedrockPacket), GameSettings{
		FallingPieceSpeed: 1,
		Seed:              "D34D-B33F",
		FieldWidth:        0,
		FieldHeight:       0,
	})

	if game.GetId() == "" {
		t.Fatalf("expected non-empty id")
	}
}

func TestGame_RLock(t *testing.T) {
	game := NewGame(nil, make(chan *BedrockPacket), GameSettings{
		FallingPieceSpeed: 1,
		Seed:              "D34D-B33F",
		FieldWidth:        0,
		FieldHeight:       0,
	})

	game.RLock()

	if game.mu.TryLock() {
		t.Fatal("expected field mutex to be not lockable")
	}
}

func TestGame_RUnlock(t *testing.T) {
	game := NewGame(nil, make(chan *BedrockPacket), GameSettings{
		FallingPieceSpeed: 1,
		Seed:              "D34D-B33F",
		FieldWidth:        0,
		FieldHeight:       0,
	})

	game.RLock()
	game.RUnlock()

	if !game.mu.TryLock() {
		t.Fatal("expected field mutex to be lockable")
	}
}

func TestGame_StartStop(t *testing.T) {
	game := NewGame(nil, make(chan *BedrockPacket), GameSettings{
		FallingPieceSpeed: 1,
		Seed:              "D34D-B33F",
		FieldWidth:        6,
		FieldHeight:       6,
	})

	game.Start()
	game.Stop()
}

func TestGame_StartEmitsEvent(t *testing.T) {
	bus := event.NewBus()

	bus.Start()

	game := NewGame(bus, make(chan *BedrockPacket), GameSettings{
		FallingPieceSpeed: 1,
		Seed:              "D34D-B33F",
		FieldWidth:        6,
		FieldHeight:       6,
	})

	eventCount := 0
	expectedEventCount := 1

	bus.Subscribe(EventGameStart, func(event *event.Event) {
		if g, _ := event.Source.(*Game); g != game {
			t.Fatalf("event source is not game")
		}

		eventCount += 1
	})

	game.Start()
	game.Stop()

	bus.Stop()

	if eventCount != expectedEventCount {
		t.Fatalf("missing events. exepected %d, got %d", expectedEventCount, eventCount)
	}
}

func TestGame_GetField(t *testing.T) {
	game := NewGame(nil, make(chan *BedrockPacket), GameSettings{
		FallingPieceSpeed: 1,
		Seed:              "D34D-B33F",
		FieldWidth:        3,
		FieldHeight:       3,
	})

	field := game.GetField()

	if field == nil {
		t.Fatal("expected field, got nil")
	}

	if field.width != 3 {
		t.Fatalf("expected field width 3, got %d", field.width)
	}

	if field.height != 3 {
		t.Fatalf("expected field height 3, got %d", field.height)
	}
}

func TestGame_IsRunning(t *testing.T) {
	game := NewGame(nil, make(chan *BedrockPacket), GameSettings{
		FallingPieceSpeed: 1,
		Seed:              "D34D-B33F",
		FieldWidth:        3,
		FieldHeight:       3,
	})

	game.Start()

	if game.IsRunning() == false {
		t.Fatal("expected game to be running")
	}
}

func TestGame_GetCommandChannel(t *testing.T) {
	game := NewGame(nil, make(chan *BedrockPacket), GameSettings{
		FallingPieceSpeed: 1,
		Seed:              "D34D-B33F",
		FieldWidth:        3,
		FieldHeight:       3,
	})

	channel := game.GetCommandChannel()

	if channel == nil {
		t.Fatal("expected game command channel to be not nil")
	}
}
