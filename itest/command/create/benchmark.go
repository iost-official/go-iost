package create

import "github.com/urfave/cli"

var benchmarkCommand = cli.Command{
	Name:      "benchmark",
	ShortName: "bench",
	Usage:     "create benchmark transaction for test",
	Action:    benchmarkAction,
}

var benchmarkAction = func(c *cli.Context) error {
	return nil
}
