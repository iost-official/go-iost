package create

import (
	"github.com/iost-official/go-iost/v3/crypto"
	"github.com/iost-official/go-iost/v3/itest"
	"github.com/urfave/cli/v2"
)

var keyCommand = &cli.Command{
	Name:   "key",
	Usage:  "create key pair for test",
	Flags:  keyFlags,
	Action: keyAction,
}

var keyFlags = []cli.Flag{
	&cli.IntFlag{
		Name:  "number, n",
		Value: 100,
		Usage: "number of key pair",
	},
	&cli.StringFlag{
		Name:  "algorithm, a",
		Value: "ed25519",
		Usage: "the algorithm for creating key pair, ed25519 or secp256k1",
	},
	&cli.StringFlag{
		Name:  "output, o",
		Value: "keys.json",
		Usage: "the output file name for creating key pair",
	},
}

var keyAction = func(c *cli.Context) error {
	num := c.Int("number")
	algo := crypto.NewAlgorithm(c.String("algorithm"))
	ofile := c.String("output")

	keys := make([]*itest.Key, 0)
	for i := 0; i < num; i++ {
		key := itest.NewKey(nil, algo)
		keys = append(keys, key)
	}

	if err := itest.DumpKeys(keys, ofile); err != nil {
		return err
	}

	return nil
}
