package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/iost-official/go-iost/ilog"

	"github.com/iost-official/go-iost/core/tx"
	. "github.com/iost-official/go-iost/verifier"
	"github.com/iost-official/go-iost/vm/database"
	. "github.com/smartystreets/goconvey/convey"
)

func prepareToken(t *testing.T, s *Simulator, pubAcc *TestAccount) {
	r, err := s.Call("token.iost", "create", fmt.Sprintf(`["%v", "%v", %v, {}]`, "iost", acc0.ID, "21000000000"), pubAcc.ID, pubAcc.KeyPair)
	if err != nil || r.Status.Code != tx.Success {
		t.Fatal(err, r)
	}
	for _, acc := range testAccounts {
		s.Call("token.iost", "issue", fmt.Sprintf(`["%v", "%v", "%v"]`, "iost", acc.ID, "2000000000"), pubAcc.ID, pubAcc.KeyPair)
	}
	s.Visitor.Commit()
}

func prepareVote(t *testing.T, s *Simulator, acc *TestAccount) (*tx.TxReceipt, error) {
	// deploy vote.iost
	setNonNativeContract(s, "vote.iost", "vote_common.js", ContractPath)
	s.Call("vote.iost", "init", `[]`, acc.ID, acc.KeyPair)

	// deploy voteresult
	err := setNonNativeContract(s, "Contractvoteresult", "voteresult.js", "./test_data/")
	if err != nil {
		t.Fatal(err)
	}
	s.Visitor.MPut("system.iost-contract_owner", "Contractvoteresult", `s`+acc.ID)
	s.SetGas(acc.ID, 1e8)
	s.SetRAM(acc.ID, 1e8)

	config := make(map[string]interface{})
	config["resultNumber"] = 2
	config["minVote"] = 10
	config["options"] = []string{"option1", "option2", "option3", "option4"}
	config["anyOption"] = false
	config["freezeTime"] = 0
	params := []interface{}{
		acc0.ID,
		"test vote",
		config,
	}
	b, _ := json.Marshal(params)
	r, err := s.Call("vote.iost", "NewVote", string(b), acc.ID, acc.KeyPair)
	s.Visitor.Commit()
	return r, err
}

func Test_NewVote(t *testing.T) {
	ilog.Stop()
	Convey("test NewVote", t, func() {
		s := NewSimulator()
		defer s.Clear()

		createAccountsWithResource(s)
		prepareToken(t, s, acc0)

		Convey("test NewVote", func() {
			r, err := prepareVote(t, s, acc0)

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(1999999000*1e8))
			So(database.MustUnmarshal(s.Visitor.Get("vote.iost-current_id")), ShouldEqual, `"1"`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1")), ShouldEqual, `{"deleted":0,"description":"test vote","resultNumber":2,"minVote":10,"anyOption":false,"freezeTime":0,"deposit":"1000","optionId":4,"optionNum":4,"options":{"option1":{"id":"0","deleted":0,"clearTime":-1},"option2":{"id":"1","deleted":0,"clearTime":-1},"option3":{"id":"2","deleted":0,"clearTime":-1},"option4":{"id":"3","deleted":0,"clearTime":-1}},"preResult":{"0":0,"1":0,"2":0,"3":0}}`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "0")), ShouldEqual, `"0"`)

			r, err = s.Call("Contractvoteresult", "GetResult", `["1"]`, acc0.ID, acc0.KeyPair)

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(database.MustUnmarshal(s.Visitor.MGet("Contractvoteresult-vote_result", "1")), ShouldEqual, `[]`)
		})

	})
}

func Test_AddOption(t *testing.T) {
	ilog.Stop()
	Convey("test vote", t, func() {
		s := NewSimulator()
		defer s.Clear()

		createAccountsWithResource(s)
		prepareToken(t, s, acc0)
		prepareVote(t, s, acc0)

		Convey("test AddOption", func() {
			r, err := s.Call("vote.iost", "AddOption", `["1", "option5", true]`, acc0.ID, acc0.KeyPair)

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"0", "1", "2", "3", "4"})
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "4")), ShouldEqual, `"0"`)

			r, err = s.Call("vote.iost", "GetOption", `["1", "option5"]`, acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(r.Returns[0], ShouldEqual, `["{\"votes\":\"0\",\"deleted\":0,\"clearTime\":-1}"]`)
		})
	})
}

