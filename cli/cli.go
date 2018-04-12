package cli

import (
	"os"
	"fmt"
	"log"
	"github.com/iost-official/prototype/tx/tx"
	"github.com/urfave/cli"
)

func Run(){
	app := cli.NewApp()
	app.Name = "blockchain-tx"
	app.Usage = "Test the transaction part of blockchain."
	app.Commands = []cli.Command{
		{
			Name: "getbalance",
			Aliases: []string{"a"},
			Usage: "Get balance of ADDRESS",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "address",
					Value:		  "",
					Usage:       "address of account",
				},
			},
			Action:  func(c *cli.Context) error {
				address := c.String("address")
				if address == ""{
					fmt.Println("Address can't be empty!")
					return nil
				}
				getBalance(address)
				return nil
			},
		},
		{
			Name: "createblockchain",
			Aliases: []string{"c"},
			Usage: "Create a blockchain and send genesis block reward to ADDRESS",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "address",
					Value:		 "",
					Usage:       "address of account",
				},
			},
			Action:  func(c *cli.Context) error {
				address := c.String("address")
				if address == ""{
					fmt.Println("Address can't be empty!")
					return nil
				}
				createBlockchain(address)
				return nil
			},
		},
		{
			Name: "printchain",
			Aliases: []string{"p"},
			Usage: "Print all the blocks of the blockchain",
			Action:  func(c *cli.Context) error {
				printChain()
				return nil
			},
		},
		{
			Name: "send",
			Aliases: []string{"s"},
			Usage: "Send AMOUNT of coins from FROM address to TO",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "from",
					Value: 		 "",
					Usage:       "address of from",
				},
				cli.StringFlag{
					Name:		"to",
					Value: 		"",
					Usage: 		"address of to",
				},
				cli.IntFlag{
					Name: 		"amount",
					Value: 		-1,
					Usage:		"amount of btc to be sent",
				},
			},
			Action:  func(c *cli.Context) error {
				from := c.String("from")
				to := c.String("to")
				amount := c.Int("amount")
				if from == "" || to == ""{
					fmt.Println("Address can't be empty!")
					return nil
				}
				if amount <= 0{
					fmt.Println("Amount must be greater than 0.")
					return nil
				}
				send(from, to, amount)
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}


func createBlockchain(address string) {
	bc, to_print := transaction.CreateBlockchain(address)

	if bc == nil {
		fmt.Println(to_print)
		return
	}

	fmt.Println("Done!")
}

func getBalance(address string) {
	bc, to_print := transaction.NewBlockchain(address)

	if bc == nil {
		fmt.Println(to_print)
	}

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}


func printChain() {
	bc, to_print := transaction.NewBlockchain("")

	if bc == nil {
		fmt.Println(to_print)
		return
	}

	if bc == nil {
		return
	}

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		for _, tx := range(block.Transactions) {
			fmt.Printf(tx.String())
		}
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func send(from, to string, amount int) {
	bc, to_print := transaction.NewBlockchain(from)

	if bc == nil {
		fmt.Println(to_print)
		return
	}


	tx, to_print := transaction.NewUTXOTransaction(from, to, amount, bc)

	if tx == nil {
		fmt.Println(to_print)
		return
	}

	bc.MineBlock([]*transaction.Transaction{tx})
	fmt.Println("Success!")
}