package run

import (
	"fmt"

	"github.com/iost-official/go-iost/ilog"
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
	num := c.Int("number")
	keysfile := c.GlobalString("keys")
	configfile := c.GlobalString("config")

	ilog.Infof("Load config from file...")

	keys, err := itest.LoadKeys(keysfile)
	if err != nil {
		return fmt.Errorf("load keys failed: %v", err)
	}

	itc, err := itest.LoadITestConfig(configfile)
	if err != nil {
		return fmt.Errorf("load itest config failed: %v", err)
	}

	it := itest.New(itc, keys)

	ilog.Infof("Load config from file successful!")

	if _, err := it.CreateAccountN(num); err != nil {
		return err
	}

	return nil
}
