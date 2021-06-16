package main

import (
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/global"
	"github.com/iost-official/go-iost/v3/core/version"
	"github.com/iost-official/go-iost/v3/ilog"
	"github.com/iost-official/go-iost/v3/iserver"
	"github.com/iost-official/go-iost/v3/metrics"
	flag "github.com/spf13/pflag"
)

var (
	configFile = flag.StringP("config", "f", "", "Configuration `file`")
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
		consoleWriter.SetLevel(ilog.NewLevel(logConfig.ConsoleLog.Level))
		logger.AddWriter(consoleWriter)
	}
	if logConfig.FileLog != nil && logConfig.FileLog.Enable {
		fileWriter := ilog.NewFileWriter(logConfig.FileLog.Path)
		fileWriter.SetLevel(ilog.NewLevel(logConfig.FileLog.Level))
		logger.AddWriter(fileWriter)
	}
	ilog.InitLogger(logger)
}

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
	}

	if *configFile == "" {
		repoDir, ok := os.LookupEnv("GOBASE")
		if !ok {
			repoDir = "."
		}
		*configFile = repoDir + "/config/iserver.yml"
	}

	conf := common.NewConfig(*configFile)

	global.SetGlobalConf(conf)
	version.InitChainConf(conf)

	initLogger(conf.Log)

	ilog.Infof("Config Information:\n%v", strings.ReplaceAll(conf.YamlString(), conf.ACC.SecKey, conf.ACC.SecKey[:3]+"******"))

	ilog.Infof("build time:%v", global.BuildTime)
	ilog.Infof("git hash:%v", global.GitHash)
	ilog.Infof("code version:%v", global.CodeVersion)

	err := initMetrics(conf.Metrics)
	if err != nil {
		ilog.Errorf("init metrics failed. err=%v", err)
	}

	server := iserver.New(conf)
	err = server.Start()
	if err != nil {
		ilog.Fatalf("start iserver failed. err=%v", err)
	} else {
		ilog.Info("iserver started")
	}

	waitExit()

	server.Stop()
	ilog.Stop()
}

func waitExit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	i := <-c
	ilog.Infof("IOST server received interrupt[%v], shutting down...", i)
}
