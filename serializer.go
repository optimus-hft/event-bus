package eventbus

import (
	"slices"
	"sync"
)

type callback func()

type queueItem struct {
	callback func()
	sequence uint64
}

// Serializer can be used to serialize execution of some goroutines according to their sequence number.
// If goroutine A with a higher sequence number gets scheduled before goroutine B with a lower sequence number, A will be queued and will be executed after execution of B.
type Serializer struct {
	mu                   sync.Mutex
	queue                []queueItem
	lastExecutedSequence uint64
}

// Execute runs the given callbacks serially according to each callback sequence number.
func (s *Serializer) Execute(cb callback, sequence uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := queueItem{
		callback: cb,
		sequence: sequence,
	}
	i, _ := slices.BinarySearchFunc(s.queue, item, func(a, b queueItem) int {
		if a.sequence < b.sequence {
			return -1
		} else if a.sequence == b.sequence {
			return 0
		} else {
			return 1
		}
	})
	s.queue = slices.Insert(s.queue, i, item)

	for {
		if len(s.queue) == 0 {
			break
		}

		item := s.queue[0]

		if item.sequence != s.lastExecutedSequence && item.sequence != s.lastExecutedSequence+1 {
			break
		}

		s.queue = s.queue[1:]
		item.callback()
		s.lastExecutedSequence = item.sequence
	}
}

// NewSerializer creates a new serializer.
func NewSerializer() *Serializer {
	return &Serializer{
		queue: make([]queueItem, 0),
	}
}
