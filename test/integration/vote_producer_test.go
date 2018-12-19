package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	. "github.com/iost-official/go-iost/verifier"
	"github.com/iost-official/go-iost/vm/database"
	. "github.com/smartystreets/goconvey/convey"
)


func initProducer(s *Simulator) {
	for _, acc := range testAccounts[:6] {
		r, err := s.Call("vote_producer.iost", "InitProducer", fmt.Sprintf(`["%v", "%v"]`, acc.ID, acc.KeyPair.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
	}
}

func prepareFakeBase(t *testing.T, s *Simulator) {
	// deploy fake base.iost
	err := setNonNativeContract(s, "base.iost", "base.js", "./test_data/")
	if err != nil {
		t.Fatal(err)
	}
}

func prepareProducerVote(t *testing.T, s *Simulator, acc *TestAccount) {
	// deploy vote.iost
	setNonNativeContract(s, "vote.iost", "vote_common.js", ContractPath)
	r, err := s.Call("vote.iost", "init", `[]`, acc.ID, acc.KeyPair)
	if err != nil || r.Status.Code != tx.Success {
		t.Fatal(err, r)
	}

	// deploy vote_producer.iost
	setNonNativeContract(s, "vote_producer.iost", "vote.js", ContractPath)

	r, err = s.Call("token.iost", "issue", fmt.Sprintf(`["%v", "%v", "%v"]`, "iost", "vote_producer.iost", "1000"), acc.ID, acc.KeyPair)
	if err != nil || r.Status.Code != tx.Success {
		t.Fatal(err, r)
	}

	r, err = s.Call("vote_producer.iost", "init", `[]`, acc.ID, acc.KeyPair)
	if err != nil || r.Status.Code != tx.Success {
		t.Fatal(err, r)
	}

	s.Visitor.Commit()
}

func prepareNewProducerVote(t *testing.T, s *Simulator, acc1 *TestAccount) {
	s.Head.Number = 0
	// deploy vote.iost
	setNonNativeContract(s, "vote.iost", "vote_common.js", ContractPath)
	r, err := s.Call("vote.iost", "init", `[]`, acc1.ID, acc1.KeyPair)
	if err != nil || r.Status.Code != tx.Success {
		t.Fatal(err, r)
	}

	// deploy vote_producer.iost
	setNonNativeContract(s, "vote_producer.iost", "vote_producer.js", ContractPath)
	r, err = s.Call("vote_producer.iost", "InitAdmin", fmt.Sprintf(`["%v"]`, acc1.ID), acc1.ID, acc1.KeyPair)
	if err != nil || r.Status.Code != tx.Success {
		t.Fatal(err, r)
	}

	r, err = s.Call("token.iost", "issue", fmt.Sprintf(`["%v", "%v", "%v"]`, "iost", "vote_producer.iost", "1000"), acc1.ID, acc1.KeyPair)
	if err != nil || r.Status.Code != tx.Success {
		t.Fatal(err, r)
	}

	r, err = s.Call("vote_producer.iost", "init", `[]`, acc1.ID, acc1.KeyPair)
	if err != nil || r.Status.Code != tx.Success {
		t.Fatal(err, r)
	}

	s.Visitor.Commit()
}

func Test_InitProducer(t *testing.T) {
	ilog.Stop()
	Convey("test InitProducer", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareToken(t, s, acc0)
		prepareProducerVote(t, s, acc0)

		So(database.MustUnmarshal(s.Visitor.Get("vote.iost-current_id")), ShouldEqual, `"1"`)
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-voteId")), ShouldEqual, `"1"`)
		Convey("test init producer", func() {
			initProducer(s)
			list, _ := json.Marshal([]string{acc0.KeyPair.ID, acc3.KeyPair.ID, acc1.KeyPair.ID, acc4.KeyPair.ID, acc5.KeyPair.ID, acc2.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(list))

			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{acc0.ID, acc1.ID, acc2.ID, acc3.ID, acc4.ID, acc5.ID})
		})
	})
}

func Test_RegisterProducer(t *testing.T) {
	ilog.Stop()
	Convey("test RegisterProducer", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareToken(t, s, acc0)
		prepareProducerVote(t, s, acc0)
		initProducer(s)

		Convey("test register/unregister", func() {
			r, err := s.Call("vote_producer.iost", "RegisterProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc6.ID, acc6.KeyPair.ID), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote_producer.iost-producerTable", acc6.ID)), ShouldEqual, fmt.Sprintf(`{"pubkey":"%s","loc":"loc","url":"url","netId":"netId","online":false,"registerFee":"200000000"}`, acc6.KeyPair.ID))
			So(s.Visitor.TokenBalance("iost", acc6.ID), ShouldEqual, int64(1800000000*1e8))
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc6.ID)), ShouldEqual, `["0",false,-1]`)

			r, err = s.Call("vote_producer.iost", "UnregisterProducer", fmt.Sprintf(`["%v"]`, acc6.ID), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(s.Visitor.MHas("vote_producer.iost-producerTable", acc6.ID), ShouldEqual, false)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc6.ID)), ShouldEqual, `["0",true,-1]`)
		})
	})
}

