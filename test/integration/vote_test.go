package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/iost-official/go-iost/ilog"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	. "github.com/iost-official/go-iost/verifier"
	"github.com/iost-official/go-iost/vm/database"
	. "github.com/smartystreets/goconvey/convey"
)

func prepareToken(t *testing.T, s *Simulator, kp *account.KeyPair) {
	r, err := s.Call("token.iost", "create", fmt.Sprintf(`["%v", "%v", %v, {}]`, "iost", testID[0], "21000000000"), kp.ID, kp)
	if err != nil || r.Status.Code != tx.Success {
		t.Fatal(err, r)
	}
	for i := 0; i < 18; i += 2 {
		s.Call("token.iost", "issue", fmt.Sprintf(`["%v", "%v", "%v"]`, "iost", testID[i], "2000000000"), kp.ID, kp)
	}
	s.Visitor.Commit()
}

func prepareVote(t *testing.T, s *Simulator, kp *account.KeyPair) (*tx.TxReceipt, error) {
	// deploy vote.iost
	setNonNativeContract(s, "vote.iost", "vote_common.js", ContractPath)
	s.Call("vote.iost", "init", `[]`, kp.ID, kp)

	// deploy voteresult
	err := setNonNativeContract(s, "Contractvoteresult", "voteresult.js", "./test_data/")
	if err != nil {
		t.Fatal(err)
	}
	s.Visitor.MPut("system.iost-contract_owner", "Contractvoteresult", `s`+kp.ID)
	s.SetGas(kp.ID, 1e8)
	s.SetRAM(kp.ID, 1e8)

	config := make(map[string]interface{})
	config["resultNumber"] = 2
	config["minVote"] = 10
	config["options"] = []string{"option1", "option2", "option3", "option4"}
	config["anyOption"] = false
	config["freezeTime"] = 0
	params := []interface{}{
		testID[0],
		"test vote",
		config,
	}
	b, _ := json.Marshal(params)
	r, err := s.Call("vote.iost", "NewVote", string(b), kp.ID, kp)
	s.Visitor.Commit()
	return r, err
}

func Test_NewVote(t *testing.T) {
	ilog.Stop()
	Convey("test NewVote", t, func() {
		s := NewSimulator()
		defer s.Clear()

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		createAccountsWithResource(s)
		prepareToken(t, s, kp)

		Convey("test NewVote", func() {
			r, err := prepareVote(t, s, kp)

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.Visitor.TokenBalance("iost", testID[0]), ShouldEqual, int64(1999999000*1e8))
			So(database.MustUnmarshal(s.Visitor.Get("vote.iost-current_id")), ShouldEqual, `"1"`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1")), ShouldEqual, `{"description":"test vote","resultNumber":2,"minVote":10,"anyOption":false,"freezeTime":0,"deposit":"1000"}`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option1")), ShouldEqual, `["0",false,-1]`)

			r, err = s.Call("Contractvoteresult", "GetResult", `["1"]`, kp.ID, kp)

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

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		createAccountsWithResource(s)
		prepareToken(t, s, kp)
		prepareVote(t, s, kp)

		Convey("test AddOption", func() {
			r, err := s.Call("vote.iost", "AddOption", `["1", "option5", true]`, kp.ID, kp)

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"option1", "option2", "option3", "option4", "option5"})
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option5")), ShouldEqual, `["0",false,-1]`)

			r, err = s.Call("vote.iost", "GetOption", `["1", "option5"]`, kp.ID, kp)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(r.Returns[0], ShouldEqual, `["{\"votes\":\"0\",\"deleted\":false,\"clearTime\":-1}"]`)
		})
	})
}

func Test_RemoveOption(t *testing.T) {
	ilog.Stop()
	Convey("test vote", t, func() {
		s := NewSimulator()
		defer s.Clear()

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		createAccountsWithResource(s)
		prepareToken(t, s, kp)
		prepareVote(t, s, kp)

		Convey("test RemoveOption", func() {
			r, err := s.Call("vote.iost", "RemoveOption", `["1", "option2", true]`, kp.ID, kp)

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"option1", "option2", "option3", "option4"})
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option2")), ShouldEqual, `["0",true,-1]`)
		})
	})
}

