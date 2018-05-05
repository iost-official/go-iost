package console

import (
	"fmt"
	"github.com/iost-official/prototype/tx/tx"
)

func Printchain() Cmd {
	return Cmd{
		Name:  "printchain",
		Usage: `printchain - Print all the blocks of the blockchain`,
		Exec: func(host *Console, args []string) string {
			bc, to_print := transaction.NewBlockchain("", Db)

			if bc == nil {
				return to_print
			}

			//defer bc.Db.Close()

			bci := bc.Iterator()

			for {
				block := bci.Next()

				to_print += fmt.Sprintf("Prev hash: %x\n", block.PrevBlockHash)
				to_print += fmt.Sprintf("Hash: %x\n", block.Hash)
				for _, tx := range block.Transactions {
					to_print += tx.String()
				}
				to_print += "\n"

				if len(block.PrevBlockHash) == 0 {
					break
				}
			}
			return to_print
		},
	}
}
