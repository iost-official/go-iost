// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iwallet

/*
import (
	"fmt"
	"io/ioutil"
	"os"

	"strings"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)


// signCmd represents the sign command
var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign to .sc file",
	Long:  `Sign to .sc file`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println(`Error: source file not given`)
			return
		}
		path := args[0]
		fi, err := os.Open(path)
		if err != nil {
			fmt.Println("Error: input file not found")
		}
		defer fi.Close()
		fd, err := ioutil.ReadAll(fi)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		var mtx tx.Tx
		err = mtx.Decode(fd)
		if err != nil {
			fmt.Println("file broken: ", err.Error())
		}

		fsk, err := os.Open(kpPath)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer fsk.Close()
		seckey, err := ioutil.ReadAll(fsk)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		acc, err := account.NewKeyPair(loadBytes(string(seckey)), getSignAlgo(signAlgo))
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		sig, err := tx.SignTxContent(&mtx, "todo", acc) // TODO 修改iwallet
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if len(args) < 2 {
			dest = args[0][:strings.LastIndex(args[0], ".")]
			dest = dest + ".sig"
		} else {
			dest = args[1]
		}

		sigRaw, err := sig.Encode()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		err = saveTo(dest, sigRaw)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(signCmd)

	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	signCmd.Flags().StringVarP(&kpPath, "key-path", "k", home+"/.iwallet/id_ed25519", "Set path of sec-key")
	signCmd.Flags().StringVarP(&signAlgo, "signAlgo", "a", "ed25519", "Sign algorithm")

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// signCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// signCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
*/
