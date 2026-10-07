// replay 工具：用本地快照状态作为起点，从源 BlockChainDB 逐块读取并用
// chainbase.Add 走生产级验证路径重放，专用于验证 goja VM 与主网（V8）
// 执行结果的共识一致性。遇到 receipt 不一致（如 gas 差异）即停止并打印
// 详细交易内容，修复后重新运行即可从上次的持久化高度继续。
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/iost-official/go-iost/v3/chainbase"
	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/iost-official/go-iost/v3/core/global"
	"github.com/iost-official/go-iost/v3/core/tx"
	"github.com/iost-official/go-iost/v3/core/version"
	"github.com/iost-official/go-iost/v3/ilog"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		dbPath  = flag.String("db", "data/replay/storage/", "重放目标的 storage 目录（内含快照解压出的 BlockChainDB/StateDB/BlockCacheWAL）")
		srcPath = flag.String("src", "data/srcdb/BlockChainDB", "源 BlockChainDB（从已同步的 V8 节点拷贝，提供待重放的块）")
		from    = flag.Int64("from", 0, "起始块高度（0 = 重放库当前 head+1）")
		to      = flag.Int64("to", 0, "结束块高度（0 = 源库 head）")
		limit   = flag.Int64("limit", 0, "最多处理块数（0 = 不限）")
	)
	flag.Parse()

	logger := ilog.New()
	consoleWriter := ilog.NewConsoleWriter()
	consoleWriter.SetLevel(ilog.LevelWarn) // 只看 Warn 以上（含 verify 失败详情）
	logger.AddWriter(consoleWriter)
	ilog.InitLogger(logger)

	conf := &common.Config{
		ACC:     &common.ACCConfig{ID: "replay", SecKey: "2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1", Algorithm: "ed25519"},
		DB:      &common.DBConfig{LdbPath: *dbPath},
		P2P:     &common.P2PConfig{ChainID: 1024},
		Version: &common.VersionConfig{NetName: "mainnet", ProtocolVersion: "1.0"},
	}
	global.SetGlobalConf(conf)
	version.InitChainConf(conf)
	tx.ChainID = conf.P2P.ChainID

	// 重放是离线验证：墙钟 deadline 只适用于在线出块/验证，历史块重放时
	// 机器负载造成的偶发超时（execution killed）是误杀，这里把时间上限放大。
	common.MaxBlockTimeLimit = time.Hour
	common.MaxTxTimeLimit = time.Hour

	src, err := block.NewBlockChain(*srcPath)
	if err != nil {
		fmt.Println("open source BlockChainDB failed:", err)
		return 1
	}
	srcLen := src.Length()
	fmt.Printf("source chain length: %d\n", srcLen)

	cBase, err := chainbase.New(conf)
	if err != nil {
		fmt.Println("chainbase init failed:", err)
		return 1
	}
	defer cBase.Close()

	head := cBase.HeadBlock().Head.Number + 1 // blockcache 恢复后的真实 head（含 WAL 块）
	start := *from
	if start == 0 {
		start = head
	}
	end := *to
	if end == 0 || end > srcLen {
		end = srcLen
	}
	fmt.Printf("replay range: [%d, %d), replay-chain head: %d\n", start, end, head)

	startTime := time.Now()
	lastTime := startTime
	lastN := start
	processed := int64(0)
	for n := start; n < end; n++ {
		if *limit > 0 && processed >= *limit {
			break
		}
		blk, err := src.GetBlockByNumber(n)
		if err != nil {
			fmt.Printf("STOP: source chain missing block %d: %v\n", n, err)
			break
		}
		ta := time.Now()
		if err := cBase.Add(blk, true, false); err != nil {
			fmt.Printf("\nDIVERGENCE at block %d: %v\n", n, err)
			for i, t := range blk.Txs {
				fmt.Printf("  tx[%d] %s\n", i, t.String())
			}
			return 2
		}
		processed++
		if d := time.Since(ta); d > 100*time.Millisecond {
			fmt.Printf("SLOW block %d took %v\n", n, d)
		}
		if time.Since(lastTime) > 5*time.Second {
			rate := float64(n-lastN) / time.Since(lastTime).Seconds()
			fmt.Printf("[%s] replayed to %d (%.0f blk/s)\n", time.Now().Format("15:04:05"), n, rate)
			lastTime = time.Now()
			lastN = n
		}
	}
	fmt.Printf("DONE: replayed %d blocks, no divergence in [%d, %d)\n", processed, start, end)
	return 0
}
