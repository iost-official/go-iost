package run

import (
	"fmt"

	"github.com/iost-official/go-iost/itest"
	log "github.com/sirupsen/logrus"
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
		Name:  "number, n",
		Value: 100,
		Usage: "number of account",
	},
}

var AccountCaseAction = func(c *cli.Context) error {
	num := c.Int("number")
	keysfile := c.GlobalString("keys")
	configfile := c.GlobalString("config")

	log.Infof("Load config from file...")

	keys, err := itest.LoadKeys(keysfile)
	if err != nil {
		return fmt.Errorf("load keys failed: %v", err)
	}

	itc, err := itest.LoadITestConfig(configfile)
	if err != nil {
		return fmt.Errorf("load itest config failed: %v", err)
	}

	it := itest.New(itc, keys)

	log.Infof("Load config from file successful!")

	log.Infof("Create %v account...", num)

	for i := 0; i < num; i++ {
		name := fmt.Sprintf("account%04d", i)
		_, err := it.CreateAccount(name)
		if err != nil {
			return err
		}
		// TODO Get account by rpc, and compare account result
	}

	log.Infof("Create %v account successful!", c.Int("number"))
	return nil
}
