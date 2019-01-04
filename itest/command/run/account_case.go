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
	Action:    AccountCaseAction,
}

// AccountCaseAction is the action of account test case
var AccountCaseAction = func(c *cli.Context) error {
	anum := c.GlobalInt("anum")
	output := c.GlobalString("account")
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

	if err := itest.DumpAccounts(accounts, output); err != nil {
		return err
	}

	return nil
}
