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
	. "github.com/smartystreets/goconvey/convey"
)

func prepareToken(t *testing.T, s *Simulator, kp *account.KeyPair) {
	r, err := s.Call("iost.token", "create", fmt.Sprintf(`["%v", "%v", %v, {}]`, "iost", testID[0], "21000000000"), kp.ID, kp)
	if err != nil || r.Status.Code != tx.Success {
		t.Fatal(err, r)
	}
	for i := 0; i < 18; i += 2 {
		s.Call("iost.token", "issue", fmt.Sprintf(`["%v", "%v", "%v"]`, "iost", testID[i], "2000000000"), kp.ID, kp)
	}
	s.Visitor.Commit()
}

func prepareVote(t *testing.T, s *Simulator, kp *account.KeyPair) (*tx.TxReceipt, error) {
	// deploy iost.vote
	setNonNativeContract(s, "iost.vote", "vote_common.js", ContractPath)
	s.Call("iost.vote", "init", `[]`, kp.ID, kp)

	// deploy voteresult
	err := setNonNativeContract(s, "Contractvoteresult", "voteresult.js", "./test_data/")
	if err != nil {
		t.Fatal(err)
	}

	config := make(map[string]interface{})
	config["resultNumber"] = 2
	config["minVote"] = 10
	config["options"] = []string{"option1", "option2", "option3", "option4"}
	config["anyOption"] = false
	config["unvoteInterval"] = 0
	params := []interface{}{
		testID[0],
		"test vote",
		config,
	}
	b, _ := json.Marshal(params)
	r, err := s.Call("iost.vote", "NewVote", string(b), kp.ID, kp)
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

		prepareContract(t, s)
		prepareToken(t, s, kp)

		Convey("test NewVote", func() {
			r, err := prepareVote(t, s, kp)

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.Visitor.TokenBalance("iost", testID[0]), ShouldEqual, int64(1999999000*1e8))
			So(s.Visitor.Get("iost.vote-current_id"), ShouldEqual, `s"1"`)
			So(s.Visitor.MGet("iost.vote-voteInfo", "1"), ShouldEqual, `s{"description":"test vote","resultNumber":2,"minVote":10,"anyOption":false,"unvoteInterval":0,"deposit":"1000"}`)
			So(s.Visitor.MGet("iost.vote-v-1", "option1"), ShouldEqual, `s["0",false,-1]`)

			r, err = s.Call("Contractvoteresult", "GetResult", `["1"]`, kp.ID, kp)

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.Visitor.MGet("Contractvoteresult-vote-result", "1"), ShouldEqual, `s[]`)
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

		prepareContract(t, s)
		prepareToken(t, s, kp)
		prepareVote(t, s, kp)

		Convey("test AddOption", func() {
			r, err := s.Call("iost.vote", "AddOption", `["1", "option5", true]`, kp.ID, kp)

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.Visitor.MKeys("iost.vote-v-1"), ShouldResemble, []string{"option1", "option2", "option3", "option4", "option5"})
			So(s.Visitor.MGet("iost.vote-v-1", "option5"), ShouldEqual, `s["0",false,-1]`)
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

		prepareContract(t, s)
		prepareToken(t, s, kp)
		prepareVote(t, s, kp)

		Convey("test RemoveOption", func() {
			r, err := s.Call("iost.vote", "RemoveOption", `["1", "option2", true]`, kp.ID, kp)

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.Visitor.MKeys("iost.vote-v-1"), ShouldResemble, []string{"option1", "option2", "option3", "option4"})
			So(s.Visitor.MGet("iost.vote-v-1", "option2"), ShouldEqual, `s["0",true,-1]`)
		})
	})
}

