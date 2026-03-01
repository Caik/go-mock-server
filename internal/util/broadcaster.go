package util

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type Broadcaster[T any] struct {
	mu                  sync.RWMutex
	subscribers         map[string]chan T
	subscribersAcceptFn map[string]func(event T) bool
}

func (b *Broadcaster[T]) Subscribe(subscriberId string, acceptFn func(event T) bool) <-chan T {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.subscribers == nil {
		b.subscribers = make(map[string]chan T)
		b.subscribersAcceptFn = make(map[string]func(event T) bool)
	}

	if channel, exists := b.subscribers[subscriberId]; exists {
		return channel
	}

	b.subscribers[subscriberId] = make(chan T)
	b.subscribersAcceptFn[subscriberId] = acceptFn

	return b.subscribers[subscriberId]
}

func (b *Broadcaster[T]) Unsubscribe(subscriberId string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	channel, exists := b.subscribers[subscriberId]

	if !exists {
		return
	}

	close(channel)
	delete(b.subscribers, subscriberId)
	delete(b.subscribersAcceptFn, subscriberId)
}

func (b *Broadcaster[T]) Publish(event T, uuid string) {
	log.Info().
		Str("uuid", uuid).
		Msg("starting to broadcast event")

	b.mu.RLock()
	subscribersCopy := make(map[string]chan T, len(b.subscribers))
	acceptFnCopy := make(map[string]func(event T) bool, len(b.subscribersAcceptFn))

	for k, v := range b.subscribers {
		subscribersCopy[k] = v
	}

	for k, v := range b.subscribersAcceptFn {
		acceptFnCopy[k] = v
	}

	b.mu.RUnlock()

	var wg sync.WaitGroup
	wg.Add(len(subscribersCopy))

	publishFn := func(subscriberId string, channel chan<- T) {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				// Channel was closed, subscriber unsubscribed during publish
				log.Debug().
					Str("uuid", uuid).
					Str("subscriber_id", subscriberId).
					Msg("subscriber unsubscribed during publish")
			}
		}()

		acceptFn, exists := acceptFnCopy[subscriberId]

		if !exists {
			log.Warn().
				Str("uuid", uuid).
				Str("subscriber_id", subscriberId).
				Msg("bad configuration found, accept function should not be null for a subscriber")
		}

		if exists && !acceptFn(event) {
			return
		}

		channel <- event
	}

	for id, channel := range subscribersCopy {
		go publishFn(id, channel)
	}

	// waiting for all subscribers to receive the event
	wg.Wait()

	log.Info().
		Str("uuid", uuid).
		Msg("finished broadcasting")
}

func (b *Broadcaster[T]) PublishAsync(event T, uuid string) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		b.Publish(event, uuid)
		wg.Done()
	}()

	return &wg
}
