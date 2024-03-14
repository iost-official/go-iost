package run

import (
	"github.com/iost-official/go-iost/v3/itest"
	"github.com/urfave/cli/v2"
)

// ContractCaseCommand is the command of contract test case
var ContractCaseCommand = &cli.Command{
	Name:    "contract_case",
	Aliases: []string{"c_case"},
	Usage:   "run contract test case",
	Flags:   ContractCaseFlags,
	Action:  ContractCaseAction,
}

// ContractCaseFlags ...
var ContractCaseFlags = []cli.Flag{
	&cli.IntFlag{
		Name:  "number, n",
		Value: 1000,
		Usage: "number of transaction",
	},
	&cli.StringFlag{
		Name:  "account, a",
		Value: "accounts.json",
		Usage: "load accounts from `FILE`",
	},
	&cli.StringFlag{
		Name:  "output, o",
		Value: "accounts.json",
		Usage: "output of account information",
	},
	&cli.IntFlag{
		Name:  "memo, m",
		Value: 0,
		Usage: "The size of a random memo message that would be contained in the transaction",
	},
}

// ContractCaseAction is the action of contract test case
var ContractCaseAction = func(c *cli.Context) error {
	afile := c.String("account")
	output := c.String("output")
	tnum := c.Int("number")
	keysfile := c.String("keys")
	configfile := c.String("config")
	codefile := c.String("code")
	abifile := c.String("abi")
	memoSize := c.Int("memo")

	it, err := itest.Load(keysfile, configfile)
	if err != nil {
		return err
	}

	contract, err := itest.LoadContract(codefile, abifile)
	if err != nil {
		return err
	}

	accounts, err := itest.LoadAccounts(afile)
	if err != nil {
		return err
	}

	cid, err := it.SetContract(contract)
	if err != nil {
		return err
	}

	if _, err := it.ContractTransferN(cid, tnum, accounts, memoSize, true); err != nil {
		return err
	}

	if err := it.CheckAccounts(accounts); err != nil {
		return err
	}

	if err := itest.DumpAccounts(accounts, output); err != nil {
		return err
	}

	return nil
}
