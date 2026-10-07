// dumpstate：从重放 StateDB 导出指定合约的代码（用于 gas 差异分析）。
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/db"
)

func main() {
	dbPath := flag.String("db", "data/replay/storage/StateDB", "StateDB 路径")
	cid := flag.String("c", "", "合约 ID")
	out := flag.String("o", "", "输出文件（默认打印到 stdout）")
	flag.Parse()

	m, err := db.NewMVCCDB(*dbPath)
	if err != nil {
		fmt.Println("open statedb failed:", err)
		os.Exit(1)
	}
	defer m.Close()

	tag := m.CurrentTag()
	fmt.Fprintf(os.Stderr, "state tag: %s\n", tag)

	code, err := m.Get("state", "c-"+*cid)
	if err != nil {
		fmt.Println("get contract failed:", err)
		os.Exit(1)
	}
	con := &contract.Contract{}
	if err := con.Decode(code); err != nil {
		fmt.Println("decode contract failed:", err)
		os.Exit(1)
	}
	if *out == "" {
		fmt.Println(con.Code)
	} else {
		os.WriteFile(*out, []byte(con.Code), 0644)
		fmt.Fprintf(os.Stderr, "written to %s (%d bytes)\n", *out, len(con.Code))
	}
}
