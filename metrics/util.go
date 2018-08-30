package metrics

import (
	"net"
	"time"
)

// isAddrAvailable returns a flag indicating whether or not a TCP address is available.
func isAddrAvailable(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
