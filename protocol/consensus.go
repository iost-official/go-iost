package protocol

import (
	"sync"
	"time"

	"github.com/iost-official/PrototypeWorks/iosbase"
)

type Character int

const (
	Primary Character = iota
	Backup
	Idle
)

const (
	Version    = 1
	Port       = 12306
)

type Consensus struct {
	recorder, replica Component
	db                Database
	router            Router
}