func Test_LogInOut(t *testing.T) {
	ilog.Stop()
	Convey("test RegisterProducer", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareToken(t, s, acc0)
		prepareProducerVote(t, s, acc0)
		initProducer(s)

		Convey("test login/logout", func() {
			r, err := s.Call("vote_producer.iost", "RegisterProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc6.ID, acc6.KeyPair.ID), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, acc6.ID), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote_producer.iost-producerTable", acc6.ID)), ShouldEqual, fmt.Sprintf(`{"pubkey":"%s","loc":"loc","url":"url","netId":"netId","online":true,"registerFee":"200000000"}`, acc6.KeyPair.ID))

			r, err = s.Call("vote_producer.iost", "LogOutProducer", fmt.Sprintf(`["%v"]`, acc6.ID), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote_producer.iost-producerTable", acc6.ID)), ShouldEqual, fmt.Sprintf(`{"pubkey":"%s","loc":"loc","url":"url","netId":"netId","online":false,"registerFee":"200000000"}`, acc6.KeyPair.ID))

			r, _ = s.Call("vote_producer.iost", "LogOutProducer", fmt.Sprintf(`["%v"]`, acc0.ID), acc0.ID, acc0.KeyPair)
			So(r.Status.Message, ShouldNotEqual, "")
			So(r.Status.Code, ShouldEqual, 4)
			So(r.Status.Message, ShouldContainSubstring, "producer in pending list or in current list, can't logout")
		})
	})
}

