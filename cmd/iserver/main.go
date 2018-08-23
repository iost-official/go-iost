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

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/consensus"
	"github.com/iost-official/Go-IOS-Protocol/consensus/synchronizer"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	flag "github.com/spf13/pflag"
)

type ServerExit interface {
	Stop()
}

var serverExit []ServerExit

var (
	configfile = flag.StringP("config", "f", "", "Configuration `file`")
	help       = flag.BoolP("help", "h", false, "Display available options")
)

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

	ilog.Info("Config Information:\n%v", conf.YamlString())

	glb, err := global.New(conf)
	if err != nil {
		ilog.Fatal("create global failed. err=%v", err)
	}

	p2pService, err := p2p.NewNetService(conf.P2P)
	if err != nil {
		ilog.Fatal("network initialization failed, stop the program! err:%v", err)
	}
	err = p2pService.Start()
	if err != nil {
		ilog.Fatal("start p2pservice failed. err=%v", err)
	}

	serverExit = append(serverExit, p2pService)

	accSecKey := glb.Config().ACC.SecKey
	acc, err := account.NewAccount(common.Base58Decode(accSecKey))
	if err != nil {
		ilog.Fatal("NewAccount failed, stop the program! err:%v", err)
	}
	account.MainAccount = acc

	blkCache, err := blockcache.NewBlockCache(glb)
	if err != nil {
		ilog.Fatal("blockcache initialization failed, stop the program! err:%v", err)
	}

	sync, err := synchronizer.NewSynchronizer(glb, blkCache, p2pService)
	if err != nil {
		ilog.Fatal("synchronizer initialization failed, stop the program! err:%v", err)
	}
	err = sync.Start()
	if err != nil {
		ilog.Fatal("start synchronizer failed. err=%v", err)
	}
	serverExit = append(serverExit, sync)

	var txp txpool.TxPool
	txp, err = txpool.NewTxPoolImpl(glb, blkCache, p2pService)
	if err != nil {
		ilog.Fatal("txpool initialization failed, stop the program! err:%v", err)
	}
	txp.Start()
	serverExit = append(serverExit, txp)

	consensus, err := consensus.Factory(
		"pob",
		acc, glb, blkCache, txp, p2pService, sync, account.WitnessList) //witnessList)
	if err != nil {
		ilog.Fatal("consensus initialization failed, stop the program! err:%v", err)
	}
	consensus.Start()
	serverExit = append(serverExit, consensus)

	exitLoop()

}

func exitLoop() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	i := <-c
	ilog.Info("IOST server received interrupt[%v], shutting down...", i)
	for _, s := range serverExit {
		s.Stop()
	}
	ilog.Stop()
}
