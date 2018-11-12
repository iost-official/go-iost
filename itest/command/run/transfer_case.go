package run

import "github.com/urfave/cli"

var TransferCaseCommand = cli.Command{
	Name:      "transfer_case",
	ShortName: "t_case",
	Usage:     "run transfer test case",
	Flags:     TransferCaseFlags,
	Action:    TransferCaseAction,
}

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

var TransferCaseAction = func(c *cli.Context) error {
	return nil
}
