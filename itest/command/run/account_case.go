package run

import (
	"fmt"

	"github.com/urfave/cli"
)

var AccountCaseCommand = cli.Command{
	Name:      "account_case",
	ShortName: "a_case",
	Usage:     "run account test case",
	Flags:     AccountCaseFlags,
	Action:    AccountCaseAction,
}

var AccountCaseFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "number, n",
		Value: 100,
		Usage: "number of account",
	},
}

var AccountCaseAction = func(c *cli.Context) error {
	fmt.Printf("Run account test with %v account...", c.Int("number"))

	fmt.Printf("Run account test with %v account successful!", c.Int("number"))
	return nil
}
