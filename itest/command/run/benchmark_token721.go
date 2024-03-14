package run

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/iost-official/go-iost/v3/core/tx"
	"github.com/iost-official/go-iost/v3/ilog"
	"github.com/iost-official/go-iost/v3/itest"
	"github.com/urfave/cli/v2"
)

// BenchmarkToken721Command is the subcommand for benchmark.
var BenchmarkToken721Command = &cli.Command{
	Name:    "benchmarkToken721",
	Aliases: []string{"benchT721"},
	Usage:   "Run token benchmark by given tps",
	Flags:   BenchmarkToken721Flags,
	Action:  BenchmarkToken721Action,
}

// BenchmarkToken721Flags is the list of flags for benchmark.
var BenchmarkToken721Flags = []cli.Flag{
	&cli.IntFlag{
		Name:  "tps",
		Value: 100,
		Usage: "The expected ratio of transactions per second",
	},
	&cli.BoolFlag{
		Name:  "check",
		Usage: "if check receipt",
	},
}

const (
	createToken721        = "create"
	issueToken721         = "issue"
	transferToken721      = "transfer"
	balanceOfToken721     = "balanceOf"
	ownerOfToken721       = "ownerOf"
	tokenOfOwnerToken721  = "tokenOfOwnerByIndex"
	tokenMetadataToken721 = "tokenMetadata"
)

type token721Info struct {
	sym     string
	issuer  string
	balance map[string][]string
	acclist []string
	supply  int
}

