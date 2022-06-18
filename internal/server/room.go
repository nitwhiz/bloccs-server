package server

import (
	"bloccs-server/pkg/bloccs"
	"bloccs-server/pkg/event"
	"context"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

const EventUpdateBedrockTargetMap = "room_update_bedrock_target_map"
const EventRoomStart = "room_start"
const EventRoomStop = "room_stop"

type Room struct {
	Id               string                  `json:"id"`
	Players          map[string]*Player      `json:"players"`
	Games            map[string]*bloccs.Game `json:"games"`
	HostPlayer       *Player                 `json:"hostPlayer"`
	BedrockTargetMap map[string]string       `json:"bedrockTargetMap"`
	bedrockChannel   chan *bloccs.BedrockPacket
	eventHandlerIds  map[string]int
	eventBus         *event.Bus
	createAt         time.Time
	gamesRunning     bool
	isStopping       bool
	ctx              context.Context
	stopFunc         context.CancelFunc
	waitGroup        *sync.WaitGroup
	mu               *sync.RWMutex
}

func NewRoom() *Room {
	b := event.NewBus()
	b.Start()

	r := &Room{
		Id:               uuid.NewString(),
		Players:          map[string]*Player{},
		HostPlayer:       nil,
		Games:            map[string]*bloccs.Game{},
		BedrockTargetMap: map[string]string{},
		bedrockChannel:   make(chan *bloccs.BedrockPacket),
		eventHandlerIds:  map[string]int{},
		eventBus:         b,
		createAt:         time.Now(),
		gamesRunning:     false,
		isStopping:       false,
		ctx:              nil,
		stopFunc:         nil,
		waitGroup:        &sync.WaitGroup{},
		mu:               &sync.RWMutex{},
	}

	return r
}

func (r *Room) RLock() {
	r.mu.RLock()
}

func (r *Room) RUnlock() {
	r.mu.RUnlock()
}

func (r *Room) HasGamesRunning() bool {
	defer r.mu.RUnlock()
	r.mu.RLock()

	return r.gamesRunning
}

func (r *Room) GetId() string {
	return r.Id
}

func (r *Room) GetPlayerCount() int {
	defer r.mu.RUnlock()
	r.mu.RLock()

	return len(r.Players)
}

func (r *Room) hypervisorTick() {
	defer r.mu.RUnlock()
	r.mu.RLock()

	runningGames := 0

	for _, g := range r.Games {
		if !g.IsGameOver() {
			runningGames++
		}
	}

	if runningGames <= 1 {
		go r.Stop()
	}
}

func (r *Room) hypervisor() {
	defer r.waitGroup.Done()
	r.waitGroup.Add(1)

	// todo: use global ticker?
	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			r.hypervisorTick()
			break
		}
	}
}

func (r *Room) Join(conn *websocket.Conn) error {
	g := bloccs.NewGame(r.eventBus, r.bedrockChannel, bloccs.GameSettings{
		FallingPieceSpeed: 1,
		Seed:              r.Id,
		FieldWidth:        10,
		FieldHeight:       20,
	})

	p := NewPlayer(conn, g.GetId(), g.GetCommandChannel())

	if err := r.handshakeHello(p); err != nil {
		return err
	}

	r.RegisterPlayer(p, g)

	p.Listen(func() {
		r.RemoveGame(g)
	})

	return nil
}

func (r *Room) AreGamesRunning() bool {
	return r.gamesRunning
}

func (r *Room) ShouldStop() bool {
	defer r.mu.RUnlock()
	r.mu.RLock()

	if !r.isStopping && len(r.Players) == 0 && time.Since(r.createAt) > time.Minute*15 {
		return true
	}

	return false
}

func (r *Room) Start() {
	defer r.mu.Unlock()
	r.mu.Lock()

	if r.gamesRunning {
		return
	}

	r.reset()

	r.eventBus.Publish(event.New(EventRoomStart, r, nil))

	r.startBedrockDistributor()
	go r.hypervisor()

	for _, g := range r.Games {
		g.Start()
	}

	r.gamesRunning = true
}

func (r *Room) Stop() {
	defer r.mu.Unlock()
	r.mu.Lock()

	if r.isStopping {
		return
	}

	r.isStopping = true

	r.eventBus.Publish(event.New(EventRoomStop, r, nil))

	log.Printf("stopping room %s ...\n", r.Id)

	for _, g := range r.Games {
		g.Stop()
	}

	if r.stopFunc != nil {
		r.stopFunc()
	} else {
		log.Println("no stop func for room, stopping may never finish")
	}

	r.waitGroup.Wait()

	r.gamesRunning = false

	log.Println("stopped!")
}

func (r *Room) reset() {
	r.BedrockTargetMap = map[string]string{}

	r.gamesRunning = false
	r.isStopping = false

	r.cycleBedrockTargetMap()

	r.ctx, r.stopFunc = context.WithCancel(context.Background())
}
