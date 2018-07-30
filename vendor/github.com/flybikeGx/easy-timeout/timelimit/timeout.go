package timelimit

import (
	"time"
)

func Run(limit time.Duration, f func()) bool {
	var expire, done chan bool
	expire = make(chan bool)
	done = make(chan bool)

	go func() {
		f()
		select {
		case <-expire:
			close(done)
			return
		default:
			done <- true
			close(done)
		}
	}()

	select {
	case <-time.After(limit):
		close(expire)
		return false
	case <-done:
		return true
	}
}
