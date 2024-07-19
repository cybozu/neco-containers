package main

import (
	"sync"
	"time"
)

type Queue struct {
	queue []Machine
	mu    *sync.Mutex
}

// Get queue
func (q *Queue) get() Machine {
	var m Machine
	for {
		q.mu.Lock()
		l := len(q.queue)
		q.mu.Unlock()

		if l == 0 {
			time.Sleep(1 * time.Second)
		} else {
			q.mu.Lock()
			m = q.queue[0]
			q.queue = q.queue[1:]
			q.mu.Unlock()
			break
		}
	}
	return m
}

// Put queue
func (q *Queue) put(m []Machine) {
	q.mu.Lock()
	for i := 0; i < len(m); i++ {
		q.queue = append(q.queue, m[i])
	}
	q.mu.Unlock()
}

// Put queue
func (q *Queue) len() int {
	return len(q.queue)
}
