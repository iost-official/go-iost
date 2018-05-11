package prque

import "container/heap"

type Prque struct {
	cont *sstack
}

func New() *Prque {
	return &Prque{newSstack()}
}

func (p *Prque) Push(data interface{}, priority float32) {
	heap.Push(p.cont, &item{data, priority})
}

func (p *Prque) Pop() (interface{}, float32) {
	item := heap.Pop(p.cont).(*item)
	return item.value, item.priority
}

func (p *Prque) PopItem() interface{} {
	return heap.Pop(p.cont).(*item).value
}

func (p *Prque) Empty() bool {
	return p.cont.Len() == 0
}

func (p *Prque) Size() int {
	return p.cont.Len()
}

func (p *Prque) Reset() {
	*p = *New()
}