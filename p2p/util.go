package p2p

import (
	"fmt"
	"net"
)

// isPortAvailable returns a flag indicating whether or not a TCP port is available.
func isPortAvailable(port int) bool {
	conn, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
