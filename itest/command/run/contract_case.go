package run

import "github.com/urfave/cli"

var ContractCaseCommand = cli.Command{
	Name:      "contract_case",
	ShortName: "c_case",
	Usage:     "run contract test case",
	Action:    ContractCaseAction,
}

var ContractCaseAction = func(c *cli.Context) error {
	return nil
}
