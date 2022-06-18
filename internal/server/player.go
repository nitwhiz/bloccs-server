package server

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Player struct {
	GameId              string `json:"gameId"`
	Name                string `json:"name"`
	CreateAt            int64  `json:"createAt"`
	gameCommandChannel  chan string
	Conn                *websocket.Conn `json:"-"`
	connWriteMutex      *sync.Mutex
	connReadMutex       *sync.Mutex
	listenLoopWaitGroup *sync.WaitGroup
	cancelListenLoop    context.CancelFunc
	mu                  *sync.RWMutex
}

type StopCallbackFunc func()

func NewPlayer(conn *websocket.Conn, gameId string, gameCommandChannel chan string) *Player {
	return &Player{
		GameId:              gameId,
		Name:                "",
		CreateAt:            time.Now().UnixMilli(),
		gameCommandChannel:  gameCommandChannel,
		Conn:                conn,
		connWriteMutex:      &sync.Mutex{},
		connReadMutex:       &sync.Mutex{},
		listenLoopWaitGroup: &sync.WaitGroup{},
		cancelListenLoop:    nil,
		mu:                  &sync.RWMutex{},
	}
}

func (p *Player) RLock() {
	p.mu.RLock()
}

func (p *Player) RUnlock() {
	p.mu.RUnlock()
}

func (p *Player) GetId() string {
	return p.GameId
}

func (p *Player) Listen(stopCallback StopCallbackFunc) {
	p.Conn.SetPongHandler(func(string) error {
		return p.Conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	})

	ctx, cancel := context.WithCancel(context.Background())

	p.cancelListenLoop = cancel

	go func() {
		defer func(Conn *websocket.Conn) {
			p.listenLoopWaitGroup.Done()

			if p.cancelListenLoop != nil {
				p.cancelListenLoop()
			}

			p.listenLoopWaitGroup.Wait()

			_ = Conn.Close()
			stopCallback()
		}(p.Conn)

		p.listenLoopWaitGroup.Add(1)

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Millisecond):
				break
			}

			p.connReadMutex.Lock()

			_ = p.Conn.SetReadDeadline(time.Now().Add(time.Second * 10))
			t, msg, err := p.Conn.ReadMessage()

			p.connReadMutex.Unlock()

			if t == websocket.CloseMessage {
				return
			} else if _, ok := err.(*websocket.CloseError); ok {
				return
			}

			if err != nil {
				continue
			}

			if t == websocket.TextMessage {
				p.gameCommandChannel <- string(msg)
			}
		}
	}()

	go func() {
		pingTicker := time.NewTicker(time.Second * 5)

		defer func() {
			pingTicker.Stop()

			if p.cancelListenLoop != nil {
				p.cancelListenLoop()
			}

			p.listenLoopWaitGroup.Done()
		}()

		p.listenLoopWaitGroup.Add(1)

		for {
			select {
			case <-ctx.Done():
				return
			case <-pingTicker.C:
				if err := p.Ping(); err != nil {
					return
				}
				break
			}
		}
	}()
}

func (p *Player) Ping() error {
	defer p.connWriteMutex.Unlock()
	p.connWriteMutex.Lock()

	_ = p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 2))

	if err := p.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		return err
	}

	return nil
}

func (p *Player) ReadMessage() ([]byte, error) {
	defer p.connReadMutex.Unlock()
	p.connReadMutex.Lock()

	_ = p.Conn.SetReadDeadline(time.Now().Add(time.Second * 10))

	_, msg, err := p.Conn.ReadMessage()

	if err != nil {
		log.Println("player read error", err)
		return nil, err
	}

	return msg, nil
}

func (p *Player) SendMessage(data []byte) error {
	defer p.connWriteMutex.Unlock()
	p.connWriteMutex.Lock()

	_ = p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 5))

	err := p.Conn.WriteMessage(websocket.TextMessage, data)

	if err != nil {
		log.Println("player write error", err)
		return err
	}

	return nil
}