func Test_Vote(t *testing.T) {
	ilog.Stop()
	Convey("test vote", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0
		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		createAccountsWithResource(s)
		prepareToken(t, s, kp)
		prepareVote(t, s, kp)

		Convey("test Vote", func() {
			kp2, _ := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "5"]`, testID[2]), kp2.ID, kp2)

			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option3")), ShouldEqual, `["5",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", testID[2])), ShouldEqual, `{"option3":["5",0,"0"]}`)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{})

			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "5"]`, testID[2]), kp2.ID, kp2)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option3")), ShouldEqual, `["10",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", testID[2])), ShouldEqual, `{"option3":["10",0,"0"]}`)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{"option3"})

			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "20"]`, testID[2]), kp2.ID, kp2)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option1")), ShouldEqual, `["20",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", testID[2])), ShouldEqual, `{"option3":["10",0,"0"],"option1":["20",0,"0"]}`)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{"option3", "option1"})

			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "100"]`, testID[0]), kp.ID, kp)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option3")), ShouldEqual, `["110",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", testID[0])), ShouldEqual, `{"option3":["100",0,"0"]}`)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{"option3", "option1"})

			s.Call("Contractvoteresult", "GetResult", `["1"]`, kp.ID, kp)
			So(database.MustUnmarshal(s.Visitor.MGet("Contractvoteresult-vote_result", "1")), ShouldEqual, `[{"option":"option3","votes":"110"},{"option":"option1","votes":"20"}]`)

			r, err := s.Call("vote.iost", "GetOption", `["1", "option3"]`, kp.ID, kp)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(r.Returns[0], ShouldEqual, `["{\"votes\":\"110\",\"deleted\":false,\"clearTime\":-1}"]`)

			r, err = s.Call("vote.iost", "GetVote", fmt.Sprintf(`["1", "%v"]`, testID[2]), kp2.ID, kp2)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(r.Returns[0], ShouldEqual, `["[{\"option\":\"option3\",\"votes\":\"10\",\"voteTime\":0,\"clearedVotes\":\"0\"},{\"option\":\"option1\",\"votes\":\"20\",\"voteTime\":0,\"clearedVotes\":\"0\"}]"]`)
		})

		Convey("test Unvote", func() {
			kp2, _ := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			kp3, _ := account.NewKeyPair(common.Base58Decode(testID[5]), crypto.Secp256k1)
			// vote
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{"option1"})

			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[0]), kp.ID, kp)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{"option1", "option2"})

			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, testID[0]), kp.ID, kp)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{"option1", "option2", "option3"})

			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, testID[4]), kp3.ID, kp3)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{"option1", "option2", "option3", "option4"})

			// get result
			s.Call("Contractvoteresult", "GetResult", `["1"]`, kp.ID, kp)
			So(database.MustUnmarshal(s.Visitor.MGet("Contractvoteresult-vote_result", "1")), ShouldEqual, `[{"option":"option4","votes":"400"},{"option":"option3","votes":"300"}]`)

			// unvote
			s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option3", "100"]`, testID[0]), kp.ID, kp)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option3")), ShouldEqual, `["200",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", testID[0])), ShouldEqual, `{"option2":["100",0,"0"],"option3":["200",0,"0"]}`)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{"option1", "option2", "option3", "option4"})

			// get result
			s.Call("Contractvoteresult", "GetResult", `["1"]`, kp.ID, kp)
			So(database.MustUnmarshal(s.Visitor.MGet("Contractvoteresult-vote_result", "1")), ShouldEqual, `[{"option":"option4","votes":"400"},{"option":"option2","votes":"200"}]`)

			// unvote again
			s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option2", "95"]`, testID[0]), kp.ID, kp)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option2")), ShouldEqual, `["105",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", testID[0])), ShouldEqual, `{"option2":["5",0,"0"],"option3":["200",0,"0"]}`)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{"option1", "option2", "option3", "option4"})

			s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[2]), kp2.ID, kp2)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option2")), ShouldEqual, `["5",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", testID[2])), ShouldEqual, `{"option1":["100",0,"0"]}`)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{"option1", "option3", "option4"})

			// get result
			s.Call("Contractvoteresult", "GetResult", `["1"]`, kp.ID, kp)
			So(database.MustUnmarshal(s.Visitor.MGet("Contractvoteresult-vote_result", "1")), ShouldEqual, `[{"option":"option4","votes":"400"},{"option":"option3","votes":"200"}]`)
		})
	})
}

