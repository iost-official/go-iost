// gastrace：对指定块中的指定交易做单交易执行追踪。配合沙箱的
// IOST_GAS_TRACE 钩子（见 vm/v8vm/sandbox.go），输出 gas 计费直方图，
// 用于对比 V8 与 goja 两个引擎对同一交易的计费序列差异。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/iost-official/go-iost/v3/core/global"
	"github.com/iost-official/go-iost/v3/core/tx"
	"github.com/iost-official/go-iost/v3/core/version"
	"github.com/iost-official/go-iost/v3/db"
	"github.com/iost-official/go-iost/v3/ilog"
	"github.com/iost-official/go-iost/v3/vm"
	"github.com/iost-official/go-iost/v3/vm/database"
)

func main() {
	statePath := flag.String("state", "data/replay/storage/StateDB", "StateDB 路径（其当前 tag 应为待测块的父块状态）")
	srcPath := flag.String("src", "data/srcdb/BlockChainDB", "源 BlockChainDB 路径")
	num := flag.Int64("block", 0, "块高度")
	txIdx := flag.Int("tx", 1, "交易在块内的下标")
	isBase := flag.Bool("base", false, "按 base tx 流程执行（TriggerBlockBaseMode）")
	out := flag.String("o", "", "直方图输出文件（默认打印）")
	flag.Parse()

	os.Setenv("IOST_GAS_TRACE", "1")

	logger := ilog.New()
	consoleWriter := ilog.NewConsoleWriter()
	consoleWriter.SetLevel(ilog.LevelError)
	logger.AddWriter(consoleWriter)
	ilog.InitLogger(logger)

	conf := &common.Config{
		ACC:     &common.ACCConfig{ID: "trace", SecKey: "2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1", Algorithm: "ed25519"},
		P2P:     &common.P2PConfig{ChainID: 1024},
		Version: &common.VersionConfig{NetName: "mainnet", ProtocolVersion: "1.0"},
	}
	global.SetGlobalConf(conf)
	version.InitChainConf(conf)
	tx.ChainID = conf.P2P.ChainID

	src, err := block.NewBlockChain(*srcPath)
	if err != nil {
		fmt.Println("open source chain failed:", err)
		os.Exit(1)
	}
	blk, err := src.GetBlockByNumber(*num)
	if err != nil {
		fmt.Println("get block failed:", err)
		os.Exit(1)
	}

	stateDB, err := db.NewMVCCDB(*statePath)
	if err != nil {
		fmt.Println("open statedb failed:", err)
		os.Exit(1)
	}
	defer stateDB.Close()

	t := blk.Txs[*txIdx]
	want := blk.Receipts[*txIdx]
	fmt.Printf("block %d tx[%d] by %s: %s\n", blk.Head.Number, *txIdx, t.Publisher, t.Actions[0])
	fmt.Printf("canonical gas: %d\n", want.GasUsage)

	vi := database.NewBatchVisitor(database.NewBatchVisitorRoot(100, stateDB, blk.Head.Rules()))
	isolator := &vm.Isolator{}
	var lg ilog.Logger
	lg.Stop()
	if err := isolator.Prepare(blk.Head, vi, &lg); err != nil {
		fmt.Println("isolator prepare failed:", err)
		os.Exit(1)
	}
	if *isBase {
		isolator.TriggerBlockBaseMode()
	}
	to := common.MaxTxTimeLimit * 50
	if err := isolator.PrepareTx(t, to); err != nil {
		fmt.Println("prepare tx failed:", err)
		os.Exit(1)
	}
	start := time.Now()
	_, err = isolator.Run()
	fmt.Printf("run took %v, err=%v\n", time.Since(start), err)
	receipt, err := isolator.PayCost()
	if err != nil {
		fmt.Println("paycost failed:", err)
		os.Exit(1)
	}
	fmt.Printf("actual gas: %d (diff %d)\n", receipt.GasUsage, receipt.GasUsage-want.GasUsage)

	// 完整 receipt 落盘，用于双引擎值对比（状态/返回/事件内容）
	rd, _ := json.Marshal(receipt)
	rcPath := *out + ".receipt.json"
	if *out != "" {
		os.WriteFile(rcPath, rd, 0644)
		fmt.Println("receipt written to", rcPath)
	}

	// 从 receipt 的 returns 里取出直方图（Execute 的返回值被包装进 returns）
	for _, r := range receipt.Returns {
		idx := strings.Index(r, "GASTRACE:")
		if idx < 0 {
			continue
		}
		s := r[idx+len("GASTRACE:"):]
		if end := strings.LastIndex(s, `"]`); end >= 0 {
			s = s[:end]
		}
		s = strings.ReplaceAll(s, `\"`, `"`)
		hist := map[string]int64{}
		if err := json.Unmarshal([]byte(s), &hist); err == nil {
			data, _ := json.MarshalIndent(hist, "", "  ")
			if *out != "" {
				os.WriteFile(*out, data, 0644)
				fmt.Println("histogram written to", *out)
			} else {
				fmt.Println(string(data))
			}
		} else {
			fmt.Println("parse histogram failed:", err, "\nraw:", r[:min(200, len(r))])
		}
	}
}
