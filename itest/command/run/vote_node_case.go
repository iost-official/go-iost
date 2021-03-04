package run

import (
	"github.com/iost-official/go-iost/v3/itest"
	"github.com/urfave/cli"
)

// VoteNodeCaseCommand is vote or unvote for the test of the authentication node
var VoteNodeCaseCommand = cli.Command{
	Name:      "vote_node_case",
	ShortName: "n_case",
	Usage:     "run a test to vote or unvote for node",
	Flags:     VoteNodeCaseFlags,
	Action:    VoteNodeCaseAction,
}

// VoteNodeCaseFlags is the flags of vote test case
var VoteNodeCaseFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "number, n",
		Value: 0,
		Usage: "number of vote",
	},
	cli.BoolFlag{
		Name:  "unvote, u",
		Usage: "Cancel vote based on configuration file",
	},
}

// VoteNodeCaseAction is the action of vote test case
var VoteNodeCaseAction = func(c *cli.Context) error {
	tnum := c.Int("number")
	unvote := c.Bool("unvote")
	keysfile := c.GlobalString("keys")
	configfile := c.GlobalString("config")
	afile := c.GlobalString("account")

	it, err := itest.Load(keysfile, configfile)
	if err != nil {
		return err
	}

	accounts, err := itest.LoadAccounts(afile)
	if err != nil {
		return err
	}

	if unvote {
		if err := it.CancelVoteNode(tnum, accounts); err != nil {
			return err
		}
	} else {
		if err := it.VoteNode(tnum, accounts); err != nil {
			return err
		}
	}

	return nil
}
