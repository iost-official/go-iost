package console

import (
	"fmt"
	"strings"
	"github.com/iost-official/prototype/tx/tx"
	"bufio"
	"os"
	"strconv"
	"github.com/iost-official/prototype/p2p"
	"sync"
	"github.com/iost-official/prototype/iostdb"
	"github.com/iost-official/prototype/tx/min_framework"
	"io/ioutil"
)

var wg sync.WaitGroup
var done = make(chan struct{})
var nn *p2p.NaiveNetwork
var db *iostdb.LDBDatabase

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
			close(done)
			wg.Wait()
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
			bc, to_print := transaction.NewBlockchain("", db)

			if bc == nil {
				return to_print
			}

			//defer bc.Db.Close()

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
			bc, to_print := transaction.CreateBlockchain(args[0], db, nn)

			if bc == nil {
				return to_print
			}

			//defer bc.Db.Close()

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
			bc, to_print := transaction.NewBlockchain(address, db)

			if bc == nil {
				return to_print
			}

			//defer bc.Db.Close()

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

			bc, to_print := transaction.NewBlockchain(from, db)

			if bc == nil {
				return to_print
			}

			//defer bc.Db.Close()

			tx, to_print := transaction.NewUTXOTransaction(from, to, amount, bc)

			if tx == nil {
				return to_print
			}

			bc.MineBlock([]*transaction.Transaction{tx}, nn)
			to_print += "\nSuccess!\n"
			return to_print
		},
	}
	RegistCmd(send)

	connect := Cmd{
		Name:  "connect",
		Usage: `connect PORT - Connect to the network. Listen to PORT`,
		Exec: func(args []string) string {
			if len(args) != 1 {
				return "Invalid arguments!\n"
			}
			port, err := strconv.Atoi(args[0])
			if err != nil {
				return "Invalid arguments!\n"
			}

			dirname, _ := ioutil.TempDir(os.TempDir(), min_framework.DbFile)
			db, err = iostdb.NewLDBDatabase(dirname, 0, 0)
			if err != nil{
				return "Can't open database"
			}

			nn = p2p.NewNaiveNetwork()
			lis, err := nn.Listen(uint16(port))
			if err != nil {
				return fmt.Sprint(err) + "\n"
			}

			wg.Add(1)
			go func(<-chan p2p.Request, ) {
				defer wg.Done()
				for {
					select{
					case message := <-lis:
						//fmt.Printf("\n%+v\n>", message)
						encodedBlock := message.Body
						block := transaction.DeserializeBlock(encodedBlock)
						err1 := db.Put(block.Hash, encodedBlock)
						err2 := db.Put([]byte("l"), block.Hash)
						if err1 != nil || err2 != nil {
							fmt.Printf("Write to database error! \nSync failed.\n>")
						}else{
							fmt.Printf("Sync successfully!\n>")
						}
					case <-done:
						fmt.Printf("Port %d is done\n", port)
						return
					}
				}
			}(lis)
			return fmt.Sprintf("Connected with port %d successfully!\n", port)
		},
	}
	RegistCmd(connect)

	broadcast := Cmd{
		Name:  "broadcast",
		Usage: `broadcast`,
		Exec: func(args []string) string {
			nn.Send(p2p.Request{
				Time:    1,
				From:    "test1",
				To:      "test2",
				ReqType: 1,
				Body:    []byte{1, 1, 2, 3},
			})
			return "Broadcast successfully!\n"
		},
	}
	RegistCmd(broadcast)
}
