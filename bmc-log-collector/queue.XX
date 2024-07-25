package main

import (
// "fmt"
)

type MessageQueue struct {
	queue chan Machine
}

// Get queue
func (q *MessageQueue) get2() Machine {
	return <-q.queue
}

// Put
//func (q *MessageQueue) put2(m Machine) {
//	q.queue <- m
//}

// Put
func (q *MessageQueue) put3(m []Machine) {
	for i := 0; i < len(m); i++ {
		q.queue <- m[i]
	}
	//q.queue <- m
}

func (q *MessageQueue) len2() int {
	return len(q.queue)

}
