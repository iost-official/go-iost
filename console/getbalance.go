package console

import (
	"fmt"
	"github.com/iost-official/prototype/tx/tx"
)

func Getbalance() Cmd {
	return Cmd{
		Name:  "getbalance",
		Usage: `getbalance ADDRESS - Get balance of ADDRESS`,
		Exec: func(host *Console, args []string) string {
			if len(args) != 1 {
				return "Invalid arguments!\n"
			}
			address := args[0]
			bc, toPrint := transaction.NewBlockchain(address)

			if bc == nil {
				return toPrint
			}

			defer bc.Db.Close()

			balance := 0
			UTXOs := bc.FindUTXO(address)

			for _, out := range UTXOs {
				balance += out.Value
			}

			return fmt.Sprintf("Balance of '%s': %d\n", address, balance)
		},
	}
}
