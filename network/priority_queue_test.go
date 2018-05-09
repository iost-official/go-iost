package network

import (
	"testing"
	"time"

	"math/rand"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPriorityQueue_Pop(t *testing.T) {
	q := newQueue()

	if q.Len() != 0 {
		t.Fatal()
	}

	Convey("Request push pop test", t, func() {
		q.Push(Request{Timestamp: time.Now().UnixNano(), Priority: 3})
		q.Push(Request{Timestamp: time.Now().UnixNano(), Priority: 1})
		q.Push(Request{Timestamp: time.Now().UnixNano(), Priority: 2})
		p1 := q.Pop()
		So(p1.(Request).Priority, ShouldEqual, int8(1))
		p2 := q.Pop()
		So(p2.(Request).Priority, ShouldEqual, int8(2))
		p3 := q.Pop()
		So(p3.(Request).Priority, ShouldEqual, int8(3))
		So(q.Len(), ShouldEqual, int8(0))
	})
}
func BenchmarkPriorityQueue_Push(b *testing.B) {
	b.StopTimer()
	q := newQueue()

	if q.Len() != 0 {
		b.Fatal()
	}
	b.StartTimer()
	for i := 0; i < 100000; i++ {
		q.Push(Request{Timestamp: time.Now().UnixNano(), Priority: int8(rand.Intn(100))})
	}
}

func BenchmarkPriorityQueue_Pop(b *testing.B) {
	b.StopTimer()
	q := newQueue()

	if q.Len() != 0 {
		b.Fatal()
	}
	for i := 0; i < 100000; i++ {
		q.Push(Request{Timestamp: time.Now().UnixNano(), Priority: int8(rand.Intn(100))})
	}
	b.StartTimer()
	for q.Len() > 0 {
		q.Pop()
	}
}
