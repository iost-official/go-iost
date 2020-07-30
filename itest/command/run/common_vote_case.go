package run

import (
	"fmt"

	"github.com/bitly/go-simplejson"
	"github.com/iost-official/go-iost/itest"
	"github.com/urfave/cli"
)

// CommonVoteCaseCommand is the command of common vote test case
var CommonVoteCaseCommand = cli.Command{
	Name:      "common_vote_case",
	ShortName: "cv_case",
	Usage:     "run common VoteProducer test case",
	Flags:     CommonVoteCaseFlags,
	Action:    CommonVoteCaseAction,
}

// CommonVoteCaseFlags is the flags of vote test case
var CommonVoteCaseFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "account, a",
		Value: "accounts.json",
		Usage: "load accounts from `FILE`",
	},
}

// CommonVoteCaseAction is the action of vote test case
var CommonVoteCaseAction = func(c *cli.Context) error {
	afile := c.String("account")
	keysfile := c.GlobalString("keys")
	configfile := c.GlobalString("newVoteConfig")
	it, err := itest.Load(keysfile, configfile)
	if err != nil {
		return err
	}
	client := it.GetClients()[0]
	accounts, err := itest.LoadAccounts(afile)
	if err != nil {
		return err
	}

	newVoteConfig := make(map[string]interface{})
	newVoteConfig["resultNumber"] = 2
	newVoteConfig["minVote"] = 10
	newVoteConfig["options"] = []string{"option1", "option2", "option3", "option4"}
	newVoteConfig["anyOption"] = false
	newVoteConfig["freezeTime"] = 0
	bank := it.GetDefaultAccount()
	hash, err := it.CallActionWithRandClient(it.GetDefaultAccount(), "vote.iost", "newVote", bank.ID, "test vote", newVoteConfig)
	if err != nil {
		return err
	}
	receipt, err := client.GetReceipt(hash)
	if err != nil {
		return err
	}
	js, err := simplejson.NewJson([]byte(receipt.Returns[0]))
	if err != nil {
		return err
	}
	voteID, err := js.GetIndex(0).String()
	if err != nil {
		return err
	}
	fmt.Println("vote id is", voteID)
	allArgs := make([][]interface{}, 0)
	allArgs = append(allArgs, []interface{}{voteID, accounts[1].ID, "option3", "5"})
	allArgs = append(allArgs, []interface{}{voteID, accounts[1].ID, "option3", "5"})
	allArgs = append(allArgs, []interface{}{voteID, accounts[1].ID, "option1", "20"})
	allArgs = append(allArgs, []interface{}{voteID, accounts[0].ID, "option3", "100"})
	var callingAccounts = []*itest.Account{accounts[1], accounts[1], accounts[1], accounts[0]}

	res := make(chan interface{})
	go func() {
		for idx := range allArgs {
			go func(i int, res chan interface{}) {
				_, err := it.CallActionWithRandClient(callingAccounts[i], "vote.iost", "vote", allArgs[i]...)
				res <- err
			}(idx, res)
		}
	}()
	for range allArgs {
		switch value := (<-res).(type) {
		case error:
			return value.(error)
		}
	}

	checkReturn := func(actionName string, expected string, args ...interface{}) error {
		hash, err := client.CallAction(true, bank, "vote.iost", actionName, args...)
		if err != nil {
			return err
		}
		receipt, err = client.GetReceipt(hash)
		if err != nil {
			return err
		}
		js, err := simplejson.NewJson([]byte(receipt.Returns[0]))
		if err != nil {
			return err
		}
		result, err := js.GetIndex(0).String()
		if err != nil {
			return err
		}
		if result != expected {
			return fmt.Errorf("invalid return %v, expect %v", result, expected)
		}
		return nil
	}
	res2 := make(chan error)
	go func() {
		res2 <- checkReturn("getResult", `[{"option":"option3","votes":"110"},{"option":"option1","votes":"20"}]`, voteID)
	}()
	go func() {
		res2 <- checkReturn("getOption", `{"votes":"110","deleted":false,"clearTime":-1}`, voteID, "option3")
	}()
	//go func() {
	//	res2 <- checkReturn("getVote", `["[{\"option\":\"option3\",\"votes\":\"10\",\"voteTime\":0,\"clearedVotes\":\"0\"},{\"option\":\"option1\",\"votes\":\"20\",\"voteTime\":0,\"clearedVotes\":\"0\"}]"]`, voteID, accounts[1].ID)
	//} ()
	for i := 0; i < 2; i++ {
		if err := <-res2; err != nil {
			return err
		}
	}
	return nil
}
