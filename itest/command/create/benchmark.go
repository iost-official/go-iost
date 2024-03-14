package create

import "github.com/urfave/cli/v2"

var benchmarkCommand = &cli.Command{
	Name:    "benchmark",
	Aliases: []string{"bench"},
	Usage:   "create benchmark transaction for test",
	Action:  benchmarkAction,
}

var benchmarkAction = func(c *cli.Context) error {
	return nil
}
