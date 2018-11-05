package create

import "github.com/urfave/cli"

var Command = cli.Command{
	Name:  "create",
	Usage: "create data for test",
	Subcommands: []cli.Command{
		keyCommand,
		accountCommand,
		benchmarkCommand,
	},
}

var keyCommand = cli.Command{
	Name:   "key",
	Usage:  "create key pair for test",
	Action: keyAction,
}

var accountCommand = cli.Command{
	Name:   "account",
	Usage:  "create account transaction for test",
	Action: accountAction,
}

var benchmarkCommand = cli.Command{
	Name:      "benchmark",
	ShortName: "bench",
	Usage:     "create benchmark transaction for test",
	Action:    benchmarkAction,
}

var keyAction = func(c *cli.Context) error {
	return nil
}

var accountAction = func(c *cli.Context) error {
	return nil
}

var benchmarkAction = func(c *cli.Context) error {
	return nil
}
