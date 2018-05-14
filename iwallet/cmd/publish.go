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

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"context"

	"errors"

	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/tx"
	pb "github.com/iost-official/prototype/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "sign to a .sc file with .sig files, and publish it",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println(`invalid input, check
	iwallet publish -h`)
		}

		scf, err := os.Open(args[0])
		if err != nil {
			fmt.Printf("Error in File %v: %v\n", args[0], err.Error())
			return

		}
		defer scf.Close()
		sc, err := ioutil.ReadAll(scf)
		if err != nil {
			fmt.Println("Read error", err)
			return

		}
		var mtx tx.Tx
		err = mtx.Decode(sc)
		if err != nil {
			fmt.Println(err.Error())
		}

		for i, v := range args {
			if i == 0 {
				continue
			}
			sigf, err := os.Open(v)
			if err != nil {
				fmt.Printf("Error in File %v: %v\n", args[0], err.Error())
				sigf.Close()
				return
			}
			sig, err := ioutil.ReadAll(sigf)
			sigf.Close()
			if err != nil {
				fmt.Println("Error: Illegal sig file", err)
				return
			}
			var sign common.Signature
			err = sign.Decode(sig)
			if err != nil {
				fmt.Println("Error: Illegal sig file", err)
				return
			}
			if !mtx.VerifySigner(sign) {
				fmt.Printf("Error: Sign %v wrong\n", v)
				return
			}
			mtx.Signs = append(mtx.Signs, sign)
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

		acc, err := account.NewAccount(LoadBytes(string(seckey)))
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		stx, err := tx.SignTx(mtx, acc)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		dest = ChangeSuffix(args[0], ".tx")

		SaveTo(dest, stx.Encode())

		if !isLocal {
			err = sendTx(stx)
			if err != nil {
				fmt.Println(err.Error())
			}
		}

		fmt.Println("ok")
	},
}

var isLocal bool
var server string

func init() {
	rootCmd.AddCommand(publishCmd)

	publishCmd.Flags().StringVarP(&kpPath, "key-path", "k", "~/.ssh/id_secp", "Set path of sec-key")
	publishCmd.Flags().BoolVar(&isLocal, "local", false, "Set to not send tx to server")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// publishCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// publishCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func sendTx(stx tx.Tx) error {
	conn, err := grpc.Dial(server, grpc.WithDefaultCallOptions())
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewCliClient(conn)
	resp, err := client.PublishTx(context.Background(), &pb.Transaction{Tx: stx.Encode()})
	if err != nil {
		return err
	}
	switch resp.Code {
	case 0:
		return nil
	case -1:
		return errors.New("tx rejected")
	default:
		return errors.New("unknown return")
	}
}
