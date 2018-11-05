package create

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/itest"
	"github.com/urfave/cli"
)

var keyCommand = cli.Command{
	Name:   "key",
	Usage:  "create key pair for test",
	Flags:  keyFlags,
	Action: keyAction,
}

var keyFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "number, n",
		Value: 100,
		Usage: "number of key pair",
	},
	cli.StringFlag{
		Name:  "algorithm, a",
		Value: "ed25519",
		Usage: "the algorithm for creating key pair, ed25519 or secp256k1",
	},
	cli.StringFlag{
		Name:  "output, o",
		Value: "keys.txt",
		Usage: "the output file name for creating key pair",
	},
}

var keyAction = func(c *cli.Context) error {
	num := c.Int("number")
	algo := crypto.NewAlgorithm(c.String("algorithm"))
	ofile := c.String("output")

	f, err := os.Create(ofile)
	if err != nil {
		return err
	}
	defer f.Close()

	for i := 0; i < num; i++ {
		key := itest.NewKey(algo)
		b, err := json.Marshal(key)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(f, string(b)); err != nil {
			return err
		}
	}
	return nil
}
