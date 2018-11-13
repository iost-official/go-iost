package run

import (
	"github.com/iost-official/go-iost/itest"
	"github.com/urfave/cli"
)

// TransferCaseCommand is the command of transfer test case
var TransferCaseCommand = cli.Command{
	Name:      "transfer_case",
	ShortName: "t_case",
	Usage:     "run transfer test case",
	Flags:     TransferCaseFlags,
	Action:    TransferCaseAction,
}

// TransferCaseFlags is the flags of transfer test case
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

// TransferCaseAction is the action of transfer test case
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

	if err := it.TransferN(tnum, accounts); err != nil {
		return err
	}

	return nil
}
