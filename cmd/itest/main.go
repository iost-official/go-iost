package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/iost-official/go-iost/itest/command/create"
	"github.com/iost-official/go-iost/itest/command/run"
	"github.com/urfave/cli"
)

func main() {
	log.SetLevel(log.DebugLevel)

	app := cli.NewApp()
	app.Name = "itest"
	app.Usage = "The cli tool for testing the IOST testnet"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		create.Command,
		run.Command,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("Run itest failed: %v", err)
	}
}
