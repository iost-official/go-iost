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

package new_cmd

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/consensus"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
	"github.com/iost-official/Go-IOS-Protocol/core/state"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/log"
	"github.com/iost-official/Go-IOS-Protocol/metrics"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"github.com/iost-official/Go-IOS-Protocol/rpc"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"os/signal"
	"syscall"

	"github.com/iost-official/Go-IOS-Protocol/consensus/pob"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/vm"
)

var cfgFile string
var logFile string
var dbFile string
var cpuprofile string
var memprofile string

type ServerExit interface {
	Stop()
}

var serverExit []ServerExit

func goroutineHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	p := pprof.Lookup("goroutine")
	p.WriteTo(w, 1)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iserver",
	Short: "IOST server",
	Long:  `IOST server`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		conf, err := common.NewConfig(viper.GetViper())
		if err != nil {
			os.Exit(1)
		}

		if err := conf.LocalConfig(); err != nil {
			os.Exit(1)
		}

		if conf.LogPath != "" {
			log.Path = conf.LogPath
		}

		glbl, err := global.New(conf)
		if err != nil {
			os.Exit(1)
		}

		// Log Server Information
		log.NewLogger("iost")
		log.Log.I("Version:  %v", "1.0")

		log.Log.I("cfgFile: %v", glbl.Config().CfgFile)
		log.Log.I("logFile: %v", glbl.Config().LogFile)
		log.Log.I("ldb.path: %v", glbl.Config().LdbPath)
		log.Log.I("dbFile: %v", glbl.Config().DbFile)

		// Start CPU Profile
		if cpuprofile != "" {
			f, err := os.Create(cpuprofile)
			if err != nil {
				log.Log.E("could not create CPU profile: ", err)
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Log.E("could not start CPU profile: ", err)
			}
		}

		log.Log.I("1.Start the P2P networks")

		logPath := viper.GetString("net.log-path")
		nodeTablePath := viper.GetString("net.node-table-path")
		nodeID := viper.GetString("net.node-id") //optional
		listenAddr := viper.GetString("net.listen-addr")
		regAddr := viper.GetString("net.register-addr")
		rpcPort := viper.GetString("net.rpc-port")
		target := viper.GetString("net.target") //optional
		port := viper.GetInt64("net.port")
		metricsPort := viper.GetString("net.metrics-port")

		log.Log.I("net.log-path:  %v", logPath)
		log.Log.I("net.node-table-path:  %v", nodeTablePath)
		log.Log.I("net.node-id:   %v", nodeID)
		log.Log.I("net.listen-addr:  %v", listenAddr)
		log.Log.I("net.register-addr:  %v", regAddr)
		log.Log.I("net.target:  %v", target)
		log.Log.I("net.port:  %v", port)
		log.Log.I("net.rpcPort:  %v", rpcPort)
		log.Log.I("net.metricsPort:  %v", metricsPort)

		if logPath == "" || nodeTablePath == "" || listenAddr == "" || regAddr == "" || port <= 0 || rpcPort == "" {
			log.Log.E("Network config initialization failed, stop the program!")
			os.Exit(1)
		}

		log.Log.I("network instance")
		p2pService, err := p2p.NewDefault()
		if err != nil {
			log.Log.E("Network initialization failed, stop the program! err:%v", err)
			os.Exit(1)
		}
		serverExit = append(serverExit, p2pService)

		accSecKey := viper.GetString("account.sec-key")
		//fmt.Printf("account.sec-key:  %v\n", accSecKey)

		acc, err := account.NewAccount(common.Base58Decode(accSecKey))
		if err != nil {
			log.Log.E("NewAccount failed, stop the program! err:%v", err)
			os.Exit(1)
		}

		account.MainAccount = acc

		log.Log.I("account ID = %v", acc.ID)

		// init servi
		sp, err := tx.NewServiPool(len(account.GenesisAccount), 100)
		if err != nil {
			log.Log.E("NewServiPool failed, stop the program! err:%v", err)
			os.Exit(1)
		}
		tx.Data = tx.NewHolder(acc, state.StdPool, sp)
		tx.Data.Spool.Restore()
		bu, _ := tx.Data.Spool.BestUser()

		if len(bu) != len(account.GenesisAccount) {
			tx.Data.Spool.ClearBtu()
			for k, v := range account.GenesisAccount {
				ser, err := tx.Data.Spool.User(vm.IOSTAccount(k))
				if err == nil {
					ser.SetBalance(v)
				}

			}
			tx.Data.Spool.Flush()
		}
		witnessList := make([]string, 0)

		bu, err = tx.Data.Spool.BestUser()
		if err != nil {
			for k, _ := range account.GenesisAccount {
				witnessList = append(witnessList, k)
			}
		} else {
			for _, v := range bu {
				witnessList = append(witnessList, string(v.Owner()))
			}
		}

		for i, witness := range witnessList {
			log.Log.I("witnessList[%v] = %v", i, witness)
		}

		var blkCache blockcache.BlockCache
		blkCache, err = blockcache.NewBlockCache(glbl)
		if err != nil {
			log.Log.E("blockcache initialization failed, stop the program! err:%v", err)
			os.Exit(1)
		}

		sync, err = consensus_common.NewSynchronizer(glbl, blkCache, p2pService)
		if err != nil {
			log.Log.E("synchronizer initialization failed, stop the program! err:%v", err)
			os.Exit(1)
		}
		serverExit = append(serverExit, sync)

		var txpool new_txpool.TxPool
		txPool, err = new_txpool.NewTxPoolImpl(glbl, blkCache, p2pService)
		if err != nil {
			log.Log.E("NewTxPoolServer failed, stop the program! err:%v", err)
			os.Exit(1)
		}
		txPool.Start()
		serverExit = append(serverExit, txPool)

		consensus, err := consensus.ConsensusFactory(
			consensus.CONSENSUS_POB,
			acc, blockChain, state.StdPool, witnessList)
		if err != nil {
			log.Log.E("consensus initialization failed, stop the program! err:%v", err)
			os.Exit(1)
		}
		consensus.Run()
		serverExit = append(serverExit, consensus)

		err = rpc.Server(rpcPort)
		if err != nil {
			log.Log.E("RPC initialization failed, stop the program! err:%v", err)
			os.Exit(1)
		}

		recorder := pob.NewRecorder()
		recorder.Listen()

		// Start Metrics Server
		if metricsPort != "" {
			metrics.NewServer(metricsPort)
		}

		////////////probe//////////////////
		log.Report(&log.MsgNode{
			SubType: "online",
		})
		///////////////////////////////////

		exitLoop()

	},
}

