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
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/iost-official/prototype/network"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/account"
	"os/signal"
	"syscall"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/consensus"
)

var cfgFile string

type ServerExit interface {
	Stop()
}

var serverExit []ServerExit

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iserver",
	Short: "Blockchain system",
	Long:  `Blockchain system`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("Version:  %v\n", "1.0")

		//初始化网络
		fmt.Println("1.Start the P2P networks")

		logPath := viper.GetString("net.log-path")
		nodeTablePath := viper.GetString("net.node-table-path")
		nodeID := viper.GetString("net.node-id") //optional
		listenAddr := viper.GetString("net.listen-addr")
		target := viper.GetString("net.target") //optional
		port := viper.GetInt64("net.port")

		fmt.Printf("net.log-path:  %v\n", logPath)
		fmt.Printf("net.node-table-path:  %v\n", nodeTablePath)
		fmt.Printf("net.node-id:   %v\n", nodeID)
		fmt.Printf("net.listen-addr:  %v\n", listenAddr)
		fmt.Printf("net.target:  %v\n", target)
		fmt.Printf("net.port:  %v\n", port)

		if logPath == "" || nodeTablePath == "" || listenAddr == "" || port <= 0 {
			fmt.Println("Network config initialization failed, stop the program!")
			os.Exit(1)
		}

		net, err := network.GetInstance(
			&network.NetConifg{
				LogPath:       logPath,
				NodeTablePath: nodeTablePath,
				NodeID:        nodeID,
				ListenAddr:    listenAddr},
			target,
			uint16(port))
		if err != nil {

			fmt.Printf("Network initialization failed, stop the program! err:%v", err)
			os.Exit(1)
		}
		serverExit = append(serverExit, net)

		//启动共识
		fmt.Println("2.Start Consensus Services")
		accSecKey := viper.GetString("account.sec-key")
		//fmt.Printf("account.sec-key:  %v\n", accSecKey)

		acc, err := account.NewAccount(common.Base58Decode(accSecKey))
		if err != nil {
			fmt.Printf("NewAccount failed, stop the program! err:%v\n", err)
			os.Exit(1)
		}

		//fmt.Printf("account PubKey = %v\n", common.Base58Encode(acc.Pubkey))
		//fmt.Printf("account SecKey = %v\n", common.Base58Encode(acc.Seckey))
		fmt.Printf("account ID = %v\n", acc.ID)

		if state.StdPool == nil {
			fmt.Printf("StdPool initialization failed, stop the program!")
			os.Exit(1)
		}

		blockChain, err := block.NewBlockChain()
		if err != nil {
			fmt.Printf("NewBlockChain failed, stop the program! err:%v", err)
			os.Exit(1)
		}

		witnessList := viper.GetStringSlice("consensus.witness-list")

		for i, witness := range witnessList {
			fmt.Printf("witnessList[%v] = %v\n", i, witness)
		}

		consensus, err := consensus.ConsensusFactory(
			consensus.CONSENSUS_DPOS,
			acc, blockChain, state.StdPool, witnessList)
		if err != nil {
			fmt.Printf("consensus initialization failed, stop the program! err:%v", err)
			os.Exit(1)
		}

		serverExit = append(serverExit, consensus)
		//启动RPC
		//rpc.Server()

		//等待推出信号
		exitLoop()
	},
}

func exitLoop() {
	exit := make(chan bool)
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer signal.Stop(c)
	defer close(exit)

	go func() {

		<- c
		fmt.Printf("iserver received interrupt, shutting down...")

		for _, s := range serverExit {
			s.Stop()
		}

		os.Exit(0)
	}()

	<-exit
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.iserver.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".iserver" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".iserver")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

}
