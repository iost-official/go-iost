package run

import (
	"github.com/iost-official/go-iost/itest"
	"github.com/urfave/cli"
)

// AccountCaseCommand is the command of account test case
var AccountCaseCommand = cli.Command{
	Name:      "account_case",
	ShortName: "a_case",
	Usage:     "run account test case",
	Flags:     AccountCaseFlags,
	Action:    AccountCaseAction,
}

// AccountCaseFlags is the flags of account test case
var AccountCaseFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "account, a",
		Value: 10,
		Usage: "number of account",
	},
}

// AccountCaseAction is the action of account test case
var AccountCaseAction = func(c *cli.Context) error {
	anum := c.Int("account")
	keysfile := c.GlobalString("keys")
	configfile := c.GlobalString("config")

	it, err := itest.Load(keysfile, configfile)
	if err != nil {
		return err
	}

	if _, err := it.CreateAccountN(anum); err != nil {
		return err
	}

	return nil
}