func Test_Vote(t *testing.T) {
	ilog.Stop()
	Convey("test vote", t, func() {
		s := NewSimulator()
		defer s.Clear()

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(t, s)
		prepareToken(t, s, kp)
		prepareVote(t, s, kp)

		Convey("test Vote", func() {
			kp2, _ := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option3", "5"]`, testID[2]), kp2.ID, kp2)

			So(s.Visitor.MGet("iost.vote-v-1", "option3"), ShouldEqual, `s["5",false,-1]`)
			So(s.Visitor.MGet("iost.vote-u-1", testID[2]), ShouldEqual, `s{"option3":["5",0,"0"]}`)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{})

			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option3", "5"]`, testID[2]), kp2.ID, kp2)
			So(s.Visitor.MGet("iost.vote-v-1", "option3"), ShouldEqual, `s["10",false,-1]`)
			So(s.Visitor.MGet("iost.vote-u-1", testID[2]), ShouldEqual, `s{"option3":["10",0,"0"]}`)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{"option3"})

			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option1", "20"]`, testID[2]), kp2.ID, kp2)
			So(s.Visitor.MGet("iost.vote-v-1", "option1"), ShouldEqual, `s["20",false,-1]`)
			So(s.Visitor.MGet("iost.vote-u-1", testID[2]), ShouldEqual, `s{"option3":["10",0,"0"],"option1":["20",0,"0"]}`)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{"option3", "option1"})

			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option3", "100"]`, testID[0]), kp.ID, kp)
			So(s.Visitor.MGet("iost.vote-v-1", "option3"), ShouldEqual, `s["110",false,-1]`)
			So(s.Visitor.MGet("iost.vote-u-1", testID[0]), ShouldEqual, `s{"option3":["100",0,"0"]}`)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{"option3", "option1"})

			s.Call("Contractvoteresult", "GetResult", `["1"]`, kp.ID, kp)
			So(s.Visitor.MGet("Contractvoteresult-vote-result", "1"), ShouldEqual, `s[{"option":"option3","votes":"110"},{"option":"option1","votes":"20"}]`)
		})

		Convey("test Unvote", func() {
			kp2, _ := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			kp3, _ := account.NewKeyPair(common.Base58Decode(testID[5]), crypto.Secp256k1)
			// vote
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{"option1"})

			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[0]), kp.ID, kp)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{"option1", "option2"})

			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, testID[0]), kp.ID, kp)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{"option1", "option2", "option3"})

			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, testID[4]), kp3.ID, kp3)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{"option1", "option2", "option3", "option4"})

			// get result
			s.Call("Contractvoteresult", "GetResult", `["1"]`, kp.ID, kp)
			So(s.Visitor.MGet("Contractvoteresult-vote-result", "1"), ShouldEqual, `s[{"option":"option4","votes":"400"},{"option":"option3","votes":"300"}]`)

			// unvote
			s.Call("iost.vote", "Unvote", fmt.Sprintf(`["1", "%v", "option3", "100"]`, testID[0]), kp.ID, kp)
			So(s.Visitor.MGet("iost.vote-v-1", "option3"), ShouldEqual, `s["200",false,-1]`)
			So(s.Visitor.MGet("iost.vote-u-1", testID[0]), ShouldEqual, `s{"option2":["100",0,"0"],"option3":["200",0,"0"]}`)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{"option1", "option2", "option3", "option4"})

			// get result
			s.Call("Contractvoteresult", "GetResult", `["1"]`, kp.ID, kp)
			So(s.Visitor.MGet("Contractvoteresult-vote-result", "1"), ShouldEqual, `s[{"option":"option4","votes":"400"},{"option":"option2","votes":"200"}]`)

			// unvote again
			s.Call("iost.vote", "Unvote", fmt.Sprintf(`["1", "%v", "option2", "95"]`, testID[0]), kp.ID, kp)
			So(s.Visitor.MGet("iost.vote-v-1", "option2"), ShouldEqual, `s["105",false,-1]`)
			So(s.Visitor.MGet("iost.vote-u-1", testID[0]), ShouldEqual, `s{"option2":["5",0,"0"],"option3":["200",0,"0"]}`)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{"option1", "option2", "option3", "option4"})

			s.Call("iost.vote", "Unvote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[2]), kp2.ID, kp2)
			So(s.Visitor.MGet("iost.vote-v-1", "option2"), ShouldEqual, `s["5",false,-1]`)
			So(s.Visitor.MGet("iost.vote-u-1", testID[2]), ShouldEqual, `s{"option1":["100",0,"0"]}`)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{"option1", "option3", "option4"})

			// get result
			s.Call("Contractvoteresult", "GetResult", `["1"]`, kp.ID, kp)
			So(s.Visitor.MGet("Contractvoteresult-vote-result", "1"), ShouldEqual, `s[{"option":"option4","votes":"400"},{"option":"option3","votes":"200"}]`)
		})
	})
}

func Test_DelVote(t *testing.T) {
	ilog.Stop()
	Convey("test DelVote", t, func() {
		s := NewSimulator()
		defer s.Clear()

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(t, s)
		prepareToken(t, s, kp)
		prepareVote(t, s, kp)

		Convey("test delete vote", func() {
			kp2, _ := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			kp3, _ := account.NewKeyPair(common.Base58Decode(testID[5]), crypto.Secp256k1)
			// vote
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[0]), kp.ID, kp)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, testID[0]), kp.ID, kp)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, testID[4]), kp3.ID, kp3)

			// del vote
			s.Call("iost.vote", "DelVote", `["1"]`, kp.ID, kp)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{})
			So(s.Visitor.MKeys("iost.vote-v-1"), ShouldResemble, []string{})

			// unvote part
			s.Call("iost.vote", "Unvote", fmt.Sprintf(`["1", "%v", "option2", "95"]`, testID[0]), kp.ID, kp)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{})
			So(s.Visitor.MKeys("iost.vote-v-1"), ShouldResemble, []string{})
			So(s.Visitor.MGet("iost.vote-u-1", testID[0]), ShouldEqual, `s{"option2":["5",0,"0"],"option3":["300",0,"0"]}`)

			// unvote all
			s.Call("iost.vote", "Unvote", fmt.Sprintf(`["1", "%v", "option2", "5"]`, testID[0]), kp.ID, kp)
			s.Call("iost.vote", "Unvote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, testID[0]), kp.ID, kp)
			So(s.Visitor.MHas("iost.vote-u-1", testID[0]), ShouldEqual, false)
		})
	})
}

