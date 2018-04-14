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
			bc, toPrint := transaction.NewBlockchain("")

			if bc == nil {
				return toPrint
			}

			defer bc.Db.Close()

			bci := bc.Iterator()

			for {
				block := bci.Next()

				toPrint += fmt.Sprintf("Prev hash: %x\n", block.PrevBlockHash)
				toPrint += fmt.Sprintf("Hash: %x\n", block.Hash)
				for _, tx := range block.Transactions {
					toPrint += tx.String()
				}
				toPrint += "\n"

				if len(block.PrevBlockHash) == 0 {
					break
				}
			}
			return toPrint
		},
	}
}
