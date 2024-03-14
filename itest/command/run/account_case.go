package run

import (
	"github.com/iost-official/go-iost/v3/itest"
	"github.com/urfave/cli/v2"
)

// AccountCaseCommand is the command of account test case
var AccountCaseCommand = &cli.Command{
	Name:    "account_case",
	Aliases: []string{"a_case"},
	Usage:   "run account test case",
	Action: func(c *cli.Context) error {
		return AccountCaseAction(c)
	},
}

// AccountCaseAction is the action of account test case
var AccountCaseAction = func(c *cli.Context) error {
	anum := c.Int("anum")
	output := c.String("account")
	keysfile := c.String("keys")
	configfile := c.String("config")

	it, err := itest.Load(keysfile, configfile)
	if err != nil {
		return err
	}

	accounts, err := it.CreateAccountN(anum, false, true)
	if err != nil {
		return err
	}

	if err := itest.DumpAccounts(accounts, output); err != nil {
		return err
	}

	return nil
}
