package run

import "github.com/urfave/cli"

var Command = cli.Command{
	Name:   "run",
	Usage:  "run test by transaction data",
	Action: Action,
}

var Action = func(c *cli.Context) error {
	return nil
}
