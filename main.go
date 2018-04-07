package main

import (
	"github.com/iost-official/PrototypeWorks/cmd"
	_ "github.com/iost-official/PrototypeWorks/debug"
	_ "github.com/iost-official/PrototypeWorks/drp"
	_ "github.com/iost-official/PrototypeWorks/eds"
	_ "github.com/iost-official/PrototypeWorks/iosbase"
	_ "github.com/iost-official/PrototypeWorks/mock-libs"
	_ "github.com/iost-official/PrototypeWorks/protocol"
)

func main() {
	cmd.Execute()
}
