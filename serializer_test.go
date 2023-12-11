package eventbus_test

import (
	eventbus "github.com/optimus-hft/event-bus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSerializer_Execute(t *testing.T) {
	serializer := eventbus.NewSerializer(0)
	b1 := false
	b2 := false
	b3 := false
	b4 := false
	b5 := false

	serializer.Execute(func() {
		b1 = true
	}, 1)
	assert.Equal(t, true, b1)
	assert.Equal(t, false, b2)
	assert.Equal(t, false, b3)
	assert.Equal(t, false, b4)
	assert.Equal(t, false, b5)

	serializer.Execute(func() {
		b3 = true
	}, 3)
	assert.Equal(t, true, b1)
	assert.Equal(t, false, b2)
	assert.Equal(t, false, b3)
	assert.Equal(t, false, b4)
	assert.Equal(t, false, b5)

	serializer.Execute(func() {
		b4 = true
	}, 3)
	assert.Equal(t, true, b1)
	assert.Equal(t, false, b2)
	assert.Equal(t, false, b3)
	assert.Equal(t, false, b4)
	assert.Equal(t, false, b5)

	serializer.Execute(func() {
		b5 = true
	}, 4)
	assert.Equal(t, true, b1)
	assert.Equal(t, false, b2)
	assert.Equal(t, false, b3)
	assert.Equal(t, false, b4)
	assert.Equal(t, false, b5)

	serializer.Execute(func() {
		b2 = true
	}, 2)
	assert.Equal(t, true, b1)
	assert.Equal(t, true, b2)
	assert.Equal(t, false, b3)
	assert.Equal(t, true, b4)
	assert.Equal(t, true, b5)
}
