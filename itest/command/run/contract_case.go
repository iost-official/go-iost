package run

import "github.com/urfave/cli"

// ContractCaseCommand is the command of contract test case
var ContractCaseCommand = cli.Command{
	Name:      "contract_case",
	ShortName: "c_case",
	Usage:     "run contract test case",
	Action:    ContractCaseAction,
}

// ContractCaseAction is the action of contract test case
var ContractCaseAction = func(c *cli.Context) error {
	return nil
}
