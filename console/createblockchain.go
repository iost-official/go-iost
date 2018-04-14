package console

import "github.com/iost-official/prototype/tx/tx"

func Createblockchain() Cmd {
	c := Cmd{
		Name:  "createblockchain",
		Usage: `createblockchain ADDRESS - Create a blockchain and send genesis block reward to ADDRESS`,
	}
	c.Exec = func(host *Console, args []string) string {
		if len(args) != 1 {
			return "Invalid arguments!\n"
		}
		bc, toPrint := transaction.CreateBlockchain(args[0])

		if bc == nil {
			return toPrint
		}

		defer bc.Db.Close()

		toPrint += "Done!\n"
		return toPrint
	}
	return c
}
