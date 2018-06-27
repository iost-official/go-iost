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
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"

	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/consensus"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/blockcache"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/metrics"
	"github.com/iost-official/prototype/network"
	"github.com/iost-official/prototype/rpc"
	"github.com/iost-official/prototype/verifier"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"os/signal"
	"syscall"

	"github.com/iost-official/prototype/consensus/pob"
	"github.com/iost-official/prototype/core/txpool"
	"github.com/iost-official/prototype/vm"
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

func chainServer(chain block.Chain) {
	http.HandleFunc("/debug/goroutine", goroutineHandler)
	http.HandleFunc("/blockchain/length", func(w http.ResponseWriter, r *http.Request) {
		len := chain.Length()
		resp := map[string]interface{}{
			"len": len,
		}
		bytes, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("json encode error. err=", err)
		}
		w.Write(bytes)
	})
	http.HandleFunc("/blockchain", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		num := r.Form["number"]
		has := r.Form["hash"]
		if len(num) == 0 && len(has) == 0 {
			resp := map[string]interface{}{
				"message": "wrong parameter. missing hash or number!",
			}
			bytes, err := json.Marshal(resp)
			if err != nil {
				fmt.Println("json encode error. err=", err)
			}
			w.Write(bytes)
			return
		}
		if len(num) > 0 {
			n, _ := strconv.ParseUint(num[0], 10, 64)
			blk := chain.GetBlockByNumber(n)
			blkStr := blk.String()
			md5Ctx := md5.New()
			md5Ctx.Write([]byte(blkStr))
			cipherStr := md5Ctx.Sum(nil)
			resp := map[string]interface{}{
				"number": n,
				"block":  blkStr,
				"md5":    string(cipherStr),
			}
			bytes, err := json.Marshal(resp)
			if err != nil {
				fmt.Println("json encode error. err=", err)
			}
			w.Write(bytes)
			return
		}
		if len(has) > 0 {
			blk := chain.GetBlockByHash([]byte(has[0]))
			blkStr := blk.String()
			md5Ctx := md5.New()
			md5Ctx.Write([]byte(blkStr))
			cipherStr := md5Ctx.Sum(nil)
			resp := map[string]interface{}{
				"hash":  has[0],
				"block": blkStr,
				"md5":   string(cipherStr),
			}
			bytes, err := json.Marshal(resp)
			if err != nil {
				fmt.Println("json encode error. err=", err)
			}
			w.Write(bytes)
			return
		}
	})
	http.HandleFunc("/transaction", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if len(r.Form["hash"]) == 0 {
			resp := map[string]interface{}{
				"message": "wrong parameter. missing hash!",
			}
			bytes, err := json.Marshal(resp)
			if err != nil {
				fmt.Println("get blockchain length failed. err=", err)
			}
			w.Write(bytes)
			return
		}
		has := r.Form["hash"][0]
		tx, err := chain.GetTx([]byte(has))
		resp := map[string]interface{}{
			"err": err,
			"tx":  tx,
		}
		bytes, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("json encode error. err=", err)
		}
		w.Write(bytes)
	})
	/*  http.HandleFunc("/blockchain/length", func(w http.ResponseWriter, r *http.Request) { */
	// len := chain.Length()
	// resp := map[string]interface{}{
	// "len": len,
	// }
	// bytes, err := json.Marshal(resp)
	// if err != nil {
	// fmt.Println("get blockchain length failed. err=", err)
	// }
	// w.Write(bytes)
	/* }) */
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iserver",
	Short: "IOST server",
	Long:  `IOST server`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		lp := viper.GetString("log.path")
		if lp != "" {
			log.Path = lp
		}
		// Log Server Information
		log.NewLogger("iost")
		log.Log.I("Version:  %v", "1.0")

		log.Log.I("cfgFile: %v", viper.GetString("config"))
		log.Log.I("logFile: %v", viper.GetString("log"))
		log.Log.I("dbFile: %v", viper.GetString("db"))

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

		//初始化数据库
		ldbPath := viper.GetString("ldb.path")
		redisAddr := viper.GetString("redis.addr") //optional
		redisPort := viper.GetInt64("redis.port")

		log.Log.I("ldb.path: %v", ldbPath)
		log.Log.I("redis.addr: %v", redisAddr)
		log.Log.I("redis.port: %v", redisPort)

		tx.LdbPath = ldbPath
		block.LdbPath = ldbPath
		db.DBAddr = redisAddr
		db.DBPort = int16(redisPort)

		txDb := tx.TxDbInstance()
		if txDb == nil {
			log.Log.E("TxDbInstance failed, stop the program!")
			os.Exit(1)
		}

		err := state.PoolInstance()
		if err != nil {
			log.Log.E("PoolInstance failed, stop the program! err:%v", err)
			os.Exit(1)
		}

		if state.StdPool == nil {
			log.Log.E("StdPool initialization failed, stop the program!")
			os.Exit(1)
		}

		blockChain, err := block.Instance()
		if err != nil {
			log.Log.E("NewBlockChain failed, stop the program! err:%v", err)
			os.Exit(1)
		}
		//检查db和redis数据是否合法
		var resBlockLength uint64
		// 最少有个创世块
		resBlockLength = 1
		bn, err := state.StdPool.Get(state.Key("BlockNum"))
		if err == nil {

			val, ok := bn.(*state.VInt)
			if !ok {
				log.Log.E("Redis BlockNum empty")
				state.StdPool.Put(state.Key("BlockNum"), state.MakeVInt(1))
				state.StdPool.Flush()
			} else {
				resBlockLength = uint64(val.ToInt()) + 1
			}

		}

		bcLen := blockChain.Length()
		log.Log.I("BlockNum on Redis: %v", resBlockLength-1)
		log.Log.I("BCLen: %v", bcLen)
		if bcLen < resBlockLength {
			// 重算
			resBlockLength = 0
			state.StdPool.Delete(state.Key("iost"))
			state.StdPool.Delete(state.Key("BlockNum"))
			state.StdPool.Flush()
		}

		if bcLen > resBlockLength {
			var blk *block.Block
			var i uint64
			for i = resBlockLength; i < bcLen; i++ {
				log.Log.I("Update StatePool for number: %v", i)
				blk = blockChain.GetBlockByNumber(i)
				if blk == nil {
					break
				}
				if i == 0 {
					newPool, err := verifier.ParseGenesis(blk.Content[0].Contract, state.StdPool)
					if err != nil {
						log.Log.E("Update StatePool failed, stop the program! err:%v", err)
						os.Exit(1)
					}
					newPool.Flush()
				} else {
					newPool, err := blockcache.StdBlockVerifier(blk, state.StdPool)
					if err != nil {
						log.Log.E("Update StatePool failed, stop the program! err:%v", err)
						os.Exit(1)
					}
					newPool.Flush()
				}
			}

			if bcLen > 0 {

				blockChain.CheckLength()
				b := blockChain.Top()
				if b != nil {
					state.StdPool.Put(state.Key("BlockNum"), state.MakeVInt(int(b.Head.Number)))
					state.StdPool.Put(state.Key("BlockHash"), state.MakeVByte(b.HeadHash()))
					state.StdPool.Flush()
				}
			}
		}
		//初始化网络
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
		/*      publicIP := common.GetPulicIP() */
		// if publicIP != "" && common.IsPublicIP(net.ParseIP(publicIP)) && listenAddr != "127.0.0.1" {
		// listenAddr = publicIP
		/* } */
		net, err := network.GetInstance(
			&network.NetConfig{
				LogPath:       logPath,
				NodeTablePath: nodeTablePath,
				NodeID:        nodeID,
				RegisterAddr:  regAddr,
				ListenAddr:    listenAddr},
			target,
			uint16(port))
		if err != nil {
			log.Log.E("Network initialization failed, stop the program! err:%v", err)
			os.Exit(1)
		}
		log.LocalID = net.(*network.RouterImpl).LocalID()
		serverExit = append(serverExit, net)

		//启动共识
		accSecKey := viper.GetString("account.sec-key")
		//fmt.Printf("account.sec-key:  %v\n", accSecKey)

		acc, err := account.NewAccount(common.Base58Decode(accSecKey))
		if err != nil {
			log.Log.E("NewAccount failed, stop the program! err:%v", err)
			os.Exit(1)
		}

		account.MainAccount = acc

		//fmt.Printf("account PubKey = %v\n", common.Base58Encode(acc.Pubkey))
		//fmt.Printf("account SecKey = %v\n", common.Base58Encode(acc.Seckey))
		log.Log.I("account ID = %v", acc.ID)

		//HowHsu_Debug
		log.Log.I("blockchain db length:%d\n", blockChain.Length())

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

		consensus, err := consensus.ConsensusFactory(
			consensus.CONSENSUS_POB,
			acc, blockChain, state.StdPool, witnessList)
		if err != nil {
			log.Log.E("consensus initialization failed, stop the program! err:%v", err)
			os.Exit(1)
		}

		consensus.Run()
		serverExit = append(serverExit, consensus)
		blockCache := consensus.BlockCache()
		txPool, err := txpool.NewTxPoolServer(blockCache, blockCache.OnBlockChan())
		if err != nil {
			log.Log.E("NewTxPoolServer failed, stop the program! err:%v", err)
			os.Exit(1)
		}

		txPool.Start()
		serverExit = append(serverExit, txPool)

		//启动RPC
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

		chainServer(blockChain)

		//等待推出信号
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

	//fmt.Println(cfgFile)
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		panic(err)
	}

}
