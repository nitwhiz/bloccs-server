package event

import (
	"context"
	"log"
	"sync"
)

const All = "*"

type Handler func(event *Event)

type Bus struct {
	nextHandlerId  int
	handlers       map[Type]map[int]Handler
	waitGroup      *sync.WaitGroup
	eventWaitGroup *sync.WaitGroup
	channel        chan *Event
	ctx            context.Context
	stopFunc       context.CancelFunc
	closeAfter     int
	mu             *sync.RWMutex
}

func NewBus() *Bus {
	b := Bus{
		nextHandlerId:  0,
		handlers:       map[Type]map[int]Handler{},
		waitGroup:      &sync.WaitGroup{},
		eventWaitGroup: &sync.WaitGroup{},
		channel:        make(chan *Event, 100),
		closeAfter:     -1,
		mu:             &sync.RWMutex{},
	}

	b.ctx, b.stopFunc = context.WithCancel(context.Background())

	return &b
}

func (b *Bus) Stop() {
	b.eventWaitGroup.Wait()

	if b.stopFunc != nil {
		b.stopFunc()
	} else {
		log.Println("no stop func for bus, stopping may never finish")
	}

	b.waitGroup.Wait()
}

func (b *Bus) Subscribe(eventType Type, handler Handler) int {
	defer b.mu.Unlock()
	b.mu.Lock()

	if _, ok := b.handlers[eventType]; !ok {
		b.handlers[eventType] = map[int]Handler{}
	}

	handlerId := b.nextHandlerId

	b.handlers[eventType][handlerId] = handler

	b.nextHandlerId += 1

	return handlerId
}

func (b *Bus) Unsubscribe(handlerId int) {
	defer b.mu.Unlock()
	b.mu.Lock()

	for eventType := range b.handlers {
		if _, ok := b.handlers[eventType][handlerId]; ok {
			delete(b.handlers[eventType], handlerId)
			break
		}
	}
}

func (b *Bus) Publish(event *Event) {
	b.eventWaitGroup.Add(1)
	b.channel <- event
}

func (b *Bus) handleEvent(event *Event) {
	defer b.mu.RUnlock()
	b.mu.RLock()

	hs, ok := b.handlers[event.Type]

	if ok {
		for _, h := range hs {
			h(event)
		}
	}

	hs, ok = b.handlers[All]

	if ok {
		for _, h := range hs {
			h(event)
		}
	}
}

func (b *Bus) Start() {
	go func() {
		defer b.waitGroup.Done()
		b.waitGroup.Add(1)

		for {
			select {
			case <-b.ctx.Done():
				return
			case event := <-b.channel:
				b.handleEvent(event)
				b.eventWaitGroup.Done()
				break
			}
		}
	}()
}
