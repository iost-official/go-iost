package mclock
import (
	"time"
	"github.com/aristanetworks/goarista/monotime"
)

type AbsTime time.Duration // absolute monotonic time

func Now() AbsTime {
	return AbsTime(monotime.Now())
}
