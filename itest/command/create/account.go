package create

import "github.com/urfave/cli/v2"

var accountCommand = &cli.Command{
	Name:   "account",
	Usage:  "create account transaction for test",
	Action: accountAction,
}

var accountAction = func(c *cli.Context) error {
	return nil
}
