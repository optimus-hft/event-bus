package event_bus

import (
	"context"
	"sync"
)

type subscription[T any] struct {
	serializer *Serializer
	channel    chan T
	once       bool
}

// Bus can be used to implement event-driven architectures in Golang.
type Bus[T any] struct {
	mu               sync.Mutex
	subscriptions    map[string][]subscription[T]
	publishSequences map[string]uint64
	chanBufferSize   int
}

func (b *Bus[T]) unsubscribe(topic string, sub subscription[T]) {
	if _, ok := b.subscriptions[topic]; !ok {
		return
	}

	subIndex := FindSliceIndex(b.subscriptions[topic], sub)
	if subIndex <= -1 {
		return
	}

	close(sub.channel)

	b.subscriptions[topic] = RemoveFromSlice(b.subscriptions[topic], subIndex)
	if len(b.subscriptions[topic]) == 0 {
		delete(b.subscriptions, topic)
		delete(b.publishSequences, topic)
	}
}

func (b *Bus[T]) subscribe(topic string, once bool) (<-chan T, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subscriptions[topic]; !ok {
		b.subscriptions[topic] = make([]subscription[T], 0)
	}

	sub := subscription[T]{
		serializer: nil,
		channel:    make(chan T, b.chanBufferSize),
		once:       once,
	}
	// once subscriptions doesn't need serializer
	if !once {
		sub.serializer = NewSerializer()
	}

	b.subscriptions[topic] = append(b.subscriptions[topic], sub)

	return sub.channel, func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		b.unsubscribe(topic, sub)
	}
}

func (b *Bus[T]) Subscribe(topic string) (<-chan T, func()) {
	return b.subscribe(topic, false)
}

func (b *Bus[T]) on(ctx context.Context, topic string, callback func(event T), once bool) func() {
	channel, unsubscribe := b.subscribe(topic, once)

	go func() {
		for {
			select {
			case event := <-channel:
				callback(event)

				if once {
					return
				}
			case <-ctx.Done():
				unsubscribe()

				return
			}
		}
	}()

	return unsubscribe
}

func (b *Bus[T]) On(ctx context.Context, topic string, callback func(event T)) func() {
	return b.on(ctx, topic, callback, false)
}

func (b *Bus[T]) Once(ctx context.Context, topic string, callback func(event T)) func() {
	return b.on(ctx, topic, callback, true)
}

func (b *Bus[T]) SubscribeOnce(topic string) (<-chan T, func()) {
	return b.subscribe(topic, true)
}

func (b *Bus[T]) Publish(topic string, event T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subscriptions[topic]; !ok || len(b.subscriptions[topic]) == 0 {
		return
	}

	sequence := b.publishSequences[topic]
	b.publishSequences[topic]++

	for _, sub := range b.subscriptions[topic] {
		if sub.once {
			b.unsubscribe(topic, sub)
		}

		go func(sub subscription[T]) {
			if sub.serializer == nil {
				sub.channel <- event

				return
			}

			sub.serializer.Enqueue(func() {
				sub.channel <- event
			}, sequence)
		}(sub)
	}
}

func NewBus[T any](chanBufferSize int) *Bus[T] {
	return &Bus[T]{
		subscriptions:    make(map[string][]subscription[T]),
		publishSequences: make(map[string]uint64),
		chanBufferSize:   chanBufferSize,
	}
}
