package run

import (
	"fmt"
	"math/rand"

	"github.com/iost-official/go-iost/itest"
	"github.com/urfave/cli"
)

var TransferCaseCommand = cli.Command{
	Name:      "transfer_case",
	ShortName: "t_case",
	Usage:     "run transfer test case",
	Flags:     TransferCaseFlags,
	Action:    TransferCaseAction,
}

var TransferCaseFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "account, a",
		Value: 100,
		Usage: "number of account",
	},
	cli.IntFlag{
		Name:  "transaction, t",
		Value: 10000,
		Usage: "number of transaction",
	},
}

var TransferCaseAction = func(c *cli.Context) error {
	anum := c.Int("account")
	tnum := c.Int("transaction")
	keysfile := c.GlobalString("keys")
	configfile := c.GlobalString("config")

	it, err := itest.Load(keysfile, configfile)
	if err != nil {
		return err
	}

	accounts, err := it.CreateAccountN(anum)
	if err != nil {
		return err
	}

	for i := 0; i < tnum; i++ {
		A := accounts[rand.Intn(len(accounts))]
		B := accounts[rand.Intn(len(accounts))]
		amount := fmt.Sprintf("%f", float32(rand.Int63n(10000000000))/100000000)

		err := it.Transfer(A, "iost", B.ID, amount)
		if err != nil {
			return err
		}
	}

	return nil
}
