package run

import (
	"github.com/urfave/cli"
)

// Command is the command of run
var Command = cli.Command{
	Name:  "run",
	Usage: "run test by benchmark data",
	Flags: Flags,
	Subcommands: []cli.Command{
		AccountCaseCommand,
		TransferCaseCommand,
		ContractCaseCommand,
		BenchmarkCommand,
	},
}

// Flags is the flags of run command
var Flags = []cli.Flag{
	cli.StringFlag{
		Name:  "keys, k",
		Value: "",
		Usage: "Load keys configuration from `FILE`",
	},
	cli.StringFlag{
		Name:  "config, c",
		Value: "",
		Usage: "Load itest configuration from `FILE`",
	},
}
