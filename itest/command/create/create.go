package create

import "github.com/urfave/cli"

// Command is the command of create
var Command = cli.Command{
	Name:  "create",
	Usage: "create data for test",
	Subcommands: []cli.Command{
		keyCommand,
		accountCommand,
		benchmarkCommand,
	},
}