// BenchmarkToken721Action is the action of benchmark.
var BenchmarkToken721Action = func(c *cli.Context) error {
	itest.Interval = 1000 * time.Millisecond
	itest.InitAmount = "1000"
	itest.InitPledge = "1000"
	itest.InitRAM = "3000"
	logger := ilog.New()
	fileWriter := ilog.NewFileWriter(c.String("log"))
	fileWriter.SetLevel(ilog.LevelInfo)
	logger.AddWriter(fileWriter)
	ilog.InitLogger(logger)
	//ilog.SetLevel(ilog.LevelDebug)
	it, err := itest.Load(c.String("keys"), c.String("config"))
	if err != nil {
		return err
	}
	accountFile := c.String("account")
	accounts, err := itest.LoadAccounts(accountFile)
	if err != nil {
		if err := AccountCaseAction(c); err != nil {
			return err
		}
		if accounts, err = itest.LoadAccounts(accountFile); err != nil {
			return err
		}
	}
	accountMap := make(map[string]*itest.Account)
	for _, acc := range accounts {
		accountMap[acc.ID] = acc
	}
	tps := c.Int("tps")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	startTime := time.Now()
	ticker := time.NewTicker(time.Second)
	counter := 0
	total := 0
	slotTotal := 0
	slotStartTime := startTime
	issueNumber := 5
	checkReceiptConcurrent := 64

	tokenList := []string{}
	tokenMap := make(map[string]*token721Info)
	tokenPrefix := "t" + strconv.FormatInt(time.Now().UnixNano(), 10)[14:]
	tokenOffset := 0
	var tokenMutex sync.Mutex

	hashCh := make(chan *hashItem, 4*tps*int(itest.Timeout.Seconds()))

	for c := 0; c < checkReceiptConcurrent; c++ {
		go func(hashCh chan *hashItem) {
			counter := 0
			failedCounter := 0
			for item := range hashCh {
				client := it.GetClients()[rand.Intn(len(it.GetClients()))]
				r, err := client.CheckTransactionWithTimeout(item.hash, item.expire)
				ilog.Debugf("receipt: %v", r)
				counter++
				if err != nil {
					ilog.Errorf("check transaction failed, %v", err)
					failedCounter++
				} else {
					for i := 0; i < len(r.Receipts); i++ {
						if r.Receipts[i].FuncName == "token721.iost/issue" {
							args := make([]string, 3)
							err := json.Unmarshal([]byte(r.Receipts[i].Content), &args)
							if err != nil {
								continue
							}
							ilog.Debugf("got receipt %v %v", r.Receipts[i], args)
							tokenMutex.Lock()
							tokenSym := args[0]
							acc := args[1]
							for j := 0; j < len(r.Returns); j++ {
								ret := r.Returns[j]
								ret = ret[2:(len(ret) - 2)]
								if _, ok := tokenMap[tokenSym].balance[acc]; !ok {
									tokenMap[tokenSym].balance[acc] = make([]string, 0)
									tokenMap[tokenSym].acclist = append(tokenMap[tokenSym].acclist, acc)
								}
								tokenMap[tokenSym].balance[acc] = append(tokenMap[tokenSym].balance[acc], ret)
								retn, _ := strconv.ParseInt(ret, 10, 32)
								tokenMap[tokenSym].supply = int(math.Max(float64(tokenMap[tokenSym].supply), float64(retn)))
							}
							tokenMutex.Unlock()
							break
						} else if r.Receipts[i].FuncName == "token721.iost/create" {
							args := make([]any, 3)
							err := json.Unmarshal([]byte(r.Receipts[i].Content), &args)
							if err != nil {
								continue
							}
							ilog.Debugf("got receipt %v %v", r.Receipts[i], args)
							tokenMutex.Lock()
							tokenSym := args[0].(string)
							issuer := args[1].(string)
							tokenList = append(tokenList, tokenSym)
							tokenMap[tokenSym] = &token721Info{
								sym:     tokenSym,
								issuer:  issuer,
								balance: make(map[string][]string),
								acclist: []string{},
								supply:  0,
							}
							tokenMutex.Unlock()
							break
						}
					}
				}
				if counter%1000 == 0 {
					ilog.Warnf("check %v transaction, %v successful, %v failed. channel size %v", counter, counter-failedCounter, failedCounter, len(hashCh))
				}
				if len(hashCh) > 3*tps*int(itest.Timeout.Seconds()) {
					ilog.Infof("hash ch size too large %v", len(hashCh))
				}
			}
		}(hashCh)
	}

	check := c.Bool("check")
	contractName := "token721.iost"
	for {
		trxs := make([]*itest.Transaction, 0)
		errList := []error{}
		tokenMutex.Lock()
		for num := 0; num < tps; num++ {
			// create 1, issue 1000, transfer 1000, balanceOf 100, ownerOf 100, tokenOfOwner 100, tokenMetadata 100
			tIndex := rand.Intn(2400)
			var abiName string
			switch {
			case tIndex <= 0 || len(tokenList) < 5:
				abiName = createToken721
				tokenSym := tokenPrefix + strconv.FormatInt(int64(tokenOffset), 10)
				tokenOffset++
				from := accounts[rand.Intn(len(accounts))]

				act0 := tx.NewAction("ram.iost", "buy", fmt.Sprintf(`["%v", "%v", %v]`, "admin", from.ID, 1000))
				act1 := tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, "admin", from.ID, 10))
				tx0 := itest.NewTransaction([]*tx.Action{act0, act1})
				trx, err := it.GetDefaultAccount().Sign(tx0)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}

				act1 = tx.NewAction(contractName, abiName, fmt.Sprintf(`["%v", "%v", %v]`, tokenSym, from.ID, 100000000000))
				tx1 := itest.NewTransaction([]*tx.Action{act1})
				trx, err = from.Sign(tx1)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}
			case tIndex <= 1000 || len(tokenMap[tokenList[0]].balance) < 10:
				abiName = issueToken721
				tokenSym := tokenList[rand.Intn(len(tokenList))]
				if len(tokenMap[tokenList[0]].balance) < 10 {
					tokenSym = tokenList[0]
				}
				issuer := accountMap[tokenMap[tokenSym].issuer]
				to := accounts[rand.Intn(len(accounts))]
				act0 := tx.NewAction("ram.iost", "buy", fmt.Sprintf(`["%v", "%v", %v]`, "admin", issuer.ID, 1000))
				act1 := tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, "admin", issuer.ID, 10))
				tx0 := itest.NewTransaction([]*tx.Action{act0, act1})
				trx, err := it.GetDefaultAccount().Sign(tx0)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}
				acts := []*tx.Action{}
				for i := 0; i < issueNumber; i++ {
					acts = append(acts, tx.NewAction(contractName, abiName, fmt.Sprintf(`["%v", "%v", "%v"]`, tokenSym, to.ID, "meta"+to.ID)))
				}
				tx1 := itest.NewTransaction(acts)
				trx, err = issuer.Sign(tx1)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}
			case tIndex <= 2000:
				abiName = transferToken721
				tokenSym := tokenList[rand.Intn(len(tokenList))]
				if len(tokenMap[tokenSym].balance) == 0 {
					tokenSym = tokenList[0]
				}
				var from *itest.Account
				var tokenID string
				for k, v := range tokenMap[tokenSym].balance {
					from = accountMap[k]
					tokenID = v[0]
					break
				}
				to := accounts[rand.Intn(len(accounts))]
				act0 := tx.NewAction("ram.iost", "buy", fmt.Sprintf(`["%v", "%v", %v]`, "admin", from.ID, 1000))
				act1 := tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, "admin", from.ID, 10))
				tx0 := itest.NewTransaction([]*tx.Action{act0, act1})
				trx, err := it.GetDefaultAccount().Sign(tx0)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}

				act1 = tx.NewAction(contractName, abiName, fmt.Sprintf(`["%v", "%v", "%v", "%v"]`, tokenSym, from.ID, to.ID, tokenID))
				tx1 := itest.NewTransaction([]*tx.Action{act1})
				trx, err = from.Sign(tx1)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
					tokenMap[tokenSym].balance[from.ID] = tokenMap[tokenSym].balance[from.ID][1:]
					if len(tokenMap[tokenSym].balance[from.ID]) == 0 {
						delete(tokenMap[tokenSym].balance, from.ID)
					}
				}
			case tIndex <= 2100:
				abiName = balanceOfToken721
				tokenSym := tokenList[rand.Intn(len(tokenList))]
				if len(tokenMap[tokenSym].balance) == 0 {
					tokenSym = tokenList[0]
				}
				from := accountMap[tokenMap[tokenSym].acclist[rand.Intn(len(tokenMap[tokenSym].acclist))]]
				act1 := tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, "admin", from.ID, 10))
				tx0 := itest.NewTransaction([]*tx.Action{act1})
				trx, err := it.GetDefaultAccount().Sign(tx0)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}

				act1 = tx.NewAction(contractName, abiName, fmt.Sprintf(`["%v", "%v"]`, tokenSym, from.ID))
				tx1 := itest.NewTransaction([]*tx.Action{act1})
				trx, err = from.Sign(tx1)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}
			case tIndex <= 2200:
				abiName = ownerOfToken721
				tokenSym := tokenList[rand.Intn(len(tokenList))]
				if tokenMap[tokenSym].supply == 0 {
					tokenSym = tokenList[0]
				}
				from := accounts[rand.Intn(len(accounts))]
				tokenID := rand.Intn(tokenMap[tokenSym].supply)
				act1 := tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, "admin", from.ID, 10))
				tx0 := itest.NewTransaction([]*tx.Action{act1})
				trx, err := it.GetDefaultAccount().Sign(tx0)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}

				act1 = tx.NewAction(contractName, abiName, fmt.Sprintf(`["%v", "%v"]`, tokenSym, tokenID))
				tx1 := itest.NewTransaction([]*tx.Action{act1})
				trx, err = from.Sign(tx1)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}
			case tIndex <= 2300:
				abiName = tokenOfOwnerToken721
				tokenSym := tokenList[rand.Intn(len(tokenList))]
				if len(tokenMap[tokenSym].balance) == 0 {
					tokenSym = tokenList[0]
				}
				var from *itest.Account
				var idx int
				for k, v := range tokenMap[tokenSym].balance {
					from = accountMap[k]
					idx = rand.Intn(len(v))
					break
				}
				act1 := tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, "admin", from.ID, 10))
				tx0 := itest.NewTransaction([]*tx.Action{act1})
				trx, err := it.GetDefaultAccount().Sign(tx0)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}

				act1 = tx.NewAction(contractName, abiName, fmt.Sprintf(`["%v", "%v", %v]`, tokenSym, from.ID, idx))
				tx1 := itest.NewTransaction([]*tx.Action{act1})
				trx, err = from.Sign(tx1)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}
			case tIndex <= 2400:
				abiName = tokenMetadataToken721
				tokenSym := tokenList[rand.Intn(len(tokenList))]
				if tokenMap[tokenSym].supply == 0 {
					tokenSym = tokenList[0]
				}
				from := accounts[rand.Intn(len(accounts))]
				tokenID := rand.Intn(tokenMap[tokenSym].supply)
				act1 := tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, "admin", from.ID, 10))
				tx0 := itest.NewTransaction([]*tx.Action{act1})
				trx, err := it.GetDefaultAccount().Sign(tx0)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}

				act1 = tx.NewAction(contractName, abiName, fmt.Sprintf(`["%v", "%v"]`, tokenSym, tokenID))
				tx1 := itest.NewTransaction([]*tx.Action{act1})
				trx, err = from.Sign(tx1)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}
			}
		}
		tokenMutex.Unlock()
		hashList, tmpList := it.SendTransactionN(trxs, false)
		errList = append(errList, tmpList...)
		ilog.Warnf("Send %v trxs, got %v hash, %v err", len(trxs), len(hashList), len(errList))

		if check {
			expire := time.Now().Add(itest.Timeout)
			for _, hash := range hashList {
				select {
				case hashCh <- &hashItem{hash: hash, expire: expire}:
				case <-time.After(1 * time.Millisecond):
				}
			}
		}

		select {
		case <-sig:
			return fmt.Errorf("signal %v", sig)
		case <-ticker.C:
		}

		counter++
		slotTotal += len(trxs)
		if counter == 10 {
			total += slotTotal
			currentTps := float64(slotTotal) / time.Since(slotStartTime).Seconds()
			averageTps := float64(total) / time.Since(startTime).Seconds()
			ilog.Warnf("Current tps %v, Average tps %v, Total tx %v, token num %v", currentTps, averageTps, total, len(tokenList))
			counter = 0
			slotTotal = 0
			slotStartTime = time.Now()
		}
	}
}
