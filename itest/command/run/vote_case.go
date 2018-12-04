package run

import (
	"github.com/iost-official/go-iost/itest"
	"github.com/urfave/cli"
)

// VoteCaseCommand is the command of vote test case
var VoteCaseCommand = cli.Command{
	Name:      "vote_case",
	ShortName: "v_case",
	Usage:     "run Vote test case",
	Flags:     VoteCaseFlags,
	Action:    VoteCaseAction,
}

// VoteCaseFlags is the flags of vote test case
var VoteCaseFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "number, n",
		Value: 1000,
		Usage: "number of transaction",
	},
	cli.IntFlag{
		Name:  "pnumber, pn",
		Value: 1,
		Usage: "number of producer",
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

// VoteCaseAction is the action of vote test case
var VoteCaseAction = func(c *cli.Context) error {
	afile := c.String("account")
	output := c.String("output")
	tnum := c.Int("number")
	punm := c.Int("pnumber")
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

	if err := it.VoteN(tnum, punm, accounts); err != nil {
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
