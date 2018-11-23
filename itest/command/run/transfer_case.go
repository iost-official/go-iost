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
		Name:  "number, n",
		Value: 1000,
		Usage: "number of transaction",
	},
	cli.StringFlag{
		Name:  "account, a",
		Value: "accounts.json",
		Usage: "load accounts from `FILE`",
	},
	cli.StringFlag{
		Name:  "output, o",
		Value: "accounts.json",
		Usage: "output of account information",
	},
}

// TransferCaseAction is the action of transfer test case
var TransferCaseAction = func(c *cli.Context) error {
	afile := c.String("account")
	output := c.String("output")
	tnum := c.Int("number")
	keysfile := c.GlobalString("keys")
	configfile := c.GlobalString("config")

	it, err := itest.Load(keysfile, configfile)
	if err != nil {
		return err
	}

	accounts, err := itest.LoadAccounts(afile)
	if err != nil {
		return err
	}

	if err := it.TransferN(tnum, accounts); err != nil {
		return err
	}

	if err := it.CheckAccounts(accounts); err != nil {
		return err
	}

	if err := itest.DumpAccounts(accounts, output); err != nil {
		return err
	}

	return nil
}
