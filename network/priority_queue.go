package network

import (
	"container/heap"
	"sync"
)

type LessInterface interface {
	Less(target interface{}) bool
}

type sorter []LessInterface

func (s *sorter) Len() int { return len(*s) }

func (s *sorter) Less(i, j int) bool { return (*s)[i].Less((*s)[j]) }

func (s *sorter) Push(x interface{}) { *s = append(*s, x.(LessInterface)) }

func (s *sorter) Swap(i, j int) { (*s)[i], (*s)[j] = (*s)[j], (*s)[i] }

func (s *sorter) Pop() interface{} {
	n := len(*s)
	if n > 0 {
		x := (*s)[n-1]
		*s = (*s)[0 : n-1]
		return x
	}
	return nil
}

type PriorityQueue struct {
	s    *sorter
	lock sync.RWMutex
}

func newQueue() *PriorityQueue {
	q := &PriorityQueue{s: new(sorter)}
	heap.Init(q.s)
	return q
}

func (q *PriorityQueue) Len() int {
	return q.s.Len()
}

func (q *PriorityQueue) Push(x Request) {
	q.lock.Lock()
	defer q.lock.Unlock()
	heap.Push(q.s, x)
}

func (q *PriorityQueue) Pop() LessInterface {
	q.lock.Lock()
	defer q.lock.Unlock()
	return heap.Pop(q.s).(LessInterface)
}
