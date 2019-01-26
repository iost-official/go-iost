package run

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"fmt"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/itest"
	"github.com/urfave/cli"
	"math/rand"
	"sync"
)

// BenchmarkSystemCommand is the subcommand for benchmark system.iost.
var BenchmarkSystemCommand = cli.Command{
	Name:      "benchmarkSystem",
	ShortName: "benchS",
	Usage:     "Run system benchmark by given tps",
	Flags:     BenchmarkSystemFlags,
	Action:    BenchmarkSystemAction,
}

// BenchmarkSystemFlags is the list of flags for benchmark.
var BenchmarkSystemFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "tps",
		Value: 50,
		Usage: "The expected ratio of transactions per second",
	},
}

const (
	setCode       = "setCode"
	updateCode    = "updateCode"
	receipt       = "receipt"
	requireAuth   = "requireAuth"
	cancelDelaytx = "cancelDelaytx"
)

// BenchmarkSystemAction is the action of benchmark.
var BenchmarkSystemAction = func(c *cli.Context) error {
	itest.Interval = 2 * time.Millisecond
	itest.InitAmount = "1000"
	itest.InitPledge = "1000"
	itest.InitRAM = "3000"
	logger := ilog.New()
	fileWriter := ilog.NewFileWriter(c.GlobalString("log"))
	fileWriter.SetLevel(ilog.LevelInfo)
	logger.AddWriter(fileWriter)
	ilog.InitLogger(logger)
	//ilog.SetLevel(ilog.LevelDebug)
	it, err := itest.Load(c.GlobalString("keys"), c.GlobalString("config"))
	if err != nil {
		return err
	}
	accountFile := c.GlobalString("account")
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

	code := `
				class Test {
					init() {}
					hello() { return "world"; }
					can_update(data) { return blockchain.requireAuth(blockchain.contractOwner(), "active"); }
				};
				module.exports = Test;
				`
	ABI := `
				{
					"lang": "javascript",
					"version": "1.0.0",
					"abi": [
						{"name": "hello", "args": [], "amountLimit": [] },
						{"name": "can_update", "args": ["string"], "amountLimit": [] }
					]
				}
				`
	contract, err := itest.NewContract(code, ABI)
	if err != nil {
		return err
	}

	checkReceiptConcurrent := 64
	var contractMutex sync.Mutex
	contractList := make([]string, 0)
	contractMap := make(map[string]string)
	delayTxList := make([]string, 0)

	hashCh := make(chan *hashItem, 4*tps*int(itest.Timeout.Seconds()))
	for c := 0; c < checkReceiptConcurrent; c++ {
		go func(hashCh chan *hashItem) {
			counter := 0
			failedCounter := 0
			for item := range hashCh {
				client := it.GetClients()[rand.Intn(len(it.GetClients()))]
				_, err := client.CheckTransactionWithTimeout(item.hash, item.expire)
				//ilog.Infof("receipt %v", r)
				counter++
				if err != nil {
					ilog.Errorf("check transaction failed, %v", err)
					failedCounter++
				}
				if counter%1000 == 0 {
					ilog.Warnf("check %v transaction, %v successful, %v failed.", counter, counter-failedCounter, failedCounter)
				}
				if len(hashCh) > 3*tps*int(itest.Timeout.Seconds()) {
					ilog.Infof("hash ch size too large %v", len(hashCh))
				}
			}
		}(hashCh)
	}

	contractName := "system.iost"
	for {
		trxs := make([]*itest.Transaction, 0)
		errList := []error{}
		contractMutex.Lock()
		for num := 0; num < tps; num++ {
			// receipt 1, requireAuth 1, setCode 1, updateCode 1, cancelDelayTx 1
			tIndex := rand.Intn(5)
			abiName := ""
			switch true {
			case tIndex <= 0 || len(contractList) == 0:
				abiName = setCode
				from := accounts[rand.Intn(len(accounts))]

				act0 := tx.NewAction("ram.iost", "buy", fmt.Sprintf(`["%v", "%v", %v]`, "admin", from.ID, 3000))
				act1 := tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, "admin", from.ID, 100))
				tx0 := itest.NewTransaction([]*tx.Action{act0, act1})
				trx, err := it.GetDefaultAccount().Sign(tx0)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}

				act1 = tx.NewAction(contractName, abiName, fmt.Sprintf(`["%v"]`, contract))
				tx1 := itest.NewTransaction([]*tx.Action{act1})
				trx, err = from.Sign(tx1)
				if err != nil {
					errList = append(errList, err)
				} else {
					hash, err := it.SendTransaction(trx, true)
					if err != nil {
						errList = append(errList, err)
					} else {
						contractID := fmt.Sprintf("Contract%v", hash)
						contractList = append(contractList, contractID)
						contractMap[contractID] = from.ID
					}
				}
				break
			case tIndex <= 1:
				abiName = updateCode
				contractID := contractList[rand.Intn(len(contractList))]
				owner := accountMap[contractMap[contractID]]
				act1 := tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, "admin", owner.ID, 100))
				tx0 := itest.NewTransaction([]*tx.Action{act1})
				trx, err := it.GetDefaultAccount().Sign(tx0)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}

				contract.ID = contractID
				act1 = tx.NewAction(contractName, abiName, fmt.Sprintf(`["%v", ""]`, contract))
				tx1 := itest.NewTransaction([]*tx.Action{act1})
				trx, err = owner.Sign(tx1)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}
				break
			case tIndex <= 2:
				abiName = cancelDelaytx
				break
				if len(delayTxList) == 0 {
					from := accounts[rand.Intn(len(accounts))]
					act1 := tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, "admin", from.ID, 10))
					tx0 := itest.NewTransaction([]*tx.Action{act1})
					tx0.Delay = 90 * 1e9
					trx, err := it.GetDefaultAccount().Sign(tx0)
					if err != nil {
						errList = append(errList, err)
					} else {
						hash, err := it.SendTransaction(trx, false)
						if err != nil {
							errList = append(errList, err)
						} else {
							delayTxList = append(delayTxList, hash)
						}
					}
				} else {
					ilog.Infof("cancel delay tx")
					act1 := tx.NewAction(contractName, abiName, fmt.Sprintf(`["%v"]`, delayTxList[0]))
					delayTxList = delayTxList[1:]
					tx1 := itest.NewTransaction([]*tx.Action{act1})
					trx, err := it.GetDefaultAccount().Sign(tx1)
					if err != nil {
						errList = append(errList, err)
					} else {
						trxs = append(trxs, trx)
					}
				}
				break
			case tIndex <= 3:
				abiName = receipt
				from := accounts[rand.Intn(len(accounts))]
				act1 := tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, "admin", from.ID, 10))
				tx0 := itest.NewTransaction([]*tx.Action{act1})
				trx, err := it.GetDefaultAccount().Sign(tx0)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}

				act1 = tx.NewAction(contractName, abiName, fmt.Sprintf(`["%v"]`, from.ID))
				tx1 := itest.NewTransaction([]*tx.Action{act1})
				trx, err = from.Sign(tx1)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}
				break
			case tIndex <= 4:
				abiName = requireAuth
				from := accounts[rand.Intn(len(accounts))]
				act1 := tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, "admin", from.ID, 10))
				tx0 := itest.NewTransaction([]*tx.Action{act1})
				trx, err := it.GetDefaultAccount().Sign(tx0)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}

				act1 = tx.NewAction(contractName, abiName, fmt.Sprintf(`["%v", "owner"]`, from.ID))
				tx1 := itest.NewTransaction([]*tx.Action{act1})
				trx, err = from.Sign(tx1)
				if err != nil {
					errList = append(errList, err)
				} else {
					trxs = append(trxs, trx)
				}
				break
			}
		}
		contractMutex.Unlock()
		hashList, tmpList := it.SendTransactionN(trxs, false)
		errList = append(errList, tmpList...)
		ilog.Warnf("Send %v trxs, got %v hash, %v err", len(trxs), len(hashList), len(errList))

		expire := time.Now().Add(itest.Timeout)
		for _, hash := range hashList {
			select {
			case hashCh <- &hashItem{hash: hash, expire: expire}:
			case <-time.After(1 * time.Millisecond):
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
			currentTps := float64(slotTotal) / time.Now().Sub(slotStartTime).Seconds()
			averageTps := float64(total) / time.Now().Sub(startTime).Seconds()
			ilog.Warnf("Current tps %v, Average tps %v, Total tx %v, contractNum %v, delaytxNum %v", currentTps, averageTps, total, len(contractList), len(delayTxList))
			counter = 0
			slotTotal = 0
			slotStartTime = time.Now()
		}
	}
}
