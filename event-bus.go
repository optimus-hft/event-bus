package event_bus

import (
	"context"
	"sync"
)

type Bus[T any] struct {
	mu                     sync.RWMutex
	subscribedChannels     map[string][]chan T
	onceSubscribedChannels map[string][]chan T
	chanBufferSize         int
}

func (b *Bus[T]) unsubscribe(channelsMap map[string][]chan T, topic string, channel chan T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := channelsMap[topic]; !ok {
		return
	}

	chanIndex := FindSliceIndex(channelsMap[topic], channel)
	if chanIndex <= -1 {
		return
	}

	close(channel)

	channelsMap[topic] = RemoveFromSlice(channelsMap[topic], chanIndex)
	if len(channelsMap[topic]) == 0 {
		delete(channelsMap, topic)
	}
}

func (b *Bus[T]) subscribe(topic string, once bool) (chan T, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	channelsMap := b.subscribedChannels
	if once {
		channelsMap = b.onceSubscribedChannels
	}

	if _, ok := channelsMap[topic]; !ok {
		channelsMap[topic] = make([]chan T, 0)
	}

	channel := make(chan T, b.chanBufferSize)
	channelsMap[topic] = append(channelsMap[topic], channel)

	return channel, func() {
		b.unsubscribe(channelsMap, topic, channel)
	}
}

func (b *Bus[T]) Subscribe(topic string) (chan T, func()) {
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

func (b *Bus[T]) SubscribeOnce(topic string) (chan T, func()) {
	return b.subscribe(topic, true)
}

func (b *Bus[T]) Publish(topic string, event T) {
	b.mu.RLock()
	channels := b.subscribedChannels[topic]
	onceChannels := b.subscribedChannels[topic]
	b.mu.RUnlock()

	if len(channels) == 0 && len(onceChannels) == 0 {
		return
	}

	for _, channel := range channels {
		go func(channel chan T) {
			channel <- event
		}(channel)
	}

	for _, channel := range onceChannels {
		b.unsubscribe(b.onceSubscribedChannels, topic, channel)
		go func(channel chan T) {
			channel <- event
		}(channel)
	}
}

func NewBus[T any](chanBufferSize int) *Bus[T] {
	return &Bus[T]{
		subscribedChannels: make(map[string][]chan T),
		chanBufferSize:     chanBufferSize,
	}
}
