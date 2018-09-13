package common

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

func TestConfig_YamlString(t *testing.T) {
	file, err := os.Open("keypairs")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	i := 1
	for scanner.Scan() {
		var c Config
		s := strings.Split(scanner.Text(), ",")
		id, seckey := s[0], s[1]
		c.ACC = &ACCConfig{id, seckey, "ed25519"}
		c.Genesis = &GenesisConfig{true, "", []string{
			"IOSTrhxdtgy6dCib3vNkejRK49NjxpFcymUzpb7ubdAktqWQsecNF",
			"13100000000",
			"IOST4StEtgZmyc1WEQAYkBYRtn7qBhJZ41YKioVVUKGnmpiBCRxDr",
			"13200000000",
			"IOST2EZtRrnKcZWCwSujnpbV4rhLihSM5KvXNt5YpeJF3Y3NPXpWf2",
			"13300000000",
			"IOST2un4eBwRSAUzakMLrdvz61DsnNRM53662ZPUJcKsxLvYMndMEc",
			"13400000000",
			"IOSTxqz5hZQVCyq3CJmtPjFcHX3HZ3Nr8zvo6Z9nQpkoXuxtxDskk",
			"13500000000",
			"IOSTdrxMsd2NBnAd55XF8dUhoadBJS6xtabn384jBzng93PG49Bf1",
			"13600000000",
			"IOST2Y47qBC4XZZG3irLsijCyNn149c2Ty8aAxMtxJKyCFb49Uyv9Z",
			"13700000000",
			"IOST2dwiomGQXdDAaRfmZ4mhjXNcoFxFSKpsN9HsRdS9MH51gsfhxY",
			"13800000000",
			"IOST2qcwiH4pBzKDqus9SqvMfW8CN9UzSayoA8vhyLWSZd8c89FVg6",
			"13900000000",
			"IOST21WFjmEzujV9zWPhUrXaYGSyWYd8U8xzCgUf23G9YkiPVuiqPc",
			"14000000000",
			"IOST22zX7xJTVA8wkmVi1hgNYdwJ8BuYTSC6VVQ7Va8CMQQcXBz66p",
			"14100000000",
		},
			"config/",
		}
		c.VM = &VMConfig{"vm/v8vm/v8/libjs/", ""}
		c.DB = &DBConfig{"/var/lib/iserver/leveldb/"}
		c.P2P = &P2PConfig{"0.0.0.0:30000", []string{"/ip4/13.237.151.211/tcp/30000/ipfs/12D3KooWA2QZHXCLsVL9rxrtKPRqBSkQj7mCdHEhRoW8eJtn24ht"}, 1024, 1, "/var/lib/iserver/p2p/"}
		c.RPC = &RPCConfig{30001, 30002}
		c.Log = &LogConfig{
			&FileLogConfig{"/var/lib/iserver/logs/", "info", true},
			&ConsoleLogConfig{"info", true},
			true,
		}
		c.Metrics = &MetricsConfig{"47.75.42.25:9091", true, "iost-test:node" + fmt.Sprintf("%02d", i)}
		c.Debug = &DebugConfig{"0.0.0.0:30003"}
		file, err := os.Create("/Users/zhouxiao/go/src/github.com/iost-official/iost-devops/playbook/config/iost-testnet/node" + fmt.Sprintf("%02d", i) + ".yml")
		if err != nil {
			panic(err)
		}
		file.WriteString(c.YamlString())
		i++
		file.Close()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
