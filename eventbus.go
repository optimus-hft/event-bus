package eventbus

import (
	"slices"
	"sync"
)

type subscription[T any] struct {
	serializer *Serializer
	channel    chan T
	once       bool
}

// Bus can be used to implement event-driven architectures in Golang. Each bus can have multiple subscribers to different topics.
type Bus[T any] struct {
	mu               sync.Mutex
	subscriptions    map[string][]subscription[T]
	publishSequences map[string]uint64
}

func (b *Bus[T]) unsubscribe(topic string, sub subscription[T]) {
	if _, ok := b.subscriptions[topic]; !ok {
		return
	}

	subIndex := slices.Index(b.subscriptions[topic], sub)
	if subIndex <= -1 {
		return
	}

	b.subscriptions[topic] = slices.Delete(b.subscriptions[topic], subIndex, subIndex+1)
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
		channel:    make(chan T),
		once:       once,
	}
	// once subscriptions don't need serializer
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

// Subscribe returns a channel for listening on a given topic events, Also an unsubscribe functions is returned that can be used to cancel the subscription.
func (b *Bus[T]) Subscribe(topic string) (<-chan T, func()) {
	return b.subscribe(topic, false)
}

// SubscribeOnce returns a channel for listening on a given topic events, After the first event, subscription is automatically cancelled.
// Also, an unsubscribe functions is returned that can be used to cancel the subscription before receiving any events.
func (b *Bus[T]) SubscribeOnce(topic string) (<-chan T, func()) {
	return b.subscribe(topic, true)
}

func (b *Bus[T]) on(topic string, callback func(event T), once bool) func() {
	channel, unsubscribe := b.subscribe(topic, once)
	doneChannel := make(chan struct{})

	go func() {
		for {
			select {
			case event := <-channel:
				callback(event)

				if once {
					return
				}
			case <-doneChannel:
				return
			}
		}
	}()

	return func() {
		unsubscribe()
		close(doneChannel)
	}
}

// On executes the given callback whenever an event is sent to the given topic, Also an unsubscribe functions is returned that can be used to cancel the subscription.
func (b *Bus[T]) On(topic string, callback func(event T)) func() {
	return b.on(topic, callback, false)
}

// Once executes the given callback as soon as receiving the first event on the given topic. After one callback execution, Callback will not be called again.
// Also, an unsubscribe functions is returned that can be used to cancel the subscription before receiving any events.
func (b *Bus[T]) Once(topic string, callback func(event T)) func() {
	return b.on(topic, callback, true)
}

// Publish sends the given event to all the subscribers of the given topic. Publish is non-blocking and doesn't wait for subscribers.
// Although this is non-blocking, event ordering is guaranteed in a FIFO manner. Each subscriber gets events in the same order they were published.
// A slow subscriber doesn't block other subscribers, Ordering is handled for each subscriber separately and different subscribers on the same topic can be reading different events at a given time.
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

			sub.serializer.Execute(func() {
				sub.channel <- event
			}, sequence)
		}(sub)
	}
}

// New creates a new event bus.
func New[T any]() *Bus[T] {
	return &Bus[T]{
		subscriptions:    make(map[string][]subscription[T]),
		publishSequences: make(map[string]uint64),
	}
}
