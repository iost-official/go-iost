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
		bc, to_print := transaction.CreateBlockchain(args[0], Db, Nn)

		if bc == nil {
			return to_print
		}

		//defer bc.Db.Close()

		to_print += "Done!\n"
		return to_print
	}
	return c
}
