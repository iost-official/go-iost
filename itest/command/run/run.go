package run

import (
	"github.com/urfave/cli"
)

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

var Flags = []cli.Flag{
	cli.StringFlag{
		Name:  "bank, b",
		Usage: "Load bank configuration from `FILE`",
	},
	cli.StringFlag{
		Name:  "keys, k",
		Usage: "Load keys configuration from `FILE`",
	},
	cli.StringFlag{
		Name:  "clients, c",
		Usage: "Load clients configuration from `FILE`",
	},
}
