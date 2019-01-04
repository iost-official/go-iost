package run

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/itest"
	"github.com/urfave/cli"
)

// BenchmarkCommand is the subcommand for benchmark.
var BenchmarkCommand = cli.Command{
	Name:      "benchmark",
	ShortName: "bench",
	Usage:     "Run benchmark by given tps",
	Flags:     BenchmarkFlags,
	Action:    BenchmarkAction,
}

// BenchmarkFlags is the list of flags for benchmark.
var BenchmarkFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "tps",
		Value: 100,
		Usage: "The expected ratio of transactions per second",
	},
}

// BenchmarkAction is the action of benchmark.
var BenchmarkAction = func(c *cli.Context) error {
	keyFile := c.GlobalString("keys")
	configFile := c.GlobalString("config")
	it, err := itest.Load(keyFile, configFile)
	if err != nil {
		return err
	}

	accountFile := c.GlobalString("account")
	accounts, err := itest.LoadAccounts(accountFile)
	if err != nil {
		if err := AccountCaseAction(c); err != nil {
			return err
		}
		if accounts, err = itest.LoadAccounts(accountFile); err != nil {
			return err
		}
	}

	tps := c.Int("tps")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	counter := 0
	total := 0
	startTime := time.Now()

	ticker := time.NewTicker(time.Second)
	for {
		num, err := it.TransferN(tps, accounts, false)
		if err != nil {
			ilog.Infoln(err)
		}
		select {
		case <-sig:
			return itest.DumpAccounts(accounts, accountFile)
		case <-ticker.C:
		}

		counter++
		total += num
		if counter == 10 {
			ilog.Warnf("Current tps: %v", float64(total)/time.Now().Sub(startTime).Seconds())
			counter = 0
			total = 0
			startTime = time.Now()
		}
	}
}
