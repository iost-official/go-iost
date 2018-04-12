package console

import (
	"fmt"
	"strings"
	"github.com/iost-official/prototype/tx/tx"
	"bufio"
	"os"
	"strconv"
)

func Listen() {
	for {
		var cmd string
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		cmd, _ = reader.ReadString('\n')
		args := strings.Fields(cmd)
		if len(args) == 0 {
			continue
		}
		switch strings.ToLower(args[0]) {
		case "exit":
			fmt.Println("bye!")
			return
		default:
			fmt.Print(Run(args[0], args[1:]))
		}
	}
}

var cmds map[string]Cmd

func RegistCmd(cmd Cmd) {
	cmds[cmd.Name] = cmd
}

func Run(name string, args []string) string {
	name = strings.ToLower(name)
	if cmd, ok := cmds[name]; ok {
		return cmd.Exec(args)
	} else {
		return "Command not found\n"
	}
}

type Cmd struct {
	Name  string
	Usage string
	Exec  func(args []string) string
}

func init() {
	cmds = make(map[string]Cmd)

	help := Cmd{
		Name:  "help",
		Usage: `Print this manifest`,
		Exec: func(args []string) string {
			s := "======COMMAND LIST\n"
			for k, v := range cmds {
				s += fmt.Sprintf("%v :\n\t%v\n", k, v.Usage)
			}

			s += "exit :\n\tStop daemon and quit\n"
			return s
		},
	}
	RegistCmd(help)

	printchain := Cmd{
		Name:  "printchain",
		Usage: `printchain - Print all the blocks of the blockchain`,
		Exec: func(args []string) string {
			bc, to_print := transaction.NewBlockchain("")

			if bc == nil {
				return to_print
			}

			defer bc.Db.Close()

			bci := bc.Iterator()

			for {
				block := bci.Next()

				to_print += fmt.Sprintf("Prev hash: %x\n", block.PrevBlockHash)
				to_print += fmt.Sprintf("Hash: %x\n", block.Hash)
				for _, tx := range(block.Transactions) {
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
	RegistCmd(printchain)

	createblockchain := Cmd{
		Name:  "createblockchain",
		Usage: `createblockchain ADDRESS - Create a blockchain and send genesis block reward to ADDRESS`,
		Exec: func(args []string) string {
			if len(args) != 1 {
				return "Invalid arguments!\n"
			}
			bc, to_print := transaction.CreateBlockchain(args[0])

			if bc == nil {
				return to_print
			}

			defer bc.Db.Close()

			to_print += "Done!\n"
			return to_print
		},
	}
	RegistCmd(createblockchain)

	getbalance := Cmd{
		Name:  "getbalance",
		Usage: `getbalance ADDRESS - Get balance of ADDRESS`,
		Exec: func(args []string) string {
			if len(args) != 1 {
				return "Invalid arguments!\n"
			}
			address := args[0]
			bc, to_print := transaction.NewBlockchain(address)

			if bc == nil {
				return to_print
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
	RegistCmd(getbalance)

	send := Cmd{
		Name:  "send",
		Usage: `send FROM TO AMOUNT - Send AMOUNT of coins from FROM address to TO`,
		Exec: func(args []string) string {
			if len(args) != 3 {
				return "Invalid arguments!\n"
			}
			from := args[0]
			to := args[1]
			amount, err := strconv.Atoi(args[2])
			if err != nil {
				return "Invalid arguments!\n"
			}

			bc, to_print := transaction.NewBlockchain(from)

			if bc == nil {
				return to_print
			}

			defer bc.Db.Close()

			tx, to_print := transaction.NewUTXOTransaction(from, to, amount, bc)

			if tx == nil {
				return to_print
			}

			bc.MineBlock([]*transaction.Transaction{tx})
			to_print += "\nSuccess!\n"
			return to_print
		},
	}
	RegistCmd(send)
}
