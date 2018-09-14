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
	"encoding/json"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/consensus"
	"github.com/iost-official/Go-IOS-Protocol/consensus/synchronizer"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/txpool"
	"github.com/iost-official/Go-IOS-Protocol/crypto"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/metrics"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"github.com/iost-official/Go-IOS-Protocol/rpc"
	"github.com/iost-official/Go-IOS-Protocol/vm"
	flag "github.com/spf13/pflag"
)

var (
	configfile = flag.StringP("config", "f", "", "Configuration `file`")
	help       = flag.BoolP("help", "h", false, "Display available options")
)

func initMetrics(metricsConfig *common.MetricsConfig) error {
	if metricsConfig == nil || !metricsConfig.Enable {
		return nil
	}
	err := metrics.SetPusher(metricsConfig.PushAddr)
	if err != nil {
		return err
	}
	metrics.SetID(metricsConfig.ID)
	return metrics.Start()
}

func getLogLevel(l string) ilog.Level {
	switch l {
	case "debug":
		return ilog.LevelDebug
	case "info":
		return ilog.LevelInfo
	case "warn":
		return ilog.LevelWarn
	case "error":
		return ilog.LevelError
	case "fatal":
		return ilog.LevelFatal
	default:
		return ilog.LevelDebug
	}
}

func initLogger(logConfig *common.LogConfig) {
	if logConfig == nil {
		return
	}
	logger := ilog.New()
	if logConfig.AsyncWrite {
		logger.AsyncWrite()
	}
	if logConfig.ConsoleLog != nil && logConfig.ConsoleLog.Enable {
		consoleWriter := ilog.NewConsoleWriter()
		consoleWriter.SetLevel(getLogLevel(logConfig.ConsoleLog.Level))
		logger.AddWriter(consoleWriter)
	}
	if logConfig.FileLog != nil && logConfig.FileLog.Enable {
		fileWriter := ilog.NewFileWriter(logConfig.FileLog.Path)
		fileWriter.SetLevel(getLogLevel(logConfig.FileLog.Level))
		logger.AddWriter(fileWriter)
	}
	ilog.InitLogger(logger)
}

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
	}

	if *configfile == "" {
		*configfile = os.Getenv("GOPATH") + "/src/github.com/iost-official/Go-IOS-Protocol/config/iserver.yaml"
	}

	conf := common.NewConfig(*configfile)

	initLogger(conf.Log)
	ilog.Infof("Config Information:\n%v", conf.YamlString())

	vm.SetUp(conf.VM)

	err := initMetrics(conf.Metrics)
	if err != nil {
		ilog.Errorf("init metrics failed. err=%v", err)
	}

	glb, err := global.New(conf)
	if err != nil {
		ilog.Fatalf("create global failed. err=%v", err)
	}
	if conf.Genesis.CreateGenesis {
		genesisBlock, _ := glb.BlockChain().GetBlockByNumber(0)
		ilog.Errorf("createGenesisHash: %v", common.Base58Encode(genesisBlock.HeadHash()))
	}
	var app common.App

	p2pService, err := p2p.NewNetService(conf.P2P)
	if err != nil {
		ilog.Fatalf("network initialization failed, stop the program! err:%v", err)
	}
	app = append(app, p2pService)

	accSecKey := conf.ACC.SecKey
	acc, err := account.NewAccount(common.Base58Decode(accSecKey), getSignAlgo(conf.ACC.Algorithm))
	if err != nil {
		ilog.Fatalf("NewAccount failed, stop the program! err:%v", err)
	}

	blkCache, err := blockcache.NewBlockCache(glb)
	if err != nil {
		ilog.Fatalf("blockcache initialization failed, stop the program! err:%v", err)
	}

	sync, err := synchronizer.NewSynchronizer(glb, blkCache, p2pService)
	if err != nil {
		ilog.Fatalf("synchronizer initialization failed, stop the program! err:%v", err)
	}
	app = append(app, sync)

	var txp txpool.TxPool
	txp, err = txpool.NewTxPoolImpl(glb, blkCache, p2pService)
	if err != nil {
		ilog.Fatalf("txpool initialization failed, stop the program! err:%v", err)
	}
	app = append(app, txp)

	rpcServer := rpc.NewRPCServer(txp, blkCache, glb, p2pService)
	app = append(app, rpcServer)

	jsonRPCServer := rpc.NewJSONServer(glb)
	app = append(app, jsonRPCServer)
	consensus, err := consensus.Factory("pob", acc, glb, blkCache, txp, p2pService, sync)
	if err != nil {
		ilog.Fatalf("consensus initialization failed, stop the program! err:%v", err)
	}
	app = append(app, consensus)

	err = app.Start()
	if err != nil {
		ilog.Fatal("start iserver failed. err=%v", err)
	}

	if conf.Debug != nil {
		startDebugServer(conf.Debug.ListenAddr, blkCache, p2pService)
	}

	waitExit()

	app.Stop()
	ilog.Stop()
}

func waitExit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	i := <-c
	ilog.Infof("IOST server received interrupt[%v], shutting down...", i)
}

func startDebugServer(addr string, blkCache blockcache.BlockCache, p2pService p2p.Service) {
	http.HandleFunc("/debug/blockcache/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte(blkCache.Draw()))
	})
	http.HandleFunc("/debug/p2p/neighbors/", func(rw http.ResponseWriter, r *http.Request) {
		neighbors := p2pService.NeighborStat()
		bytes, _ := json.MarshalIndent(neighbors, "", "    ")
		rw.Write(bytes)
	})

	go func() {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			ilog.Errorf("start debug server failed. err=%v", err)
		}
	}()
}

func getSignAlgo(algo string) crypto.Algorithm {
	switch algo {
	case "secp256k1":
		return crypto.Secp256k1
	case "ed25519":
		return crypto.Ed25519
	default:
		return crypto.Ed25519
	}
}
