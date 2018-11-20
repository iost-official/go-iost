package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/iost-official/go-iost/common"
	"gopkg.in/yaml.v2"
)

var (
	cluster = flag.String("c", "default", "The name of cluster")
	master  = flag.Int("m", 3, "The number of master node")
	slave   = flag.Int("s", 1, "The number of slave node")
)

func main() {
	//	num := flag.String("n", "47", "number")
	//	flag.Parse()
	//
	//	n, err := strconv.Atoi(*num)
	//	if err != nil {
	//		n = 47
	//	}
	//
	//	file, err := os.Create("keypairs2")
	//	if err != nil {
	//		panic(err)
	//	}
	//	defer file.Close()
	//
	//	for i := 0; i < n; i++ {
	//		ac, _ := account.NewAccount(nil, crypto.Ed25519)
	//		file.WriteString(fmt.Sprintf("%s,%s\n", account.GetIDByPubkey(ac.Pubkey), common.Base58Encode(ac.Seckey)))
	//	}

	flag.Parse()
	genconfig()
}

func genconfig() {
	pushAddr := os.Getenv("PROMETHEUS_HOSTNAME")
	username := os.Getenv("PROMETHEUS_USERNAME")
	password := os.Getenv("PROMETHEUS_PASSWORD")
	consoleLog := true
	seedNodes := []string{fmt.Sprintf("/dns4/iserver-0.iserver.%s.svc.cluster.local/tcp/30000/ipfs/12D3KooWA2QZHXCLsVL9rxrtKPRqBSkQj7mCdHEhRoW8eJtn24ht", *cluster)}

	file, err := os.Open("keypairs")
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	ids := make([]string, 0)
	seckeys := make([]string, 0)
	//initialCoins := make([]string, 0)
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), ",")
		ids = append(ids, s[0])
		seckeys = append(seckeys, s[1])
		//initialCoins = append(initialCoins, s[2])
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	file.Close()

	WitnessInfo := make([]*common.Witness, 0)
	for i := 0; i < *master; i++ {
		witness := &common.Witness{
			ID:      fmt.Sprintf("producer%05d", i),
			Owner:   ids[i],
			Active:  ids[i],
			Balance: int64(1000000000),
		}
		WitnessInfo = append(WitnessInfo, witness)
	}

	adminInfo := &common.Witness{
		ID:      "admin",
		Owner:   "IOST2mCzj85xkSvMf1eoGtrexQcwE6gK8z5xr6Kc48DwxXPCqQJva4",
		Active:  "IOST2mCzj85xkSvMf1eoGtrexQcwE6gK8z5xr6Kc48DwxXPCqQJva4",
		Balance: int64(21000000000) - int64(1000000000)*int64(*master),
	}

	tokenInfo := &common.TokenInfo{
		FoundationAccount: "foundation",
		IOSTTotalSupply:   90000000000,
		IOSTDecimal:       8,
		RAMTotalSupply:    9000000000000000000,
		RAMGenesisAmount:  137438953472,
	}

	foundationInfo := &common.Witness{
		ID:      "foundation",
		Owner:   "IOST2mCzj85xkSvMf1eoGtrexQcwE6gK8z5xr6Kc48DwxXPCqQJva4",
		Active:  "IOST2mCzj85xkSvMf1eoGtrexQcwE6gK8z5xr6Kc48DwxXPCqQJva4",
		Balance: 0,
	}

	Genesis := &common.GenesisConfig{
		CreateGenesis:    true,
		InitialTimestamp: "2018-01-02T15:04:03Z",
		TokenInfo:        tokenInfo,
		WitnessInfo:      WitnessInfo,
		AdminInfo:        adminInfo,
		FoundationInfo:   foundationInfo,
		ContractPath:     "contract/",
	}

	genesisfile, err := os.Create("genesis.yml")
	if err != nil {
		log.Fatal(err)
	}
	bs, err := yaml.Marshal(Genesis)
	if err != nil {
		log.Fatalf("Unable to marshal config to YAML: %v", err)
	}
	genesisfile.WriteString(string(bs))
	genesisfile.Close()

	VM := &common.VMConfig{
		JsPath:   "vm/v8vm/v8/libjs/",
		LogLevel: "",
	}
	DB := &common.DBConfig{
		LdbPath: "/data/storage/",
	}

	P2P := &common.P2PConfig{
		ListenAddr: "0.0.0.0:30000",
		SeedNodes:  seedNodes,
		ChainID:    1024,
		Version:    1,
		DataPath:   "/data/p2p/",
	}
	RPC := &common.RPCConfig{
		JSONPort: 30001,
		GRPCPort: 30002,
	}
	Log := &common.LogConfig{
		FileLog: &common.FileLogConfig{
			Path:   "/data/logs/",
			Level:  "info",
			Enable: false,
		},
		ConsoleLog: &common.ConsoleLogConfig{
			Level:  "info",
			Enable: consoleLog,
		},
		AsyncWrite: true,
	}
	Debug := &common.DebugConfig{
		ListenAddr: "0.0.0.0:30003",
	}

	for i := 0; i < *master+*slave; i++ {
		ACC := &common.ACCConfig{
			ID:        fmt.Sprintf("producer%05d", i),
			SecKey:    seckeys[i],
			Algorithm: "ed25519",
		}
		Metrics := &common.MetricsConfig{
			PushAddr: pushAddr,
			Enable:   true,
			ID:       *cluster + ":node-" + fmt.Sprintf("%d", i),
			Username: username,
			Password: password,
		}
		c := &common.Config{
			ACC:     ACC,
			Genesis: "/var/lib/iserver/genesis.yml",
			VM:      VM,
			DB:      DB,
			P2P:     P2P,
			RPC:     RPC,
			Log:     Log,
			Metrics: Metrics,
			Debug:   Debug,
		}

		if i == 0 {
			c.P2P.DataPath = "/var/lib/iserver/"
		} else {
			c.P2P.DataPath = "/data/p2p/"
		}

		file, err := os.Create(fmt.Sprintf("iserver-%d.yml", i))
		if err != nil {
			log.Fatal(err)
		}
		file.WriteString(c.YamlString())
		file.Close()
	}
}
