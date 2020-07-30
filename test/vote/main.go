package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/iost-official/go-iost/sdk"

	"github.com/iost-official/go-iost/account"
	rpcpb "github.com/iost-official/go-iost/rpc/pb"

	"github.com/iost-official/go-iost/iwallet"
)

var (
	iostSDKs     = make(map[string]*sdk.IOSTDevSDK)
	witness      = []string{}
	accounts     = []string{}
	server       = "localhost:30002"
	contractName = ""
	pledgeGAS    = int64(0)
	exchangeIOST = false
)

func init() {
	log.SetOutput(os.Stdout)
	rand.Seed(time.Now().Unix())
}

var signAlgo = "ed25519"

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
		iostSDK := sdk.NewIOSTDevSDK()
		iostSDK.SetChainID(1024)
		iostSDK.SetSignAlgo("ed25519")
		kp, err := iwallet.LoadKeyPair(a)
		if err != nil {
			panic(err)
		}
		iostSDK.SetAccount(a, kp)
		iostSDK.SetServer(server)
		iostSDK.SetTxInfo(2000000, 1, 300, 0, nil)
		iostSDK.SetCheckResult(true, 3, 10)
		iostSDK.SetVerbose(true)
		iostSDKs[a] = iostSDK
	}
}

func prepareAccounts() {
	iostSDK := iostSDKs["admin"]
	kp, err := iwallet.LoadKeyPair("admin")
	if err != nil {
		panic(err)
	}
	iostSDK.SetAccount("admin", kp)
	if pledgeGAS > 0 {
		err = iostSDK.PledgeForGasAndRAM(pledgeGAS, 0)
		if err != nil {
			log.Fatalf("pledge gas and ram err: %v", err)
		}
	}
	for _, acc := range accounts {
		newKp, err := account.NewKeyPair(nil, sdk.GetSignAlgoByName(signAlgo))
		if err != nil {
			log.Fatalf("create key pair failed %v", err)
		}
		k := newKp.ReadablePubkey()
		okey := k
		akey := k

		_, err = iostSDK.CreateNewAccount(acc, okey, akey, 1024, 1000, 2100000)
		if err != nil {
			log.Fatalf("create new account error %v", err)
		}
		err = iwallet.SaveAccount(acc, newKp)
		if err != nil {
			log.Fatalf("saveAccount failed %v", err)
		}
		iostSDKs[acc].SetAccount(acc, newKp)
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
	_, txHash, err := iostSDKs["admin"].PublishContract(codePath, abiPath, "", false, "")
	if err != nil {
		log.Fatalf("publish contract error: %v", err)
	}
	contractName = "Contract" + txHash
}

func vote() {
	for _, acc := range accounts {
		iostSDK := iostSDKs[acc]
		iostSDK.SendTxFromActions([]*rpcpb.Action{
			sdk.NewAction(contractName, "vote", fmt.Sprintf(`["%s","%s","%v"]`, acc, witness[rand.Intn(len(witness))], (rand.Intn(10)+2)*100000)),
		})
	}
}

func unvote() {
	for _, acc := range accounts {
		iostSDK := iostSDKs[acc]
		iostSDK.SendTxFromActions([]*rpcpb.Action{
			sdk.NewAction(contractName, "unvote", fmt.Sprintf(`["%s","%s","%v"]`, acc, witness[rand.Intn(len(witness))], (rand.Intn(10)+2)*1000)),
		})
	}
}

func issueIOST() {
	iostSDKs["admin"].SendTxFromActions([]*rpcpb.Action{
		sdk.NewAction(contractName, "issueIOST", `[]`),
	})
}

func withdrawBlockBonus() {
	if !exchangeIOST {
		return
	}
	for _, acc := range witness {
		iostSDK := iostSDKs[acc]
		iostSDK.SendTxFromActions([]*rpcpb.Action{
			sdk.NewAction(contractName, "exchangeIOST", `[]`),
		})
	}
}

func withdrawVoteBonus() {
	for _, acc := range witness {
		iostSDK := iostSDKs[acc]
		iostSDK.SendTxFromActions([]*rpcpb.Action{
			sdk.NewAction(contractName, "candidateWithdraw", `[]`),
		})
	}
}

func topupVoterBonus() {
	iostSDK := iostSDKs["admin"]
	for _, acc := range witness {
		iostSDK.SendTxFromActions([]*rpcpb.Action{
			sdk.NewAction(contractName, "topupVoterBonus", fmt.Sprintf(`["%v", "%v"]`, acc, (rand.Intn(10)+2)*100000)),
		})
	}
}

func withdrawVoterBonus() {
	for _, acc := range accounts {
		iostSDK := iostSDKs[acc]
		iostSDK.SendTxFromActions([]*rpcpb.Action{
			sdk.NewAction(contractName, "voterWithdraw", `[]`),
		})
	}
}

func checkResult() {
	iostSDKs["admin"].SendTxFromActions([]*rpcpb.Action{
		sdk.NewAction(contractName, "checkResult", `[]`),
	})
}