func exitLoop() {
	exit := make(chan bool)
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		i := <-c
		log.Log.I("IOST server received interrupt[%v], shutting down...", i)

		for _, s := range serverExit {
			if s != nil {
				s.Stop()
			}
		}
		////////////probe//////////////////
		log.Report(&log.MsgNode{
			SubType: "offline",
		})
		///////////////////////////////////
		exit <- true
		// os.Exit(0)
	}()

	<-exit
	// Stop Cpu Profile
	if cpuprofile != "" {
		pprof.StopCPUProfile()
	}
	// Start Memory Profile
	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Log.E("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Log.E("could not write memory profile: ", err)
		}
		f.Close()
	}

	signal.Stop(c)
	close(exit)
	os.Exit(0)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Log.E("Execute err: %v", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "iserver.yml", "config file (default is $HOME/.iserver.yaml)")
	rootCmd.PersistentFlags().StringVar(&logFile, "log", "", "log file (default is ./iserver.log)")
	rootCmd.PersistentFlags().StringVar(&dbFile, "db", "", "database file (default is ./data.db)")
	rootCmd.PersistentFlags().StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to `file`")
	rootCmd.PersistentFlags().StringVar(&memprofile, "memprofile", "", "write memory profile to `file`")

	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("log", rootCmd.PersistentFlags().Lookup("log"))
	viper.BindPFlag("db", rootCmd.PersistentFlags().Lookup("db"))
	viper.BindPFlag("cpuprofile", rootCmd.PersistentFlags().Lookup("cpuprofile"))
	viper.BindPFlag("memprofile", rootCmd.PersistentFlags().Lookup("memprofile"))

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
	} else {
		panic(err)
	}

}
