package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/rpc/pb"

	"github.com/iost-official/go-iost/iwallet"
)

var (
	sdks         = make(map[string]*iwallet.SDK)
	witness      = []string{}
	accounts     = []string{}
	server       = "localhost:30002"
	amountLimit  = "*:unlimited"
	contractName = ""
	pledgeGAS    = int64(0)
	exchangeIOST = false
)

func init() {
	log.SetOutput(os.Stdout)
	rand.Seed(time.Now().Unix())
}

func parseFlag() {
	s := flag.String("s", server, "rpc server")        // format: ip1:port1,ip2:port2
	a := flag.String("a", "", "account names")         // format: acc1,acc2
	w := flag.String("w", "", "witness account names") // format: acc1,acc2
	p := flag.Int64("p", 0, "pledge gas for admin")    // format: 1234
	e := flag.Bool("e", false, "call exchangeIOST")    // format: true
	flag.Parse()

	server = *s
	accounts = strings.Split(*a, ",")
	witness = strings.Split(*w, ",")
	pledgeGAS = *p
	exchangeIOST = *e

	if *a == "" {
		log.Fatalf("flag a is required")
	}
	if *w == "" {
		log.Fatalf("flag w is required")
	}
}

func initSDKs() {
	accs := append(accounts, witness...)
	accs = append(accs, "admin")
	for _, a := range accs {
		sdk := &iwallet.SDK{}
		sdk.SetSignAlgo("ed25519")
		sdk.SetAccount(a, nil)
		sdk.SetServer(server)
		sdk.SetAmountLimit(amountLimit)
		sdk.SetTxInfo(2000000, 1, 300, 0)
		sdk.SetCheckResult(true, 3, 10)
		sdk.SetVerbose(true)
		sdk.LoadAccount()
		sdks[a] = sdk
	}
}

func prepareAccounts() {
	sdk := sdks["admin"]
	err := sdk.LoadAccount()
	if err != nil {
		log.Fatalf("load account failed: %v.", err)
	}
	if pledgeGAS > 0 {
		err = sdk.PledgeForGasAndRAM(pledgeGAS, 0)
		if err != nil {
			log.Fatalf("pledge gas and ram err: %v", err)
		}
	}
	aLgo := sdk.GetSignAlgo()
	for _, acc := range accounts {
		if err := sdks[acc].LoadAccount(); err == nil {
			continue
		}
		newKp, err := account.NewKeyPair(nil, aLgo)
		if err != nil {
			log.Fatalf("create key pair failed %v", err)
		}
		k := newKp.ReadablePubkey()
		okey := k
		akey := k

		_, err = sdk.CreateNewAccount(acc, okey, akey, 1024, 1000, 2100000)
		if err != nil {
			log.Fatalf("create new account error %v", err)
		}
		err = sdk.SaveAccount(acc, newKp)
		if err != nil {
			log.Fatalf("saveAccount failed %v", err)
		}
		sdks[acc].LoadAccount()
	}
}

func main() {
	parseFlag()
	initSDKs()
	prepareAccounts()
	run()
}

func run() {
	publish()
	vote()
	issueIOST()
	withdrawBlockBonus()
	withdrawVoteBonus()
	unvote()
	topupVoterBonus()
	withdrawVoterBonus()
	checkResult()
}

func publish() {
	codePath := os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/test/vote/test_data/vote_checker.js"
	abiPath := codePath + ".abi"
	_, txHash, err := sdks["admin"].PublishContract(codePath, abiPath, "", false, "")
	if err != nil {
		log.Fatalf("publish contract error: %v", err)
	}
	contractName = "Contract" + txHash
}

func vote() {
	for _, acc := range accounts {
		sdk := sdks[acc]
		sdk.SendTx([]*rpcpb.Action{
			iwallet.NewAction(contractName, "vote", fmt.Sprintf(`["%s","%s","%v"]`, acc, witness[rand.Intn(len(witness))], (rand.Intn(10)+2)*100000)),
		})
	}
}

func unvote() {
	for _, acc := range accounts {
		sdk := sdks[acc]
		sdk.SendTx([]*rpcpb.Action{
			iwallet.NewAction(contractName, "unvote", fmt.Sprintf(`["%s","%s","%v"]`, acc, witness[rand.Intn(len(witness))], (rand.Intn(10)+2)*1000)),
		})
	}
}

func issueIOST() {
	sdks["admin"].SendTx([]*rpcpb.Action{
		iwallet.NewAction(contractName, "issueIOST", `[]`),
	})
}

func withdrawBlockBonus() {
	if !exchangeIOST {
		return
	}
	for _, acc := range witness {
		sdk := sdks[acc]
		sdk.SendTx([]*rpcpb.Action{
			iwallet.NewAction(contractName, "exchangeIOST", `[]`),
		})
	}
}

func withdrawVoteBonus() {
	for _, acc := range witness {
		sdk := sdks[acc]
		sdk.SendTx([]*rpcpb.Action{
			iwallet.NewAction(contractName, "candidateWithdraw", `[]`),
		})
	}
}

func topupVoterBonus() {
	sdk := sdks["admin"]
	for _, acc := range witness {
		sdk.SendTx([]*rpcpb.Action{
			iwallet.NewAction(contractName, "topupVoterBonus", fmt.Sprintf(`["%v", "%v"]`, acc, (rand.Intn(10)+2)*100000)),
		})
	}
}

func withdrawVoterBonus() {
	for _, acc := range accounts {
		sdk := sdks[acc]
		sdk.SendTx([]*rpcpb.Action{
			iwallet.NewAction(contractName, "voterWithdraw", `[]`),
		})
	}
}

func checkResult() {
	sdks["admin"].SendTx([]*rpcpb.Action{
		iwallet.NewAction(contractName, "checkResult", `[]`),
	})
}
