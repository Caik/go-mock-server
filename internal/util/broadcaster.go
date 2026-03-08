package util

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type Broadcaster[T any] struct {
	mu          sync.RWMutex
	subscribers map[string]*subscriber[T]
}

type subscriber[T any] struct {
	ch       chan T
	done     chan struct{}
	wg       sync.WaitGroup
	acceptFn func(event T) bool
}

func (b *Broadcaster[T]) Subscribe(subscriberId string, acceptFn func(event T) bool) <-chan T {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.subscribers == nil {
		b.subscribers = make(map[string]*subscriber[T])
	}

	if sub, exists := b.subscribers[subscriberId]; exists {
		return sub.ch
	}

	sub := &subscriber[T]{
		ch:       make(chan T),
		done:     make(chan struct{}),
		acceptFn: acceptFn,
	}
	b.subscribers[subscriberId] = sub

	return sub.ch
}

func (b *Broadcaster[T]) Unsubscribe(subscriberId string) {
	b.mu.Lock()

	sub, exists := b.subscribers[subscriberId]

	if !exists {
		b.mu.Unlock()
		return
	}

	close(sub.done)
	delete(b.subscribers, subscriberId)
	b.mu.Unlock()

	// Wait for all in-flight publish goroutines targeting this subscriber to finish,
	// then close the data channel so ranging consumers can exit cleanly.
	sub.wg.Wait()
	close(sub.ch)
}

func (b *Broadcaster[T]) Publish(event T, uuid string) {
	log.Info().
		Str("uuid", uuid).
		Msg("starting to broadcast event")

	b.mu.RLock()

	// Copy subscriber refs and increment each WaitGroup while holding RLock.
	// This ensures Unsubscribe (which needs WLock) cannot proceed past its lock
	// acquisition until we have incremented the WaitGroup, preventing a close(ch)
	// from racing with our subsequent send.
	type subRef struct {
		id       string
		ch       chan T
		done     chan struct{}
		wg       *sync.WaitGroup
		acceptFn func(event T) bool
	}

	refs := make([]subRef, 0, len(b.subscribers))
	for id, sub := range b.subscribers {
		sub.wg.Add(1)
		refs = append(refs, subRef{
			id:       id,
			ch:       sub.ch,
			done:     sub.done,
			wg:       &sub.wg,
			acceptFn: sub.acceptFn,
		})
	}

	b.mu.RUnlock()

	var wg sync.WaitGroup
	wg.Add(len(refs))

	for _, ref := range refs {
		go func(r subRef) {
			defer wg.Done()
			defer r.wg.Done()

			if r.acceptFn != nil && !r.acceptFn(event) {
				return
			}

			if r.acceptFn == nil {
				log.Warn().
					Str("uuid", uuid).
					Str("subscriber_id", r.id).
					Msg("bad configuration found, accept function should not be null for a subscriber")
			}

			select {
			case r.ch <- event:
			case <-r.done:
				log.Debug().
					Str("uuid", uuid).
					Str("subscriber_id", r.id).
					Msg("subscriber unsubscribed during publish")
			}
		}(ref)
	}

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