func Test_Vote1(t *testing.T) {
	ilog.Stop()
	Convey("test Vote", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareFakeBase(t, s)
		prepareToken(t, s, acc0)
		prepareProducerVote(t, s, acc0)
		initProducer(s)

		Convey("test vote/unvote", func() {
			s.Head.Number = 1
			r, err := s.Call("vote_producer.iost", "RegisterProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc6.ID, acc6.KeyPair.ID), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "RegisterProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc7.ID, acc7.KeyPair.ID), acc7.ID, acc7.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "RegisterProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc8.ID, acc8.KeyPair.ID), acc8.ID, acc8.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, acc6.ID), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, acc7.ID), acc7.ID, acc7.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, acc8.ID), acc8.ID, acc8.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc6.ID)), ShouldEqual, `["0",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc7.ID)), ShouldEqual, `["0",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc8.ID)), ShouldEqual, `["0",false,-1]`)

			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc6.ID, "100000000"), acc0.ID, acc0.KeyPair)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc6.ID)), ShouldEqual, `["100000000",false,-1]`)

			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc7.ID, acc6.ID, "100000000"), acc7.ID, acc7.KeyPair)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc6.ID)), ShouldEqual, `["200000000",false,-1]`)
			So(s.Visitor.MHas("vote.iost-p-1", acc6.ID), ShouldEqual, false)

			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc8.ID, acc6.ID, "100000000"), acc8.ID, acc8.KeyPair)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc6.ID)), ShouldEqual, `["300000000",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-p_1", acc6.ID)), ShouldEqual, `"300000000"`)

			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc7.ID, "215000000"), acc0.ID, acc0.KeyPair)
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc8.ID, "220000000"), acc0.ID, acc0.KeyPair)

			r, err = s.Call("vote_producer.iost", "GetProducer", fmt.Sprintf(`["%v"]`, acc6.ID), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(r.Returns[0], ShouldEqual, fmt.Sprintf(`["{\"pubkey\":\"%s\",\"loc\":\"loc\",\"url\":\"url\",\"netId\":\"netId\",\"online\":true,\"registerFee\":\"200000000\",\"voteInfo\":{\"votes\":\"300000000\",\"deleted\":false,\"clearTime\":-1}}"]`, acc6.KeyPair.ID))

			r, err = s.Call("vote_producer.iost", "GetVote", fmt.Sprintf(`["%v"]`, acc0.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(r.Returns[0], ShouldEqual, `["[{\"option\":\"user_6\",\"votes\":\"100000000\",\"voteTime\":1,\"clearedVotes\":\"0\"},{\"option\":\"user_7\",\"votes\":\"215000000\",\"voteTime\":1,\"clearedVotes\":\"0\"},{\"option\":\"user_8\",\"votes\":\"220000000\",\"voteTime\":1,\"clearedVotes\":\"0\"}]"]`)

			// do stat
			s.Head.Number = 200
			r, err = s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: 0				, 0
			// 2	: 0				, 0
			// 3	: 0				, 0
			// 4	: 0				, 0
			// 5	: 0				, 0
			// 6	: q^1*90000000	, 300000000
			// 7	: 1*5000000		, 215000000
			// 8	: 1*10000000	, 220000000
			// 0, 3, 1, 4, 5, 2
			currentList, _ := json.Marshal([]string{acc0.KeyPair.ID, acc3.KeyPair.ID, acc1.KeyPair.ID, acc4.KeyPair.ID, acc5.KeyPair.ID, acc2.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 6, 0, 3, 1, 4, 5
			pendingList, _ := json.Marshal([]string{acc6.KeyPair.ID, acc0.KeyPair.ID, acc3.KeyPair.ID, acc1.KeyPair.ID, acc4.KeyPair.ID, acc5.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc1.ID, "240000000"), acc0.ID, acc0.KeyPair)
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc2.ID, "230000000"), acc0.ID, acc0.KeyPair)
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc3.ID, "260000000"), acc0.ID, acc0.KeyPair)
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc4.ID, "250000000"), acc0.ID, acc0.KeyPair)
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: 0				, 240000000
			// 2	: q^1*20000000	, 230000000
			// 3	: 0				, 260000000
			// 4	: 0				, 250000000
			// 5	: 0				, 0
			// 6	: q^2*90000000	, 300000000
			// 7	: 10000000		, 215000000
			// 8	: 20000000		, 220000000
			// 6, 0, 3, 1, 4, 5
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 6, 2, 0, 3, 1, 4
			pendingList, _ = json.Marshal([]string{acc6.KeyPair.ID, acc2.KeyPair.ID, acc0.KeyPair.ID, acc3.KeyPair.ID, acc1.KeyPair.ID, acc4.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: 0				, 240000000
			// 2	: q^2*20000000	, 230000000
			// 3	: 0				, 260000000
			// 4	: 0				, 250000000
			// 5	: 0				, 0
			// 6	: q^3*90000000	, 240000000
			// 7	: 15000000		, 215000000
			// 8	: q^1*30000000	, 220000000
			// 6, 2, 0, 3, 1, 4
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 6, 8, 2, 0, 3, 1
			pendingList, _ = json.Marshal([]string{acc6.KeyPair.ID, acc8.KeyPair.ID, acc2.KeyPair.ID, acc0.KeyPair.ID, acc3.KeyPair.ID, acc1.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: 0				, 240000000
			// 2	: q^3*20000000	, 230000000
			// 3	: 0				, 260000000
			// 4	: q^1*40000000	, 250000000
			// 5	: 0				, 0
			// 6	: q^4*90000000	, 240000000
			// 7	: 20000000		, 215000000
			// 8	: q^2*30000000	, 220000000
			// 6, 8, 2, 0, 3, 1
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 6, 4, 8, 2, 0, 3
			pendingList, _ = json.Marshal([]string{acc6.KeyPair.ID, acc4.KeyPair.ID, acc8.KeyPair.ID, acc2.KeyPair.ID, acc0.KeyPair.ID, acc3.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: q^1*30000000	, 240000000
			// 2	: q^4*20000000	, 230000000
			// 3	: 0				, 260000000
			// 4	: q^2*40000000	, 250000000
			// 5	: 0				, 0
			// 6	: q^5*90000000	, 240000000
			// 7	: 25000000		, 215000000
			// 8	: q^3*30000000	, 220000000
			// 6, 4, 8, 2, 0, 3
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 6, 4, 1, 8, 2, 0
			pendingList, _ = json.Marshal([]string{acc6.KeyPair.ID, acc4.KeyPair.ID, acc1.KeyPair.ID, acc8.KeyPair.ID, acc2.KeyPair.ID, acc0.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: q^2*30000000	, 240000000
			// 2	: q^5*20000000	, 230000000
			// 3	: q^1*50000000	, 260000000
			// 4	: q^3*40000000	, 250000000
			// 5	: 0				, 0
			// 6	: q^6*90000000	, 240000000
			// 7	: 30000000		, 215000000
			// 8	: q^4*30000000	, 220000000
			// 6, 4, 8, 1, 2, 0
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 6, 3, 4, 1, 8, 2
			pendingList, _ = json.Marshal([]string{acc6.KeyPair.ID, acc3.KeyPair.ID, acc4.KeyPair.ID, acc1.KeyPair.ID, acc8.KeyPair.ID, acc2.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: q^3*30000000	, 240000000
			// 2	: 0				, 230000000
			// 3	: q^2*50000000	, 260000000
			// 4	: q^4*40000000	, 250000000
			// 5	: 0				, 0
			// 6	: q^7*90000000	, 240000000
			// 7	: q^1*35000000	, 215000000
			// 8	: q^4*30000000	, 220000000
			// 6, 3, 4, 1, 8, 2
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 6, 3, 4, 7, 1, 8
			pendingList, _ = json.Marshal([]string{acc6.KeyPair.ID, acc3.KeyPair.ID, acc4.KeyPair.ID, acc7.KeyPair.ID, acc1.KeyPair.ID, acc8.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: q^4*30000000	, 240000000
			// 2	: 20000000		, 230000000
			// 3	: q^3*50000000	, 260000000
			// 4	: q^5*40000000	, 250000000
			// 5	: 0				, 0
			// 6	: q^8*90000000	, 240000000
			// 7	: q^2*35000000	, 215000000
			// 8	: q^5*30000000	, 220000000
			// 6, 3, 4, 7, 1, 8
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 6, 3, 4, 7, 1, 8
			pendingList, _ = json.Marshal([]string{acc6.KeyPair.ID, acc3.KeyPair.ID, acc4.KeyPair.ID, acc7.KeyPair.ID, acc1.KeyPair.ID, acc8.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: q^5*30000000	, 240000000
			// 2	: q^1*40000000	, 230000000
			// 3	: q^4*50000000	, 260000000
			// 4	: q^6*40000000	, 250000000
			// 5	: 0				, 0
			// 6	: q^9*90000000	, 240000000
			// 7	: q^3*35000000	, 215000000
			// 8	: 0				, 220000000
			// 6, 3, 4, 7, 1, 8
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 6, 3, 2, 4, 7, 1
			pendingList, _ = json.Marshal([]string{acc6.KeyPair.ID, acc3.KeyPair.ID, acc2.KeyPair.ID, acc4.KeyPair.ID, acc7.KeyPair.ID, acc1.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))
		})
	})
}

func Test_Vote2(t *testing.T) {
	ilog.Stop()
	Convey("test Vote2", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareFakeBase(t, s)
		prepareToken(t, s, acc0)
		prepareNewProducerVote(t, s, acc0)
		initProducer(s)

		Convey("test new vote/unvote", func() {
			s.Head.Number = 1
			r, err := s.Call("vote_producer.iost", "ApplyRegister", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc6.ID, acc6.KeyPair.ID), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "ApproveRegister", fmt.Sprintf(`["%v"]`, acc6.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "ApplyRegister", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc7.ID, acc7.KeyPair.ID), acc7.ID, acc7.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "ApproveRegister", fmt.Sprintf(`["%v"]`, acc7.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "ApplyRegister", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc8.ID, acc8.KeyPair.ID), acc8.ID, acc8.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "ApproveRegister", fmt.Sprintf(`["%v"]`, acc8.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, acc6.ID), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, acc7.ID), acc7.ID, acc7.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, acc8.ID), acc8.ID, acc8.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc6.ID)), ShouldEqual, `["0",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc7.ID)), ShouldEqual, `["0",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc8.ID)), ShouldEqual, `["0",false,-1]`)

			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc6.ID, "100000000"), acc0.ID, acc0.KeyPair)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc6.ID)), ShouldEqual, `["100000000",false,-1]`)

			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc7.ID, acc6.ID, "100000000"), acc7.ID, acc7.KeyPair)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc6.ID)), ShouldEqual, `["200000000",false,-1]`)
			So(s.Visitor.MHas("vote.iost-p-1", acc6.ID), ShouldEqual, false)

			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc8.ID, acc6.ID, "100000000"), acc8.ID, acc8.KeyPair)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc6.ID)), ShouldEqual, `["300000000",false,-1]`)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-p_1", acc6.ID)), ShouldEqual, `"300000000"`)

			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc7.ID, "215000000"), acc0.ID, acc0.KeyPair)
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc8.ID, "220000000"), acc0.ID, acc0.KeyPair)

			r, err = s.Call("vote_producer.iost", "GetProducer", fmt.Sprintf(`["%v"]`, acc6.ID), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(r.Returns[0], ShouldEqual, fmt.Sprintf(`["{\"pubkey\":\"%s\",\"loc\":\"loc\",\"url\":\"url\",\"netId\":\"netId\",\"status\":1,\"online\":true,\"voteInfo\":{\"votes\":\"300000000\",\"deleted\":false,\"clearTime\":-1}}"]`, acc6.KeyPair.ID))

			r, err = s.Call("vote_producer.iost", "GetVote", fmt.Sprintf(`["%v"]`, acc0.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(r.Returns[0], ShouldEqual, `["[{\"option\":\"user_6\",\"votes\":\"100000000\",\"voteTime\":1,\"clearedVotes\":\"0\"},{\"option\":\"user_7\",\"votes\":\"215000000\",\"voteTime\":1,\"clearedVotes\":\"0\"},{\"option\":\"user_8\",\"votes\":\"220000000\",\"voteTime\":1,\"clearedVotes\":\"0\"}]"]`)

			// do stat
			// q = 0.9
			s.Head.Number = 200
			r, err = s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: 0				, 0
			// 2	: 0				, 0
			// 3	: 0				, 0
			// 4	: 0				, 0
			// 5	: 0				, 0
			// 6	: q^1*300000000	, 300000000
			// 7	: 1*215000000	, 215000000
			// 8	: 1*220000000	, 220000000
			// 0, 3, 1, 4, 5, 2
			currentList, _ := json.Marshal([]string{acc0.KeyPair.ID, acc3.KeyPair.ID, acc1.KeyPair.ID, acc4.KeyPair.ID, acc5.KeyPair.ID, acc2.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 6, 0, 3, 1, 4, 5
			pendingList, _ := json.Marshal([]string{acc6.KeyPair.ID, acc0.KeyPair.ID, acc3.KeyPair.ID, acc1.KeyPair.ID, acc4.KeyPair.ID, acc5.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc1.ID, "240000000"), acc0.ID, acc0.KeyPair)
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc2.ID, "230000000"), acc0.ID, acc0.KeyPair)
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc3.ID, "260000000"), acc0.ID, acc0.KeyPair)
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc4.ID, "250000000"), acc0.ID, acc0.KeyPair)
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: 0				, 240000000
			// 2	: 230000000		, 230000000
			// 3	: 0				, 260000000
			// 4	: 0				, 250000000
			// 5	: 0				, 0
			// 6	: q^2*300000000	, 300000000
			// 7	: 430000000		, 215000000
			// 8	: q^1*440000000	, 220000000
			// 6, 0, 3, 1, 4, 5
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 8, 6, 0, 3, 1, 4
			pendingList, _ = json.Marshal([]string{acc8.KeyPair.ID, acc6.KeyPair.ID, acc0.KeyPair.ID, acc3.KeyPair.ID, acc1.KeyPair.ID, acc4.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: 0				, 240000000
			// 2	: 460000000		, 230000000
			// 3	: 0				, 260000000
			// 4	: 0				, 250000000
			// 5	: 0				, 0
			// 6	: q^3*300000000	, 300000000
			// 7	: q^1*645000000	, 215000000
			// 8	: q^2*440000000	, 220000000
			// 8, 6, 0, 3, 1, 4
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 7, 8, 6, 0, 3, 1
			pendingList, _ = json.Marshal([]string{acc7.KeyPair.ID, acc8.KeyPair.ID, acc6.KeyPair.ID, acc0.KeyPair.ID, acc3.KeyPair.ID, acc1.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: 0				, 240000000
			// 2	: q^1*690000000	, 230000000
			// 3	: 0				, 260000000
			// 4	: 250000000		, 250000000
			// 5	: 0				, 0
			// 6	: q^4*300000000	, 300000000
			// 7	: q^2*645000000	, 215000000
			// 8	: q^3*440000000	, 220000000
			// 7, 8, 6, 0, 3, 1
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 2, 7, 8, 6, 0, 3
			pendingList, _ = json.Marshal([]string{acc2.KeyPair.ID, acc7.KeyPair.ID, acc8.KeyPair.ID, acc6.KeyPair.ID, acc0.KeyPair.ID, acc3.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: 240000000		, 240000000
			// 2	: q^2*690000000	, 230000000
			// 3	: 0				, 260000000
			// 4	: q^1*500000000	, 250000000
			// 5	: 0				, 0
			// 6	: q^5*300000000	, 300000000
			// 7	: q^3*645000000	, 215000000
			// 8	: q^4*440000000	, 220000000
			// 2, 7, 8, 6, 0, 3
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 2, 7, 4, 8, 6, 0
			pendingList, _ = json.Marshal([]string{acc2.KeyPair.ID, acc7.KeyPair.ID, acc4.KeyPair.ID, acc8.KeyPair.ID, acc6.KeyPair.ID, acc0.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: q^1*480000000	, 240000000
			// 2	: q^3*690000000	, 230000000
			// 3	: 260000000		, 260000000
			// 4	: q^2*500000000	, 250000000
			// 5	: 0				, 0
			// 6	: q^6*300000000	, 300000000
			// 7	: q^4*645000000	, 215000000
			// 8	: q^5*440000000	, 220000000
			// 2, 7, 4, 8, 6, 0
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 2, 1, 7, 4, 8, 6
			pendingList, _ = json.Marshal([]string{acc2.KeyPair.ID, acc1.KeyPair.ID, acc7.KeyPair.ID, acc4.KeyPair.ID, acc8.KeyPair.ID, acc6.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: q^2*480000000	, 240000000
			// 2	: q^4*690000000	, 230000000
			// 3	: q^1*520000000	, 260000000
			// 4	: q^3*500000000	, 250000000
			// 5	: 0				, 0
			// 6	: 0				, 300000000
			// 7	: q^5*645000000	, 215000000
			// 8	: q^6*440000000	, 220000000
			// 2, 1, 7, 4, 8, 6
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 3, 2, 1, 7, 4, 8
			pendingList, _ = json.Marshal([]string{acc3.KeyPair.ID, acc2.KeyPair.ID, acc1.KeyPair.ID, acc7.KeyPair.ID, acc4.KeyPair.ID, acc8.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: q^3*480000000	, 240000000
			// 2	: q^5*690000000	, 230000000
			// 3	: q^2*520000000	, 260000000
			// 4	: q^4*500000000	, 250000000
			// 5	: 0				, 0
			// 6	: q^1*300000000	, 300000000
			// 7	: q^6*645000000	, 215000000
			// 8	: 0				, 220000000
			// 3, 2, 1, 7, 4, 8
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 3, 2, 1, 7, 6, 4
			pendingList, _ = json.Marshal([]string{acc3.KeyPair.ID, acc2.KeyPair.ID, acc1.KeyPair.ID, acc7.KeyPair.ID, acc6.KeyPair.ID, acc4.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 0
			// 1	: q^4*480000000	, 240000000
			// 2	: q^6*690000000	, 230000000
			// 3	: q^3*520000000	, 260000000
			// 4	: q^5*500000000	, 250000000
			// 5	: 0				, 0
			// 6	: q^2*300000000	, 300000000
			// 7	: q^7*645000000	, 215000000
			// 8	: 220000000		, 220000000
			// 3, 2, 1, 7, 6, 4
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 3, 2, 1, 7, 6, 4
			pendingList, _ = json.Marshal([]string{acc3.KeyPair.ID, acc2.KeyPair.ID, acc1.KeyPair.ID, acc7.KeyPair.ID, acc6.KeyPair.ID, acc4.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))
		})
	})
}

func Test_Unregister2(t *testing.T) {
	ilog.Start()
	Convey("test Unregister2", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareFakeBase(t, s)
		prepareToken(t, s, acc0)
		prepareNewProducerVote(t, s, acc0)

		Convey("test new Unregister", func() {
			initProducer(s)
			s.Head.Number = 1
			for _, acc := range testAccounts[6:] {
				r, err := s.Call("vote_producer.iost", "ApplyRegister", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc.ID, acc.KeyPair.ID), acc.ID, acc.KeyPair)
				So(err, ShouldBeNil)
				So(r.Status.Code, ShouldEqual, tx.Success)
				r, err = s.Call("vote_producer.iost", "ApproveRegister", fmt.Sprintf(`["%v"]`, acc.ID), acc0.ID, acc0.KeyPair)
				So(err, ShouldBeNil)
				So(r.Status.Code, ShouldEqual, tx.Success)
				r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, acc.ID), acc.ID, acc.KeyPair)
				So(err, ShouldBeNil)
				So(r.Status.Code, ShouldEqual, tx.Success)
			}
			for idx, acc := range testAccounts {
				s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc.ID, idx*2+2), acc0.ID, acc0.KeyPair)
				So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc.ID)), ShouldEqual, fmt.Sprintf(`["%v",false,-1]`, idx*2+2))
			}

			// do stat
			// q = 0.9
			s.Head.Number = 200
			r, err := s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			// acc	: score			, votes
			// 0	: 0				, 2
			// 1	: 0				, 4
			// 2	: 0				, 6
			// 3	: 0				, 8
			// 4	: 0				, 10
			// 5	: 0				, 12
			// 6	: 14			, 14
			// 7	: 16			, 16
			// 8	: 18			, 18
			// 9	: q^1*20		, 20
			// 0, 3, 1, 4, 5, 2
			currentList, _ := json.Marshal([]string{acc0.KeyPair.ID, acc3.KeyPair.ID, acc1.KeyPair.ID, acc4.KeyPair.ID, acc5.KeyPair.ID, acc2.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 9, 0, 3, 1, 4, 5
			pendingList, _ := json.Marshal([]string{acc9.KeyPair.ID, acc0.KeyPair.ID, acc3.KeyPair.ID, acc1.KeyPair.ID, acc4.KeyPair.ID, acc5.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			r, err = s.Call("vote_producer.iost", "ApplyUnregister", fmt.Sprintf(`["%v"]`, acc9.ID), acc9.ID, acc9.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 2
			// 1	: 0				, 4
			// 2	: 6				, 6
			// 3	: 0				, 8
			// 4	: 0				, 10
			// 5	: 0				, 12
			// 6	: 28			, 14
			// 7	: 32			, 16
			// 8	: q^1*36		, 18
			// 9	: q^2*20		, 20
			// 9, 0, 3, 1, 4, 5
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 8, 9, 0, 3, 1, 4
			pendingList, _ = json.Marshal([]string{acc8.KeyPair.ID, acc9.KeyPair.ID, acc0.KeyPair.ID, acc3.KeyPair.ID, acc1.KeyPair.ID, acc4.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			r, err = s.Call("vote_producer.iost", "ApproveUnregister", fmt.Sprintf(`["%v"]`, acc9.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 2
			// 1	: 0				, 4
			// 2	: 12			, 6
			// 3	: 0				, 8
			// 4	: 0				, 10
			// 5	: 12			, 12
			// 6	: 42			, 14
			// 7	: q^1*48		, 16
			// 8	: q^2*36		, 18
			// 9 X	: 0				, 20
			// 8, 9, 0, 3, 1, 4
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 7, 8, 0, 3, 1, 4
			pendingList, _ = json.Marshal([]string{acc7.KeyPair.ID, acc8.KeyPair.ID, acc0.KeyPair.ID, acc3.KeyPair.ID, acc1.KeyPair.ID, acc4.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// unregister 8
			r, err = s.Call("vote_producer.iost", "ApplyUnregister", fmt.Sprintf(`["%v"]`, acc8.ID), acc8.ID, acc8.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "ApproveUnregister", fmt.Sprintf(`["%v"]`, acc8.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)

			// unregister 3
			r, err = s.Call("vote_producer.iost", "ApplyUnregister", fmt.Sprintf(`["%v"]`, acc3.ID), acc3.ID, acc3.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "ApproveUnregister", fmt.Sprintf(`["%v"]`, acc3.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0	: 0				, 2
			// 1	: 0				, 4
			// 2	: 18			, 6
			// 3 X	: 0				, 8
			// 4	: 0				, 10
			// 5	: q^1*24		, 12
			// 6	: q^2*56		, 14
			// 7	: q^3*48		, 16
			// 8 X	: 0				, 18
			// 9 X	: 0				, 20
			// 7, 8, 0, 3, 1, 4
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 6, 7, 5, 0, 1, 4
			pendingList, _ = json.Marshal([]string{acc6.KeyPair.ID, acc7.KeyPair.ID, acc5.KeyPair.ID, acc0.KeyPair.ID, acc1.KeyPair.ID, acc4.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			// force unregister all left except for 2 (or acc2.ID)
			for _, acc := range []*TestAccount{acc0, acc1, acc4, acc5, acc6, acc7} {
				r, err = s.Call("vote_producer.iost", "Unregister", fmt.Sprintf(`["%v"]`, acc.ID), acc0.ID, acc0.KeyPair)
				So(err, ShouldBeNil)
				So(r.Status.Code, ShouldEqual, tx.Success)
			}

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0 W	: 0				, 2
			// 1 W	: 0				, 4
			// 2	: q^1*24		, 6
			// 3 X	: 0				, 8
			// 4 X	: 0				, 10
			// 5 W	: q^2*24		, 12
			// 6 W	: q^3*56		, 14
			// 7 W	: q^4*48		, 16
			// 8 X	: 0				, 18
			// 9 X	: 0				, 20
			// 6, 7, 5, 0, 1, 4
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 2, 6, 7, 5, 0, 1
			pendingList, _ = json.Marshal([]string{acc2.KeyPair.ID, acc6.KeyPair.ID, acc7.KeyPair.ID, acc5.KeyPair.ID, acc0.KeyPair.ID, acc1.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

			for _, acc := range []*TestAccount{acc3, acc4, acc8, acc9} {
				r, err := s.Call("vote_producer.iost", "ApplyRegister", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc.ID, acc.KeyPair.ID), acc.ID, acc.KeyPair)
				So(err, ShouldBeNil)
				So(r.Status.Code, ShouldEqual, tx.Success)
				r, err = s.Call("vote_producer.iost", "ApproveRegister", fmt.Sprintf(`["%v"]`, acc.ID), acc0.ID, acc0.KeyPair)
				So(err, ShouldBeNil)
				So(r.Status.Code, ShouldEqual, tx.Success)
				r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, acc.ID), acc.ID, acc.KeyPair)
				So(err, ShouldBeNil)
				So(r.Status.Code, ShouldEqual, tx.Success)
			}

			// do stat
			s.Head.Number += 200
			s.Call("base.iost", "Stat", `[]`, acc0.ID, acc0.KeyPair)
			// acc	: score			, votes
			// 0 X	: 0				, 2
			// 1 X	: 0				, 4
			// 2	: q^2*24		, 6
			// 3	: q^1*8			, 8
			// 4	: q^1*10		, 10
			// 5 X	: 0				, 0
			// 6 W 	: q^4*56		, 14
			// 7 X	: 0				, 0
			// 8	: q^1*18		, 18
			// 9	: q^1*20		, 20
			// 2, 6, 7, 5, 0, 1
			currentList = pendingList
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
			// 2, 9, 8, 4, 3, 6
			pendingList, _ = json.Marshal([]string{acc2.KeyPair.ID, acc9.KeyPair.ID, acc8.KeyPair.ID, acc4.KeyPair.ID, acc3.KeyPair.ID, acc6.KeyPair.ID})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))
		})
	})
}
