package mclock

import (
	"github.com/aristanetworks/goarista/monotime"
	"time"
)

// AbsTime absolute monotonic time
type AbsTime time.Duration

// Now returns a AbsTime instance
func Now() AbsTime {
	return AbsTime(monotime.Now())
}
