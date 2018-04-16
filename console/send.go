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

			bc, to_print := transaction.NewBlockchain(from, Db)

			if bc == nil {
				return to_print
			}

			//defer bc.Db.Close()

			tx, to_print := transaction.NewUTXOTransaction(from, to, amount, bc)

			if tx == nil {
				return to_print
			}

			bc.MineBlock([]*transaction.Transaction{tx}, Nn)
			to_print += "\nSuccess!\n"
			return to_print
		},
	}
}
