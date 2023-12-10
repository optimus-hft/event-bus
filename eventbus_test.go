package eventbus

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	topicA = "tA"
	topicB = "tB"
)

var (
	eventsA = []int{0, 1, 2, 3, 4, 5}
	eventsB = []int{100, 101, 102, 103, 104, 105}
)

func TestBus_Subscribe(t *testing.T) {
	bus := New[int]()

	cA1, uA1 := bus.Subscribe(topicA)
	cA2, _ := bus.Subscribe(topicA)
	cB1, _ := bus.Subscribe(topicB)
	a1Counter, a2Counter, b1Counter, b2Counter := 0, 0, 0, 0

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	go func() {
		time.Sleep(500 * time.Millisecond) // simulates reader goroutine being blocked to make sure serializer is working correctly.

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-cA1:
				assert.Equal(t, eventsA[a1Counter], event)
				a1Counter++
			case event := <-cA2:
				assert.Equal(t, eventsA[a2Counter], event)
				a2Counter++
			case event := <-cB1:
				assert.Equal(t, eventsB[b1Counter], event)
				b1Counter++
			}
		}
	}()

	bus.Publish(topicA, eventsA[0])
	uA1()
	bus.Publish(topicA, eventsA[1])
	bus.Publish(topicB, eventsB[0])

	cB2, _ := bus.Subscribe(topicB)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-cB2:
				b2Counter++
			}
		}
	}()
	bus.Publish(topicB, eventsB[1])

	time.Sleep(1 * time.Second)

	assert.Equal(t, 1, a1Counter)
	assert.Equal(t, 2, a2Counter)
	assert.Equal(t, 2, b1Counter)
	assert.Equal(t, 1, b2Counter)

	assert.Equal(t, 2, len(bus.subscriptions))
	assert.Equal(t, 2, len(bus.publishSequences))
}

func TestBus_SubscribeOnce(t *testing.T) {
	bus := New[int]()

	cA1, uA1 := bus.SubscribeOnce(topicA)
	cA2, uA2 := bus.SubscribeOnce(topicA)
	cB1, _ := bus.SubscribeOnce(topicB)
	a1Counter, a2Counter, b1Counter := 0, 0, 0

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-cA1:
				assert.Equal(t, eventsA[a1Counter], event)
				a1Counter++
			case event := <-cA2:
				assert.Equal(t, eventsA[a2Counter], event)
				a2Counter++
			case event := <-cB1:
				assert.Equal(t, eventsB[b1Counter], event)
				b1Counter++
			}
		}
	}()

	uA1()
	uA1() // calling unsubscribe should be idempotent
	bus.Publish(topicA, eventsA[0])
	bus.Publish(topicA, eventsA[1])
	bus.Publish(topicB, eventsB[0])

	time.Sleep(1 * time.Second)

	uA2() // calling unsubscribe after an event for once should do nothing

	assert.Equal(t, 0, a1Counter)
	assert.Equal(t, 1, a2Counter)
	assert.Equal(t, 1, b1Counter)

	assert.Equal(t, 0, len(bus.subscriptions))
	assert.Equal(t, 0, len(bus.publishSequences))
}

func TestBus_On(t *testing.T) {
	bus := New[int]()

	a1Counter, a2Counter, b1Counter := 0, 0, 0

	uA1 := bus.On(topicA, func(event int) {
		assert.Equal(t, eventsA[a1Counter], event)
		a1Counter++
	})

	bus.On(topicA, func(event int) {
		assert.Equal(t, eventsA[a2Counter], event)
		a2Counter++
	})

	bus.On(topicB, func(event int) {
		assert.Equal(t, eventsB[b1Counter], event)
		b1Counter++
	})

	bus.Publish(topicA, eventsA[0])
	time.Sleep(1 * time.Second)
	uA1()
	bus.Publish(topicA, eventsA[1])
	bus.Publish(topicB, eventsB[0])

	time.Sleep(1 * time.Second)

	assert.Equal(t, 1, a1Counter)
	assert.Equal(t, 2, a2Counter)
	assert.Equal(t, 1, b1Counter)

	assert.Equal(t, 2, len(bus.subscriptions))
	assert.Equal(t, 2, len(bus.publishSequences))
}

func TestBus_Once(t *testing.T) {
	bus := New[int]()

	a1Counter, a2Counter, b1Counter := 0, 0, 0

	uA1 := bus.Once(topicA, func(event int) {
		assert.Equal(t, eventsA[a1Counter], event)
		a1Counter++
	})

	bus.Once(topicA, func(event int) {
		assert.Equal(t, eventsA[a2Counter], event)
		a2Counter++
	})

	bus.Once(topicB, func(event int) {
		assert.Equal(t, eventsB[b1Counter], event)
		b1Counter++
	})

	uA1()
	bus.Publish(topicA, eventsA[0])
	bus.Publish(topicA, eventsA[1])
	bus.Publish(topicB, eventsB[0])

	time.Sleep(1 * time.Second)

	assert.Equal(t, 0, a1Counter)
	assert.Equal(t, 1, a2Counter)
	assert.Equal(t, 1, b1Counter)

	assert.Equal(t, 0, len(bus.subscriptions))
	assert.Equal(t, 0, len(bus.publishSequences))
}
