package run

import (
	"github.com/iost-official/go-iost/itest"
	"github.com/urfave/cli"
)

var AccountCaseCommand = cli.Command{
	Name:      "account_case",
	ShortName: "a_case",
	Usage:     "run account test case",
	Flags:     AccountCaseFlags,
	Action:    AccountCaseAction,
}

var AccountCaseFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "account, a",
		Value: 100,
		Usage: "number of account",
	},
}

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
