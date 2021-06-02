package run

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/bitly/go-simplejson"
	"github.com/urfave/cli"

	"github.com/iost-official/go-iost/v3/ilog"
	"github.com/iost-official/go-iost/v3/itest"
)

// BonusCaseCommand is the command of account test case
var BonusCaseCommand = cli.Command{
	Name:      "bonus_case",
	ShortName: "b_case",
	Usage:     "run bonus test case",
	Action:    BonusCaseAction,
}

func countBlockProducedBy(it *itest.ITest, acc string, number int64) (cnt int64, err error) {
	data, _, _, err := it.GetContractStorage("vote_producer.iost", "producerTable", acc)
	if data == "" || err != nil {
		return
	}
	obj, err := simplejson.NewJson([]byte(data))
	if err != nil {
		return
	}
	pubkey, err := obj.Get("pubkey").String()
	if err != nil {
		return
	}
	ilog.Debugf("%v producerInfo = %v, pubkey = %v", acc, data, pubkey)

	checkBlockConcurrent := 64

	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	numberCh := make(chan int64, 4*checkBlockConcurrent)
	for c := 0; c < checkBlockConcurrent; c++ {
		wg.Add(1)
		go func(numberCh chan int64) {
			for num := range numberCh {
				block, err := it.GetBlockByNumber(num)
				if err != nil {
					ilog.Errorf("get block error %v: %v", num, err)
				}
				if block.Witness == pubkey {
					if num == 0 {
						ilog.Errorf("%+v %v", block, pubkey)
					}
					mu.Lock()
					cnt++
					mu.Unlock()
				}
				if cnt > 0 && cnt%1000 == 0 {
					ilog.Infof("current cnt = %v", cnt)
				}
			}
			wg.Done()
		}(numberCh)
	}
	for i := int64(0); i < number; i++ {
		numberCh <- i
	}
	close(numberCh)
	wg.Wait()
	return
}

// BonusCaseAction is the action of Bonus test case
var BonusCaseAction = func(c *cli.Context) error {
	acc := c.GlobalString("aname")
	keysfile := c.GlobalString("keys")
	configfile := c.GlobalString("config")

	it, err := itest.Load(keysfile, configfile)
	if err != nil {
		return err
	}

	data, _, number, err := it.GetContractStorage("token.iost", "TB"+acc, "contribute")
	if err != nil {
		return err
	}
	ilog.Debugf("%v contribute = %v, number = %v", acc, data, number)

	contribute, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return err
	}
	data1, _, _, err := it.GetContractStorage("bonus.iost", "blockContrib", "")
	if err != nil {
		return err
	}
	ratef, err := strconv.ParseFloat(strings.Trim(data1, "\""), 64)
	if err != nil {
		return err
	}

	cnt, err := countBlockProducedBy(it, acc, number)
	if cnt == 0 || err != nil {
		return err
	}

	rate := int64(ratef * 1e8)
	if contribute == cnt*rate {
		ilog.Infof("success: contribute = %v, cnt*rate = %v", contribute, cnt*rate)
		return nil
	} else if contribute < cnt*rate && contribute*103 > cnt*rate*100 {
		ilog.Infof("success contribute is nearly equal to cnt*rate: contribute = %v, cnt*rate = %v", contribute, cnt*rate)
		return nil
	}
	return fmt.Errorf("check contribute failed: contribute = %v, cnt = %v, rate = %v(%v)", contribute, cnt, rate, data1)
}
