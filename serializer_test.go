package eventbus_test

import (
	eventbus "github.com/optimus-hft/event-bus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSerializer_Execute(t *testing.T) {
	serializer := eventbus.NewSerializer()
	b1 := false
	b2 := false
	b3 := false
	b4 := false
	b5 := false

	serializer.Execute(func() {
		b1 = true
	}, 0)
	assert.Equal(t, b1, true)
	assert.Equal(t, b2, false)
	assert.Equal(t, b3, false)
	assert.Equal(t, b4, false)
	assert.Equal(t, b5, false)

	serializer.Execute(func() {
		b3 = true
	}, 2)
	assert.Equal(t, b1, true)
	assert.Equal(t, b2, false)
	assert.Equal(t, b3, false)
	assert.Equal(t, b4, false)
	assert.Equal(t, b5, false)

	serializer.Execute(func() {
		b4 = true
	}, 2)
	assert.Equal(t, b1, true)
	assert.Equal(t, b2, false)
	assert.Equal(t, b3, false)
	assert.Equal(t, b4, false)
	assert.Equal(t, b5, false)

	serializer.Execute(func() {
		b5 = true
	}, 3)
	assert.Equal(t, b1, true)
	assert.Equal(t, b2, false)
	assert.Equal(t, b3, false)
	assert.Equal(t, b4, false)
	assert.Equal(t, b5, false)

	serializer.Execute(func() {
		b2 = true
	}, 1)
	assert.Equal(t, b1, true)
	assert.Equal(t, b2, true)
	assert.Equal(t, b3, true)
	assert.Equal(t, b4, true)
	assert.Equal(t, b5, true)
}
