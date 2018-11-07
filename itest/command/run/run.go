package run

import (
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:  "run",
	Usage: "run test by benchmark data",
	Subcommands: []cli.Command{
		AccountCaseCommand,
		TransferCaseCommand,
		ContractCaseCommand,
		BenchmarkCommand,
	},
}
