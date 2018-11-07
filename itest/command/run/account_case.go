package run

import "github.com/urfave/cli"

var AccountCaseCommand = cli.Command{
	Name:   "account_case, a_case",
	Usage:  "run account test case",
	Action: AccountCaseAction,
}

var AccountCaseAction = func(c *cli.Context) error {
	return nil
}
