package util

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type Broadcaster[T any] struct {
	subscribers         map[string]chan T
	subscribersAcceptFn map[string]func(event T) bool
}

func (b *Broadcaster[T]) Subscribe(subscriberId string, acceptFn func(event T) bool) <-chan T {
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
	channel, exists := b.subscribers[subscriberId]

	if !exists {
		return
	}

	close(channel)
	delete(b.subscribers, subscriberId)
	delete(b.subscribersAcceptFn, subscriberId)
}

func (b *Broadcaster[T]) Publish(event T, uuid string) {
	log.WithField("uuid", uuid).
		Info("starting to broadcast event")

	var wg sync.WaitGroup
	wg.Add(len(b.subscribers))

	publishFn := func(subscriberId string, channel chan<- T) {
		defer wg.Done()

		acceptFn, exists := b.subscribersAcceptFn[subscriberId]

		if !exists {
			log.WithField("uuid", uuid).
				WithField("subscriber_id", subscriberId).
				Warn("bad configuration found, accept function should not be null for a subscriber")
		}

		if exists && !acceptFn(event) {
			return
		}

		channel <- event
	}

	for id, channel := range b.subscribers {
		go publishFn(id, channel)
	}

	// waiting for all subscribers to receive the event
	wg.Wait()

	log.WithField("uuid", uuid).
		Info("finished broadcasting")
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