func Test_MixVoteOption(t *testing.T) {
	ilog.Stop()
	Convey("test mixed", t, func() {
		s := NewSimulator()
		defer s.Clear()

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(t, s)
		prepareToken(t, s, kp)
		prepareVote(t, s, kp)

		Convey("test AddOption not clear", func() {
			kp2, _ := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			kp3, _ := account.NewKeyPair(common.Base58Decode(testID[5]), crypto.Secp256k1)
			// vote
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[0]), kp.ID, kp)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, testID[0]), kp.ID, kp)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, testID[4]), kp3.ID, kp3)

			// add option
			s.Call("iost.vote", "AddOption", `["1", "option5", false]`, kp.ID, kp)
			So(s.Visitor.MKeys("iost.vote-v-1"), ShouldResemble, []string{"option1", "option2", "option3", "option4", "option5"})

			// get result
			s.Call("Contractvoteresult", "GetResult", `["1"]`, kp.ID, kp)
			So(s.Visitor.MGet("Contractvoteresult-vote-result", "1"), ShouldEqual, `s[{"option":"option4","votes":"400"},{"option":"option3","votes":"300"}]`)

			s.Head.Number++
			// remove option
			s.Call("iost.vote", "RemoveOption", `["1", "option1", false]`, kp.ID, kp)
			So(s.Visitor.MGet("iost.vote-v-1", "option1"), ShouldEqual, `s["100",true,-1]`)

			// add option
			s.Call("iost.vote", "AddOption", `["1", "option1", false]`, kp.ID, kp)
			So(s.Visitor.MGet("iost.vote-v-1", "option1"), ShouldEqual, `s["100",false,-1]`)
			So(s.Visitor.MGet("iost.vote-p-1", "option1"), ShouldEqual, `s"100"`)
		})

		Convey("test AddOption and clear", func() {
			kp2, _ := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			kp3, _ := account.NewKeyPair(common.Base58Decode(testID[5]), crypto.Secp256k1)
			// vote
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[0]), kp.ID, kp)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option3", "200"]`, testID[0]), kp.ID, kp)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, testID[4]), kp3.ID, kp3)

			s.Head.Number++
			// remove option
			s.Call("iost.vote", "RemoveOption", `["1", "option1", false]`, kp.ID, kp)
			// add option
			s.Call("iost.vote", "AddOption", `["1", "option1", true]`, kp.ID, kp)

			So(s.Visitor.MKeys("iost.vote-v-1"), ShouldResemble, []string{"option1", "option2", "option3", "option4"})
			So(s.Visitor.MGet("iost.vote-v-1", "option1"), ShouldEqual, `s["0",false,1]`)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{"option2", "option3", "option4"})

			// vote after clear in same block
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			So(s.Visitor.MHas("iost.vote-p-1", "option1"), ShouldEqual, false)
			So(s.Visitor.MGet("iost.vote-u-1", testID[2]), ShouldEqual, `s{"option1":["200",1,"100"],"option2":["100",0,"0"]}`)

			// vote after the clear block
			s.Head.Number++
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			So(s.Visitor.MGet("iost.vote-p-1", "option1"), ShouldEqual, `s"100"`)
			So(s.Visitor.MGet("iost.vote-u-1", testID[2]), ShouldEqual, `s{"option1":["300",2,"200"],"option2":["100",0,"0"]}`)

			// vote again
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			So(s.Visitor.MGet("iost.vote-u-1", testID[2]), ShouldEqual, `s{"option1":["400",2,"200"],"option2":["100",0,"0"]}`)

			// get result
			s.Call("Contractvoteresult", "GetResult", `["1"]`, kp.ID, kp)
			So(s.Visitor.MGet("Contractvoteresult-vote-result", "1"), ShouldEqual, `s[{"option":"option4","votes":"400"},{"option":"option2","votes":"200"}]`)

			// unvote
			s.Call("iost.vote", "Unvote", fmt.Sprintf(`["1", "%v", "option1", "50"]`, testID[2]), kp2.ID, kp2)
			So(s.Visitor.MGet("iost.vote-u-1", testID[2]), ShouldEqual, `s{"option1":["350",2,"150"],"option2":["100",0,"0"]}`)

			// unvote again
			s.Call("iost.vote", "Unvote", fmt.Sprintf(`["1", "%v", "option1", "200"]`, testID[2]), kp2.ID, kp2)
			So(s.Visitor.MGet("iost.vote-u-1", testID[2]), ShouldEqual, `s{"option1":["150",2,"0"],"option2":["100",0,"0"]}`)
			So(s.Visitor.MGet("iost.vote-p-1", "option1"), ShouldEqual, `s"150"`)
		})

		Convey("test RemoveOption not force", func() {
			kp2, _ := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			kp3, _ := account.NewKeyPair(common.Base58Decode(testID[5]), crypto.Secp256k1)
			// vote
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option2", "200"]`, testID[2]), kp2.ID, kp2)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[0]), kp.ID, kp)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, testID[0]), kp.ID, kp)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, testID[4]), kp3.ID, kp3)

			// add option
			s.Call("iost.vote", "AddOption", `["1", "option5", false]`, kp.ID, kp)
			So(s.Visitor.MKeys("iost.vote-v-1"), ShouldResemble, []string{"option1", "option2", "option3", "option4", "option5"})

			// remove option
			s.Call("iost.vote", "RemoveOption", `["1", "option5", false]`, kp.ID, kp)
			// s.Call("iost.vote", "RemoveOption", `["1", "option4", false]`, kp.ID, kp) // should fail
			// s.Call("iost.vote", "RemoveOption", `["1", "option3", false]`, kp.ID, kp) // should fail
			// s.Call("iost.vote", "RemoveOption", `["1", "option2", false]`, kp.ID, kp) // should fail
			s.Call("iost.vote", "RemoveOption", `["1", "option1", false]`, kp.ID, kp)
			So(s.Visitor.MKeys("iost.vote-v-1"), ShouldResemble, []string{"option1", "option2", "option3", "option4", "option5"})
			So(s.Visitor.MGet("iost.vote-v-1", "option5"), ShouldEqual, `s["0",true,-1]`)
			So(s.Visitor.MGet("iost.vote-v-1", "option4"), ShouldEqual, `s["400",false,-1]`)
			So(s.Visitor.MGet("iost.vote-v-1", "option3"), ShouldEqual, `s["300",false,-1]`)
			So(s.Visitor.MGet("iost.vote-v-1", "option2"), ShouldEqual, `s["300",false,-1]`)
			So(s.Visitor.MGet("iost.vote-v-1", "option1"), ShouldEqual, `s["100",true,-1]`)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{"option2", "option3", "option4"})
		})

		Convey("test RemoveOption with force", func() {
			kp2, _ := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			kp3, _ := account.NewKeyPair(common.Base58Decode(testID[5]), crypto.Secp256k1)
			// vote
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option1", "100"]`, testID[2]), kp2.ID, kp2)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option2", "200"]`, testID[2]), kp2.ID, kp2)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option2", "100"]`, testID[0]), kp.ID, kp)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option3", "300"]`, testID[0]), kp.ID, kp)
			s.Call("iost.vote", "Vote", fmt.Sprintf(`["1", "%v", "option4", "400"]`, testID[4]), kp3.ID, kp3)

			// add option
			s.Call("iost.vote", "AddOption", `["1", "option5", false]`, kp.ID, kp)
			So(s.Visitor.MKeys("iost.vote-v-1"), ShouldResemble, []string{"option1", "option2", "option3", "option4", "option5"})

			// remove option
			s.Call("iost.vote", "RemoveOption", `["1", "option5", true]`, kp.ID, kp)
			s.Call("iost.vote", "RemoveOption", `["1", "option4", true]`, kp.ID, kp)
			s.Call("iost.vote", "RemoveOption", `["1", "option3", true]`, kp.ID, kp)
			s.Call("iost.vote", "RemoveOption", `["1", "option2", true]`, kp.ID, kp)
			s.Call("iost.vote", "RemoveOption", `["1", "option1", true]`, kp.ID, kp)
			So(s.Visitor.MKeys("iost.vote-v-1"), ShouldResemble, []string{"option1", "option2", "option3", "option4", "option5"})
			So(s.Visitor.MGet("iost.vote-v-1", "option5"), ShouldEqual, `s["0",true,-1]`)
			So(s.Visitor.MGet("iost.vote-v-1", "option4"), ShouldEqual, `s["400",true,-1]`)
			So(s.Visitor.MGet("iost.vote-v-1", "option3"), ShouldEqual, `s["300",true,-1]`)
			So(s.Visitor.MGet("iost.vote-v-1", "option2"), ShouldEqual, `s["300",true,-1]`)
			So(s.Visitor.MGet("iost.vote-v-1", "option1"), ShouldEqual, `s["100",true,-1]`)
			So(s.Visitor.MKeys("iost.vote-p-1"), ShouldResemble, []string{})
		})
	})
}
