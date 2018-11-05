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
