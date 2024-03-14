package main

import (
	"os"

	"github.com/iost-official/go-iost/v3/ilog"
	"github.com/iost-official/go-iost/v3/itest/command/create"
	"github.com/iost-official/go-iost/v3/itest/command/run"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "itest"
	app.Usage = "The cli tool for testing the IOST blockchain"
	app.Version = "0.0.1"
	app.Commands = []*cli.Command{
		create.Command,
		run.Command,
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "level, l",
			Value: "info",
			Usage: "The level of printing log",
		},
	}
	app.Before = func(c *cli.Context) error {
		ilog.SetLevel(ilog.NewLevel(c.String("level")))
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		ilog.Fatalf("Run itest failed: %v", err)
	}
}
