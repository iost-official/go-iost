package main

import (
	"github.com/iost-official/Go-IOS-Protocol/cmd"
	_ "github.com/iost-official/Go-IOS-Protocol/drp"
	_ "github.com/iost-official/Go-IOS-Protocol/eds"
	_ "github.com/iost-official/Go-IOS-Protocol/iosbase"
	_ "github.com/iost-official/Go-IOS-Protocol/iosbase/debug"
	_ "github.com/iost-official/Go-IOS-Protocol/iosbase/mocks"
	_ "github.com/iost-official/Go-IOS-Protocol/mock-libs/asset"
	_ "github.com/iost-official/Go-IOS-Protocol/mock-libs/block"
	_ "github.com/iost-official/Go-IOS-Protocol/mock-libs/market"
	_ "github.com/iost-official/Go-IOS-Protocol/mock-libs/transaction"
	_ "github.com/iost-official/Go-IOS-Protocol/protocol"
	_ "github.com/iost-official/Go-IOS-Protocol/protocol/mocks"
)

func main() {
	cmd.Execute()
}
