package console

import (
	"github.com/iost-official/prototype/tx/tx"
	"strconv"
)

func Send() Cmd {
	return Cmd{
		Name:  "send",
		Usage: `send FROM TO AMOUNT - Send AMOUNT of coins from FROM address to TO`,
		Exec: func(host *Console, args []string) string {
			if len(args) != 3 {
				return "Invalid arguments!\n"
			}
			from := args[0]
			to := args[1]
			amount, err := strconv.Atoi(args[2])
			if err != nil {
				return "Invalid arguments!\n"
			}

			bc, toPrint := transaction.NewBlockchain(from)

			if bc == nil {
				return toPrint
			}

			defer bc.Db.Close()

			tx, toPrint := transaction.NewUTXOTransaction(from, to, amount, bc)

			if tx == nil {
				return toPrint
			}

			bc.MineBlock([]*transaction.Transaction{tx})
			toPrint += "\nSuccess!\n"
			return toPrint
		},
	}
}
