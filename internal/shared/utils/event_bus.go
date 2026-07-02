package utils

import (
	"context"
	"sync"
)

type Event struct {
	Name    string
	Payload interface{}
}

type EventHandler func(ctx context.Context, event Event)

type EventBus struct {
	subscribers map[string][]EventHandler
	mu          sync.RWMutex
	wg          sync.WaitGroup
}

type EventBusInterface interface {
	Subscribe(topic string, handler EventHandler)
	Publish(ctx context.Context, event Event)
	Wait()
}

func NewEventBus() EventBusInterface {
	return &EventBus{
		subscribers: make(map[string][]EventHandler),
	}
}

func (bus *EventBus) Subscribe(topic string, handler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.subscribers[topic] = append(bus.subscribers[topic], handler)
}

func (bus *EventBus) Publish(ctx context.Context, event Event) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	if handlers, found := bus.subscribers[event.Name]; found {
		for _, handler := range handlers {
			bus.wg.Add(1)

			go func(h EventHandler, e Event) {
				defer bus.wg.Done()
				h(ctx, e)
			}(handler, event)
		}
	}
}

func (bus *EventBus) Wait() {
	bus.wg.Wait()
}
