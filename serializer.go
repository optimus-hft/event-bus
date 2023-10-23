package event_bus

import (
	"slices"
	"sync"
)

type callback func()

type queueItem struct {
	callback func()
	sequence uint64
}

type Serializer struct {
	mu                    sync.Mutex
	queue                 []queueItem
	sequenceToExecuteNext uint64
}

func (s *Serializer) Enqueue(cb callback, sequence uint64) {
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

		if item.sequence != s.sequenceToExecuteNext {
			break
		}

		s.queue = s.queue[1:]
		item.callback()
		s.sequenceToExecuteNext++
	}
}

func NewSerializer() *Serializer {
	return &Serializer{
		queue: make([]queueItem, 0),
	}
}
