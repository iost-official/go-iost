package main

import (
	"github.com/iost-official/PrototypeWorks/cmd"
	_ "github.com/iost-official/PrototypeWorks/drp"
	_ "github.com/iost-official/PrototypeWorks/eds"
	_ "github.com/iost-official/PrototypeWorks/iosbase"
	_ "github.com/iost-official/PrototypeWorks/iosbase/debug"
	_ "github.com/iost-official/PrototypeWorks/iosbase/mocks"
	_ "github.com/iost-official/PrototypeWorks/mock-libs/asset"
	_ "github.com/iost-official/PrototypeWorks/mock-libs/block"
	_ "github.com/iost-official/PrototypeWorks/mock-libs/market"
	_ "github.com/iost-official/PrototypeWorks/mock-libs/transaction"
	_ "github.com/iost-official/PrototypeWorks/protocol"
	_ "github.com/iost-official/PrototypeWorks/protocol/mocks"
)

func main() {
	cmd.Execute()
}
