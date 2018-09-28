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

	"strconv"
	"strings"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus"
	"github.com/iost-official/go-iost/consensus/synchronizer"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/metrics"
	"github.com/iost-official/go-iost/p2p"
	"github.com/iost-official/go-iost/rpc"
	"github.com/iost-official/go-iost/vm"
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
	err := metrics.SetPusher(metricsConfig.PushAddr, metricsConfig.Username, metricsConfig.Password)
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
		*configfile = os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/config/iserver.yml"
	}

	conf := common.NewConfig(*configfile)

	initLogger(conf.Log)
	ilog.Infof("Config Information:\n%v", conf.YamlString())

	vm.SetUp(conf.VM)

	err := initMetrics(conf.Metrics)
	if err != nil {
		ilog.Errorf("init metrics failed. err=%v", err)
	}

	bv, err := global.New(conf)
	if err != nil {
		ilog.Fatalf("create global failed. err=%v", err)
	}
	if conf.Genesis.CreateGenesis {
		genesisBlock, _ := bv.BlockChain().GetBlockByNumber(0)
		ilog.Infof("createGenesisHash: %v", common.Base58Encode(genesisBlock.HeadHash()))
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

	blkCache, err := blockcache.NewBlockCache(bv)
	if err != nil {
		ilog.Fatalf("blockcache initialization failed, stop the program! err:%v", err)
	}

	sync, err := synchronizer.NewSynchronizer(bv, blkCache, p2pService)
	if err != nil {
		ilog.Fatalf("synchronizer initialization failed, stop the program! err:%v", err)
	}
	app = append(app, sync)

	var txp txpool.TxPool
	txp, err = txpool.NewTxPoolImpl(bv, blkCache, p2pService)
	if err != nil {
		ilog.Fatalf("txpool initialization failed, stop the program! err:%v", err)
	}
	app = append(app, txp)

	rpcServer := rpc.NewRPCServer(txp, blkCache, bv, p2pService)
	app = append(app, rpcServer)

	jsonRPCServer := rpc.NewJSONServer(bv)
	app = append(app, jsonRPCServer)
	consensus, err := consensus.Factory("pob", acc, bv, blkCache, txp, p2pService)
	if err != nil {
		ilog.Fatalf("consensus initialization failed, stop the program! err:%v", err)
	}
	app = append(app, consensus)

	err = app.Start()
	if err != nil {
		ilog.Fatalf("start iserver failed. err=%v", err)
	}

	if conf.Debug != nil {
		startDebugServer(conf.Debug.ListenAddr, blkCache, p2pService, bv.BlockChain())
	}

	waitExit()

	app.Stop()
	ilog.Stop()
	bv.BlockChain().Close()
	bv.StateDB().Close()
	bv.TxDB().Close()
}

func waitExit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	i := <-c
	ilog.Infof("IOST server received interrupt[%v], shutting down...", i)
}

func startDebugServer(addr string, blkCache blockcache.BlockCache, p2pService p2p.Service, blkChain block.Chain) {
	http.HandleFunc("/debug/blockcache/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte(blkCache.Draw()))
	})
	http.HandleFunc("/debug/blockchain/", func(rw http.ResponseWriter, r *http.Request) {
		rg := r.URL.Query()
		sp := strings.Split(rg["range"][0], "-")
		start, err := strconv.Atoi(sp[0])
		if err != nil {
			return
		}
		end, err := strconv.Atoi(sp[1])
		if err != nil {
			return
		}
		rw.Write([]byte(blkChain.Draw(int64(start), int64(end))))
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
