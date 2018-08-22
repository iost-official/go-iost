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

package main

import (
	"os"
	"os/signal"
	"syscall"

	"fmt"

	"flag"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/consensus"
	"github.com/iost-official/Go-IOS-Protocol/consensus/synchronizer"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

//	"github.com/iost-official/Go-IOS-Protocol/iserver/cmd"
type ServerExit interface {
	Stop()
}

var (
	cfgFile    = "iserver.yml"
	logFile    string
	dbFile     string
	cpuprofile string
	memprofile string
)

var serverExit []ServerExit

func main() {
	//	cmd.Execute()
	flag.Parse()
	conf := common.NewConfig(*configfile)

	glb, err := global.New(conf)
	if err != nil {
		os.Exit(1)
	}
	// Log Server Information
	ilog.Info("Version:  %v", "1.0")
	ilog.Info("cfgFile: %v", glb.Config().CfgFile)
	ilog.Info("logFile: %v", glb.Config().LogFile)
	ilog.Info("ldb.path: %v", glb.Config().LdbPath)
	ilog.Info("dbFile: %v", glb.Config().DbFile)
	// Start CPU Profile
	/*
		if cpuprofile != "" {
			f, err := os.Create(cpuprofile)
			if err != nil {
				//ilog.E("could not create CPU profile: ", err)
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				//ilog.E("could not start CPU profile: ", err)
			}
		}*/

	// ilog.I("1.Start the P2P networks")

	// rpcPort := viper.GetString("net.rpc-port")
	// metricsPort := viper.GetString("net.metrics-port")

	//ilog.I("network instance")
	p2pService, err := p2p.NewDefault()
	if err != nil {
		ilog.Fatal("Network initialization failed, stop the program! err:%v", err)
	}

	serverExit = append(serverExit, p2pService)

	accSecKey := glb.Config().AccSecKey
	//fmt.Printf("account.sec-key:  %v\n", accSecKey)
	acc, err := account.NewAccount(common.Base58Decode(accSecKey))
	if err != nil {
		ilog.Fatal("NewAccount failed, stop the program! err:%v", err)
	}
	account.MainAccount = acc
	//ilog.I("account ID = %v", acc.ID)
	/*
			// init servi
			sp, err := tx.NewServiPool(len(account.GenesisAccount), 100)
			if err != nil {
				ilog.E("NewServiPool failed, stop the program! err:%v", err)
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
			ilog.I("witnessList[%v] = %v", i, witness)
		}
	*/

	var blkCache blockcache.BlockCache
	blkCache, err = blockcache.NewBlockCache(glb)
	if err != nil {
		ilog.Fatal("blockcache initialization failed, stop the program! err:%v", err)
	}

	var sync synchronizer.Synchronizer
	sync, err = synchronizer.NewSynchronizer(glb, blkCache, p2pService)
	if err != nil {
		ilog.Fatal("synchronizer initialization failed, stop the program! err:%v", err)
	}
	serverExit = append(serverExit, sync)

	var txp txpool.TxPool
	txp, err = txpool.NewTxPoolImpl(glb, blkCache, p2pService)
	if err != nil {
		ilog.Fatal("txpool initialization failed, stop the program! err:%v", err)
	}
	txp.Start()
	serverExit = append(serverExit, txp)

	var witnessList []string
	for k := range account.GenesisAccount {
		witnessList = append(witnessList, k)
	}

	consensus, err := consensus.Factory(
		"pob",
		acc, glb, blkCache, txp, p2pService, sync, witnessList) //witnessList)
	if err != nil {
		ilog.Fatal("consensus initialization failed, stop the program! err:%v", err)
	}
	consensus.Run()
	serverExit = append(serverExit, consensus)

	/*
		err = rpc.Server(rpcPort)
		if err != nil {
			//ilog.E("RPC initialization failed, stop the program! err:%v", err)
			os.Exit(1)
		}
			recorder := pob.NewRecorder()
			recorder.Listen()

			if metricsPort != "" {
				metrics.NewServer(metricsPort)
			}
	*/

	exitLoop()

}

func exitLoop() {
	exit := make(chan bool)
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		i := <-c
		ilog.Info("IOST server received interrupt[%v], shutting down...", i)

		for _, s := range serverExit {
			if s != nil {
				s.Stop()
			}
		}
		exit <- true
		// os.Exit(0)
	}()

	<-exit
	// Stop Cpu Profile
	/*
		if cpuprofile != "" {
			pprof.StopCPUProfile()
		}
	*/
	// Start Memory Profile
	/*
		if memprofile != "" {
			f, err := os.Create(memprofile)
			if err != nil {
				//ilog.E("could not create memory profile: ", err)
			}
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				//ilog.E("could not write memory profile: ", err)
			}
			f.Close()
		}

	*/
	signal.Stop(c)
	close(exit)
	os.Exit(0)
}

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
	//fmt.Println(cfgFile)
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		panic(err)
	}
}