func Test_DelVote(t *testing.T) {
	ilog.Stop()
	Convey("test DelVote", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0
		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		createAccountsWithResource(s)
		prepareToken(t, s, kp)
		prepareVote(t, s, kp)

		Convey("test delete vote", func() {
			kp2, _ := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			kp3, _ := account.NewKeyPair(common.Base58Decode(testID[5]), crypto.Secp256k1)
			// vote
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[0]), kp.ID, kp)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, testID[0]), kp.ID, kp)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, testID[4]), kp3.ID, kp3)

			// del vote
			s.Call("vote.iost", "DelVote", `["1"]`, kp.ID, kp)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{})
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{})

			// unvote part
			s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option2", "95"]`, testID[0]), kp.ID, kp)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{})
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{})
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", testID[0])), ShouldEqual, `{"option2":["5",0,"0"],"option3":["300",0,"0"]}`)

			// unvote all
			s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option2", "5"]`, testID[0]), kp.ID, kp)
			s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, testID[0]), kp.ID, kp)
			So(s.Visitor.MHas("vote.iost-u_1", testID[0]), ShouldEqual, false)
		})
	})
}

func Test_MixVoteOption(t *testing.T) {
	ilog.Stop()
	Convey("test mixed", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0
		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		createAccountsWithResource(s)
		prepareToken(t, s, kp)
		prepareVote(t, s, kp)

		Convey("test AddOption not clear", func() {
			kp2, _ := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			kp3, _ := account.NewKeyPair(common.Base58Decode(testID[5]), crypto.Secp256k1)
			// vote
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[0]), kp.ID, kp)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, testID[0]), kp.ID, kp)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, testID[4]), kp3.ID, kp3)

			// add option
			s.Call("vote.iost", "AddOption", `["1", "option5", false]`, kp.ID, kp)
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"option1", "option2", "option3", "option4", "option5"})

			// get result
			s.Call("Contractvoteresult", "GetResult", `["1"]`, kp.ID, kp)
			So(database.MustUnmarshal(s.Visitor.MGet("Contractvoteresult-vote_result", "1")), ShouldEqual, `[{"option":"option4","votes":"400"},{"option":"option3","votes":"300"}]`)

			s.Head.Number++
			// remove option
			s.Call("vote.iost", "RemoveOption", `["1", "option1", false]`, kp.ID, kp)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option1")), ShouldEqual, `["100",true,-1]`)

			// add option
			s.Call("vote.iost", "AddOption", `["1", "option1", false]`, kp.ID, kp)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option1")), ShouldEqual, `["100",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-p_1", "option1")), ShouldEqual, `"100"`)
		})

		Convey("test AddOption and clear", func() {
			kp2, _ := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			kp3, _ := account.NewKeyPair(common.Base58Decode(testID[5]), crypto.Secp256k1)
			// vote
			rs, err := s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[2]), kp2.ID, kp2)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[0]), kp.ID, kp)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "200"]`, testID[0]), kp.ID, kp)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, testID[4]), kp3.ID, kp3)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")

			s.Head.Number++
			// remove option
			rs, err = s.Call("vote.iost", "RemoveOption", `["1", "option1", false]`, kp.ID, kp)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			// add option
			rs, err = s.Call("vote.iost", "AddOption", `["1", "option1", true]`, kp.ID, kp)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")

			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"option1", "option2", "option3", "option4"})
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option1")), ShouldEqual, `["0",false,1]`)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{"option2", "option3", "option4"})

			// vote after clear in same block
			rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			So(s.Visitor.MHas("vote.iost-p_1", "option1"), ShouldEqual, false)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", testID[2])), ShouldEqual, `{"option1":["200",1,"100"],"option2":["100",0,"0"]}`)

			// vote after the clear block
			s.Head.Number++
			rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-p_1", "option1")), ShouldEqual, `"100"`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", testID[2])), ShouldEqual, `{"option1":["300",2,"200"],"option2":["100",0,"0"]}`)

			// vote again
			rs, err = s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", testID[2])), ShouldEqual, `{"option1":["400",2,"200"],"option2":["100",0,"0"]}`)

			// get result
			rs, err = s.Call("Contractvoteresult", "GetResult", `["1"]`, kp.ID, kp)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("Contractvoteresult-vote_result", "1")), ShouldEqual, `[{"option":"option4","votes":"400"},{"option":"option2","votes":"200"}]`)

			// unvote
			rs, err = s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option1", "50"]`, testID[2]), kp2.ID, kp2)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", testID[2])), ShouldEqual, `{"option1":["350",2,"150"],"option2":["100",0,"0"]}`)

			// unvote again
			rs, err = s.Call("vote.iost", "Unvote", fmt.Sprintf(`["1", "%v", "option1", "200"]`, testID[2]), kp2.ID, kp2)
			So(err, ShouldBeNil)
			So(rs.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", testID[2])), ShouldEqual, `{"option1":["150",2,"0"],"option2":["100",0,"0"]}`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-p_1", "option1")), ShouldEqual, `"150"`)
		})

		Convey("test RemoveOption not force", func() {
			kp2, _ := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			kp3, _ := account.NewKeyPair(common.Base58Decode(testID[5]), crypto.Secp256k1)
			// vote
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "200"]`, testID[2]), kp2.ID, kp2)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[0]), kp.ID, kp)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, testID[0]), kp.ID, kp)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, testID[4]), kp3.ID, kp3)

			// add option
			s.Call("vote.iost", "AddOption", `["1", "option5", false]`, kp.ID, kp)
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"option1", "option2", "option3", "option4", "option5"})

			// remove option
			s.Call("vote.iost", "RemoveOption", `["1", "option5", false]`, kp.ID, kp)
			// s.Call("vote.iost", "RemoveOption", `["1", "option4", false]`, kp.ID, kp) // should fail
			// s.Call("vote.iost", "RemoveOption", `["1", "option3", false]`, kp.ID, kp) // should fail
			// s.Call("vote.iost", "RemoveOption", `["1", "option2", false]`, kp.ID, kp) // should fail
			s.Call("vote.iost", "RemoveOption", `["1", "option1", false]`, kp.ID, kp)
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"option1", "option2", "option3", "option4", "option5"})
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option5")), ShouldEqual, `["0",true,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option4")), ShouldEqual, `["400",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option3")), ShouldEqual, `["300",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option2")), ShouldEqual, `["300",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option1")), ShouldEqual, `["100",true,-1]`)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{"option2", "option3", "option4"})
		})

		Convey("test RemoveOption with force", func() {
			kp2, _ := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			kp3, _ := account.NewKeyPair(common.Base58Decode(testID[5]), crypto.Secp256k1)
			// vote
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "200"]`, testID[2]), kp2.ID, kp2)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[0]), kp.ID, kp)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, testID[0]), kp.ID, kp)
			s.Call("vote.iost", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, testID[4]), kp3.ID, kp3)

			// add option
			s.Call("vote.iost", "AddOption", `["1", "option5", false]`, kp.ID, kp)
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"option1", "option2", "option3", "option4", "option5"})

			// remove option
			s.Call("vote.iost", "RemoveOption", `["1", "option5", true]`, kp.ID, kp)
			s.Call("vote.iost", "RemoveOption", `["1", "option4", true]`, kp.ID, kp)
			s.Call("vote.iost", "RemoveOption", `["1", "option3", true]`, kp.ID, kp)
			s.Call("vote.iost", "RemoveOption", `["1", "option2", true]`, kp.ID, kp)
			s.Call("vote.iost", "RemoveOption", `["1", "option1", true]`, kp.ID, kp)
			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"option1", "option2", "option3", "option4", "option5"})
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option5")), ShouldEqual, `["0",true,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option4")), ShouldEqual, `["400",true,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option3")), ShouldEqual, `["300",true,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option2")), ShouldEqual, `["300",true,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "option1")), ShouldEqual, `["100",true,-1]`)
			So(s.Visitor.MKeys("vote.iost-p_1"), ShouldResemble, []string{})
		})
	})
}
