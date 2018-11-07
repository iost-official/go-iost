package run

import "github.com/urfave/cli"

var TransferCaseCommand = cli.Command{
	Name:   "transfer_case, t_case",
	Usage:  "run transfer test case",
	Action: TransferCaseAction,
}

var TransferCaseAction = func(c *cli.Context) error {
	return nil
}
