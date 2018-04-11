package main

import (
	"github.com/iost-official/prototype/cmd"
	_ "github.com/iost-official/prototype/common"
	_ "github.com/iost-official/prototype/core"
	_ "github.com/iost-official/prototype/event"
	_ "github.com/iost-official/prototype/iostdb"
	_ "github.com/iost-official/prototype/p2p"
)

func main() {
	cmd.Execute()
}