func Test_RemoveOption(t *testing.T) {
	ilog.Stop()
	Convey("test vote", t, func() {
		s := NewSimulator()
		defer s.Clear()

		createAccountsWithResource(s)
		prepareToken(t, s, acc0)
		prepareVote(t, s, acc0)

		Convey("test RemoveOption", func() {
			r, err := s.Call("vote.iost", "RemoveOption", `["1", "option2", true]`, acc0.ID, acc0.KeyPair)

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"0", "1", "2", "3"})
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "1")), ShouldEqual, `"0"`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1")), ShouldEqual, `{"deleted":0,"description":"test vote","resultNumber":2,"minVote":10,"anyOption":false,"freezeTime":0,"deposit":"1000","optionId":4,"optionNum":4,"options":{"option1":{"id":"0","deleted":0,"clearTime":-1},"option2":{"id":"1","deleted":1,"clearTime":-1},"option3":{"id":"2","deleted":0,"clearTime":-1},"option4":{"id":"3","deleted":0,"clearTime":-1}},"preResult":{"0":0,"1":0,"2":0,"3":0}}`)
		})
	})
}

func Test_Vote(t *testing.T) {
	ilog.Stop()
	Convey("test vote", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareToken(t, s, acc0)
		prepareVote(t, s, acc0)

		rs, err := s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "5"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(rs.Status.Message, ShouldEqual, "")

		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "2")), ShouldEqual, `"5"`)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc1.ID)), ShouldEqual, `{"2":["5",0,"0"]}`)

		rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "5"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(rs.Status.Message, ShouldEqual, "")
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "2")), ShouldEqual, `"10"`)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc1.ID)), ShouldEqual, `{"2":["10",0,"0"]}`)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1")), ShouldContainSubstring, `"2":1`)

		rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "20"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(rs.Status.Message, ShouldEqual, "")
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "0")), ShouldEqual, `"20"`)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc1.ID)), ShouldEqual, `{"0":["20",0,"0"],"2":["10",0,"0"]}`)
		info := database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
		So(info, ShouldContainSubstring, `"2":1`)
		So(info, ShouldContainSubstring, `"0":1`)

		rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "100"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(rs.Status.Message, ShouldEqual, "")
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "2")), ShouldEqual, `"110"`)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc0.ID)), ShouldEqual, `{"2":["100",0,"0"]}`)
		info = database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
		So(info, ShouldContainSubstring, `"2":1`)
		So(info, ShouldContainSubstring, `"0":1`)

		rs, err = s.Call("Contractvoteresult", "GetResult", `["1"]`, acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(rs.Status.Message, ShouldEqual, "")
		So(database.MustUnmarshal(s.Visitor.MGet("Contractvoteresult-vote_result", "1")), ShouldEqual, `[{"option":"option3","votes":"110"},{"option":"option1","votes":"20"}]`)

		r, err := s.Call("vote.iost", "GetOption", `["1", "option3"]`, acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)
		So(r.Returns[0], ShouldEqual, `["{\"votes\":\"110\",\"deleted\":0,\"clearTime\":-1}"]`)

		r, err = s.Call("vote.iost", "GetVote", fmt.Sprintf(`["1", "%v"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)
		So(r.Returns[0], ShouldEqual, `["[{\"option\":\"option1\",\"votes\":\"20\",\"voteTime\":0,\"clearedVotes\":\"0\"},{\"option\":\"option3\",\"votes\":\"10\",\"voteTime\":0,\"clearedVotes\":\"0\"}]"]`)
	})
}

func Test_Unvote(t *testing.T) {
	ilog.Stop()
	Convey("test Unvote", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareToken(t, s, acc0)
		prepareVote(t, s, acc0)

		// vote
		s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		info := database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
		So(info, ShouldContainSubstring, `"0":1`)

		s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		info = database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
		So(info, ShouldContainSubstring, `"0":1`)
		So(info, ShouldContainSubstring, `"1":1`)

		s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "2")), ShouldEqual, `"300"`)
		info = database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
		So(info, ShouldContainSubstring, `"0":1`)
		So(info, ShouldContainSubstring, `"1":1`)
		So(info, ShouldContainSubstring, `"2":1`)

		s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, acc2.ID), acc2.ID, acc2.KeyPair)
		info = database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
		So(info, ShouldContainSubstring, `"0":1,"1":1,"2":1,"3":1`)

		// get result
		s.Call("Contractvoteresult", "GetResult", `["1"]`, acc0.ID, acc0.KeyPair)
		So(database.MustUnmarshal(s.Visitor.MGet("Contractvoteresult-vote_result", "1")), ShouldEqual, `[{"option":"option4","votes":"400"},{"option":"option3","votes":"300"}]`)

		// unvote
		s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option3", "100"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "2")), ShouldEqual, `"200"`)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc0.ID)), ShouldEqual, `{"1":["100",0,"0"],"2":["200",0,"0"]}`)
		info = database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
		So(info, ShouldContainSubstring, `"0":1,"1":1,"2":1,"3":1`)

		// get result
		s.Call("Contractvoteresult", "GetResult", `["1"]`, acc0.ID, acc0.KeyPair)
		So(database.MustUnmarshal(s.Visitor.MGet("Contractvoteresult-vote_result", "1")), ShouldEqual, `[{"option":"option4","votes":"400"},{"option":"option2","votes":"200"}]`)

		// unvote again
		s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option2", "95"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "1")), ShouldEqual, `"105"`)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc0.ID)), ShouldEqual, `{"1":["5",0,"0"],"2":["200",0,"0"]}`)
		info = database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
		So(info, ShouldContainSubstring, `"0":1,"1":1,"2":1,"3":1`)

		s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "1")), ShouldEqual, `"5"`)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc1.ID)), ShouldEqual, `{"0":["100",0,"0"]}`)
		info = database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
		So(info, ShouldContainSubstring, `"0":1,"1":0,"2":1,"3":1`)

		// get result
		s.Call("Contractvoteresult", "GetResult", `["1"]`, acc0.ID, acc0.KeyPair)
		So(database.MustUnmarshal(s.Visitor.MGet("Contractvoteresult-vote_result", "1")), ShouldEqual, `[{"option":"option4","votes":"400"},{"option":"option3","votes":"200"}]`)
	})
}

func Test_DelVote(t *testing.T) {
	ilog.Stop()
	Convey("test DelVote", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareToken(t, s, acc0)
		prepareVote(t, s, acc0)

		// vote
		s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, acc2.ID), acc2.ID, acc2.KeyPair)

		// del vote
		s.Call("vote.iost", "DelVote", `["1"]`, acc0.ID, acc0.KeyPair)
		So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{})

		// unvote part
		s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option2", "95"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{})
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc0.ID)), ShouldEqual, `{"1":["5",0,"0"],"2":["300",0,"0"]}`)

		// unvote all
		s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option2", "5"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(s.Visitor.MHas("vote.iost-u_1", acc0.ID), ShouldEqual, false)
	})
}

func Test_MixVoteOption(t *testing.T) {
	ilog.Stop()
	Convey("test mixed", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareToken(t, s, acc0)
		prepareVote(t, s, acc0)

		Convey("test AddOption not clear", func() {
			// vote
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, acc1.ID), acc1.ID, acc1.KeyPair)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, acc1.ID), acc1.ID, acc1.KeyPair)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, acc0.ID), acc0.ID, acc0.KeyPair)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, acc0.ID), acc0.ID, acc0.KeyPair)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, acc2.ID), acc2.ID, acc2.KeyPair)

			// add option
			s.Call("vote.iost", "AddOption", `["1", "option5", false]`, acc0.ID, acc0.KeyPair)
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"0", "1", "2", "3", "4"})

			// get result
			s.Call("Contractvoteresult", "GetResult", `["1"]`, acc0.ID, acc0.KeyPair)
			So(database.MustUnmarshal(s.Visitor.MGet("Contractvoteresult-vote_result", "1")), ShouldEqual, `[{"option":"option4","votes":"400"},{"option":"option3","votes":"300"}]`)

			s.Head.Number++
			// remove option
			s.Call("vote.iost", "RemoveOption", `["1", "option1", false]`, acc0.ID, acc0.KeyPair)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "0")), ShouldEqual, `"100"`)
			info := database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
			So(info, ShouldContainSubstring, `"option1":{"id":"0","deleted":1`)

			// add option
			s.Call("vote.iost", "AddOption", `["1", "option1", false]`, acc0.ID, acc0.KeyPair)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "0")), ShouldEqual, `"100"`)
			info = database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
			So(info, ShouldContainSubstring, `"0":1`)
		})

		Convey("test AddOption and clear", func() {
			// vote
			rs, err := s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, acc1.ID), acc1.ID, acc1.KeyPair)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, acc1.ID), acc1.ID, acc1.KeyPair)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, acc0.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "200"]`, acc0.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, acc2.ID), acc2.ID, acc2.KeyPair)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")

			s.Head.Number++
			// remove option
			rs, err = s.Call("vote.iost", "RemoveOption", `["1", "option1", false]`, acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			// add option
			rs, err = s.Call("vote.iost", "AddOption", `["1", "option1", true]`, acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")

			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"0", "1", "2", "3"})
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "0")), ShouldEqual, `"0"`)
			info := database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
			So(info, ShouldContainSubstring, `"0":0,"1":1,"2":1,"3":1`)

			// vote after clear in same block
			rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, acc1.ID), acc1.ID, acc1.KeyPair)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			info = database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
			So(info, ShouldContainSubstring, `"0":0`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc1.ID)), ShouldEqual, `{"0":["200",1,"100"],"1":["100",0,"0"]}`)

			// vote after the clear block
			s.Head.Number++
			rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, acc1.ID), acc1.ID, acc1.KeyPair)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			info = database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
			So(info, ShouldContainSubstring, `"0":1`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc1.ID)), ShouldEqual, `{"0":["300",2,"200"],"1":["100",0,"0"]}`)

			// vote again
			rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, acc1.ID), acc1.ID, acc1.KeyPair)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc1.ID)), ShouldEqual, `{"0":["400",2,"200"],"1":["100",0,"0"]}`)

			// get result
			rs, err = s.Call("Contractvoteresult", "GetResult", `["1"]`, acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("Contractvoteresult-vote_result", "1")), ShouldEqual, `[{"option":"option4","votes":"400"},{"option":"option1","votes":"200"}]`)

			// unvote
			rs, err = s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option1", "50"]`, acc1.ID), acc1.ID, acc1.KeyPair)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc1.ID)), ShouldEqual, `{"0":["350",2,"150"],"1":["100",0,"0"]}`)

			// unvote again
			rs, err = s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option1", "200"]`, acc1.ID), acc1.ID, acc1.KeyPair)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc1.ID)), ShouldEqual, `{"0":["150",2,"0"],"1":["100",0,"0"]}`)
			info = database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
			So(info, ShouldContainSubstring, `"0":1`)
		})

		Convey("test RemoveOption not force", func() {
			// vote
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, acc1.ID), acc1.ID, acc1.KeyPair)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "200"]`, acc1.ID), acc1.ID, acc1.KeyPair)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, acc0.ID), acc0.ID, acc0.KeyPair)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, acc0.ID), acc0.ID, acc0.KeyPair)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, acc2.ID), acc2.ID, acc2.KeyPair)

			// add option
			s.Call("vote.iost", "AddOption", `["1", "option5", false]`, acc0.ID, acc0.KeyPair)
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"0", "1", "2", "3", "4"})

			// remove option
			s.Call("vote.iost", "RemoveOption", `["1", "option5", false]`, acc0.ID, acc0.KeyPair)
			// s.Call("vote.iost", "RemoveOption", `["1", "option4", false]`,acc0.ID, acc0.KeyPair) // should fail
			// s.Call("vote.iost", "RemoveOption", `["1", "option3", false]`,acc0.ID, acc0.KeyPair) // should fail
			// s.Call("vote.iost", "RemoveOption", `["1", "option2", false]`,acc0.ID, acc0.KeyPair) // should fail
			s.Call("vote.iost", "RemoveOption", `["1", "option1", false]`, acc0.ID, acc0.KeyPair)
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"0", "1", "2", "3", "4"})
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "4")), ShouldEqual, `"0"`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "3")), ShouldEqual, `"400"`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "2")), ShouldEqual, `"300"`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "1")), ShouldEqual, `"300"`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "0")), ShouldEqual, `"100"`)
			info := database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
			So(info, ShouldContainSubstring, `"0":0,"1":1,"2":1,"3":1`)
		})

		Convey("test RemoveOption with force", func() {
			// vote
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, acc1.ID), acc1.ID, acc1.KeyPair)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "200"]`, acc1.ID), acc1.ID, acc1.KeyPair)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, acc0.ID), acc0.ID, acc0.KeyPair)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, acc0.ID), acc0.ID, acc0.KeyPair)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, acc2.ID), acc2.ID, acc2.KeyPair)

			// add option
			s.Call("vote.iost", "AddOption", `["1", "option5", false]`, acc0.ID, acc0.KeyPair)
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"0", "1", "2", "3", "4"})

			// remove option
			s.Call("vote.iost", "RemoveOption", `["1", "option5", true]`, acc0.ID, acc0.KeyPair)
			s.Call("vote.iost", "RemoveOption", `["1", "option4", true]`, acc0.ID, acc0.KeyPair)
			s.Call("vote.iost", "RemoveOption", `["1", "option3", true]`, acc0.ID, acc0.KeyPair)
			s.Call("vote.iost", "RemoveOption", `["1", "option2", true]`, acc0.ID, acc0.KeyPair)
			s.Call("vote.iost", "RemoveOption", `["1", "option1", true]`, acc0.ID, acc0.KeyPair)
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"0", "1", "2", "3", "4"})
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "4")), ShouldEqual, `"0"`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "3")), ShouldEqual, `"400"`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "2")), ShouldEqual, `"300"`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "1")), ShouldEqual, `"300"`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "0")), ShouldEqual, `"100"`)
			info := database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
			So(info, ShouldContainSubstring, `"0":0,"1":0,"2":0,"3":0,"4":0`)
		})
	})
}
