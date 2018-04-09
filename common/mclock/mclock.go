package mclock

import (
	"github.com/aristanetworks/goarista/monotime"
	"time"
)

type AbsTime time.Duration // absolute monotonic time

func Now() AbsTime {
	return AbsTime(monotime.Now())
}
