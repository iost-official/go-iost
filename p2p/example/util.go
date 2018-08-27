package main

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

type colorNum int

const (
	grey   colorNum = 30
	green  colorNum = 32
	yellow colorNum = 33
	blue   colorNum = 34
)

func isPortAvailable(port int) bool {
	conn, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func randomPort() int {
	for i := 0; i < 10; i++ {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		p := r.Intn(64511) + 1025
		if isPortAvailable(p) {
			return p
		}
	}
	return -1
}

func color(text string, n colorNum) string {
	return fmt.Sprintf("\033[1;%dm%s\033[0m", n, text)
}

func shortID(id string) string {
	return id[len(id)-6:]
}
