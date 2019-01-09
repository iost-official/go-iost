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
		r, err := s.Call("vote_producer.iost", "initProducer", fmt.Sprintf(`["%v", "%v"]`, acc.ID, acc.KeyPair.ReadablePubkey()), acc0.ID, acc0.KeyPair)
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
	r, err = s.Call("vote_producer.iost", "initAdmin", fmt.Sprintf(`["%v"]`, acc1.ID), acc1.ID, acc1.KeyPair)
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
	Convey("test initProducer", t, func() {
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
			list, _ := json.Marshal([]string{acc0.KeyPair.ReadablePubkey(), acc3.KeyPair.ReadablePubkey(), acc1.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey(), acc5.KeyPair.ReadablePubkey(), acc2.KeyPair.ReadablePubkey()})
			So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(list))

			So(s.Visitor.MKeys("vote.iost-v_1"), ShouldResemble, []string{"0", "1", "2", "3", "4", "5"})
		})
	})
}

func Test_RegisterProducer(t *testing.T) {
	ilog.Stop()
	Convey("test registerProducer", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareToken(t, s, acc0)
		prepareProducerVote(t, s, acc0)
		initProducer(s)

		Convey("test register/unregister", func() {
			r, err := s.Call("vote_producer.iost", "registerProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc6.ID, acc6.KeyPair.ReadablePubkey()), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote_producer.iost-producerTable", acc6.ID)), ShouldEqual, fmt.Sprintf(`{"pubkey":"%s","loc":"loc","url":"url","netId":"netId","online":false,"registerFee":"200000000"}`, acc6.KeyPair.ReadablePubkey()))
			So(s.Visitor.TokenBalance("iost", acc6.ID), ShouldEqual, int64(1800000000*1e8))
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "6")), ShouldEqual, `"0"`)

			r, err = s.Call("vote_producer.iost", "unregisterProducer", fmt.Sprintf(`["%v"]`, acc6.ID), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(s.Visitor.MHas("vote_producer.iost-producerTable", acc6.ID), ShouldEqual, false)
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "6")), ShouldEqual, `"0"`)
		})
	})
}

func Test_LogInOut(t *testing.T) {
	ilog.Stop()
	Convey("test registerProducer", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareToken(t, s, acc0)
		prepareProducerVote(t, s, acc0)
		initProducer(s)

		Convey("test login/logout", func() {
			r, err := s.Call("vote_producer.iost", "registerProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc6.ID, acc6.KeyPair.ReadablePubkey()), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			r, err = s.Call("vote_producer.iost", "logInProducer", fmt.Sprintf(`["%v"]`, acc6.ID), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote_producer.iost-producerTable", acc6.ID)), ShouldEqual, fmt.Sprintf(`{"pubkey":"%s","loc":"loc","url":"url","netId":"netId","online":true,"registerFee":"200000000"}`, acc6.KeyPair.ReadablePubkey()))

			r, err = s.Call("vote_producer.iost", "logOutProducer", fmt.Sprintf(`["%v"]`, acc6.ID), acc6.ID, acc6.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote_producer.iost-producerTable", acc6.ID)), ShouldEqual, fmt.Sprintf(`{"pubkey":"%s","loc":"loc","url":"url","netId":"netId","online":false,"registerFee":"200000000"}`, acc6.KeyPair.ReadablePubkey()))

			r, _ = s.Call("vote_producer.iost", "logOutProducer", fmt.Sprintf(`["%v"]`, acc0.ID), acc0.ID, acc0.KeyPair)
			So(r.Status.Message, ShouldNotEqual, "")
			So(r.Status.Code, ShouldEqual, 4)
			So(r.Status.Message, ShouldContainSubstring, "producer in pending list or in current list, can't logout")
		})
	})
}

func Test_Vote1(t *testing.T) {
	t.Skip()
	ilog.Stop()
	Convey("test vote", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareFakeBase(t, s)
		prepareToken(t, s, acc0)
		prepareProducerVote(t, s, acc0)
		initProducer(s)

		s.Head.Number = 1
		r, err := s.Call("vote_producer.iost", "registerProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc6.ID, acc6.KeyPair.ReadablePubkey()), acc6.ID, acc6.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(r.Status.Code, ShouldEqual, tx.Success)
		r, err = s.Call("vote_producer.iost", "registerProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc7.ID, acc7.KeyPair.ReadablePubkey()), acc7.ID, acc7.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(r.Status.Code, ShouldEqual, tx.Success)
		r, err = s.Call("vote_producer.iost", "registerProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc8.ID, acc8.KeyPair.ReadablePubkey()), acc8.ID, acc8.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(r.Status.Code, ShouldEqual, tx.Success)
		r, err = s.Call("vote_producer.iost", "logInProducer", fmt.Sprintf(`["%v"]`, acc6.ID), acc6.ID, acc6.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(r.Status.Code, ShouldEqual, tx.Success)
		r, err = s.Call("vote_producer.iost", "logInProducer", fmt.Sprintf(`["%v"]`, acc7.ID), acc7.ID, acc7.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(r.Status.Code, ShouldEqual, tx.Success)
		r, err = s.Call("vote_producer.iost", "logInProducer", fmt.Sprintf(`["%v"]`, acc8.ID), acc8.ID, acc8.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(r.Status.Code, ShouldEqual, tx.Success)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "6")), ShouldEqual, `"0"`)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "7")), ShouldEqual, `"0"`)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "8")), ShouldEqual, `"0"`)

		s.Call("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc6.ID, "100000000"), acc0.ID, acc0.KeyPair)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "6")), ShouldEqual, `"100000000"`)

		s.Call("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc7.ID, acc6.ID, "100000000"), acc7.ID, acc7.KeyPair)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "6")), ShouldEqual, `"200000000"`)
		info := database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
		So(info, ShouldContainSubstring, `"0":0,"1":0,"2":0,"3":0,"4":0,"5":0,"6":1,"7":0,"8":0`)

		s.Call("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc8.ID, acc6.ID, "100000000"), acc8.ID, acc8.KeyPair)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", "6")), ShouldEqual, `"300000000"`)
		info = database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", "1"))
		So(info, ShouldContainSubstring, `"0":0,"1":0,"2":0,"3":0,"4":0,"5":0,"6":1,"7":0,"8":0`)

		s.Call("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc7.ID, "215000000"), acc0.ID, acc0.KeyPair)
		s.Call("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc8.ID, "220000000"), acc0.ID, acc0.KeyPair)

		r, err = s.Call("vote_producer.iost", "getProducer", fmt.Sprintf(`["%v"]`, acc6.ID), acc6.ID, acc6.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)
		So(r.Returns[0], ShouldEqual, fmt.Sprintf(`["{\"pubkey\":\"%s\",\"loc\":\"loc\",\"url\":\"url\",\"netId\":\"netId\",\"online\":true,\"registerFee\":\"200000000\",\"voteInfo\":{\"votes\":\"300000000\",\"deleted\":0,\"clearTime\":-1}}"]`, acc6.KeyPair.ReadablePubkey()))

		r, err = s.Call("vote_producer.iost", "getVote", fmt.Sprintf(`["%v"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)
		So(r.Returns[0], ShouldEqual, `["[{\"option\":\"user_6\",\"votes\":\"100000000\",\"voteTime\":1,\"clearedVotes\":\"0\"},{\"option\":\"user_7\",\"votes\":\"215000000\",\"voteTime\":1,\"clearedVotes\":\"0\"},{\"option\":\"user_8\",\"votes\":\"220000000\",\"voteTime\":1,\"clearedVotes\":\"0\"}]"]`)

		// do stat
		s.Head.Number = 2000
		r, err = s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
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
		// 8, 0, 1, 4, 5, 2
		currentList, _ := json.Marshal([]string{acc8.KeyPair.ReadablePubkey(), acc0.KeyPair.ReadablePubkey(), acc1.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey(), acc5.KeyPair.ReadablePubkey(), acc2.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
		// 6, 0, 3, 1, 4, 5
		pendingList, _ := json.Marshal([]string{acc6.KeyPair.ReadablePubkey(), acc0.KeyPair.ReadablePubkey(), acc3.KeyPair.ReadablePubkey(), acc1.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey(), acc5.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

		// do stat
		s.Call("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc1.ID, "240000000"), acc0.ID, acc0.KeyPair)
		s.Call("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc2.ID, "230000000"), acc0.ID, acc0.KeyPair)
		s.Call("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc3.ID, "260000000"), acc0.ID, acc0.KeyPair)
		s.Call("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc4.ID, "250000000"), acc0.ID, acc0.KeyPair)
		s.Head.Number += 2000
		s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
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
		pendingList, _ = json.Marshal([]string{acc6.KeyPair.ReadablePubkey(), acc2.KeyPair.ReadablePubkey(), acc0.KeyPair.ReadablePubkey(), acc3.KeyPair.ReadablePubkey(), acc1.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

		// do stat
		s.Head.Number += 2000
		s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
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
		pendingList, _ = json.Marshal([]string{acc6.KeyPair.ReadablePubkey(), acc8.KeyPair.ReadablePubkey(), acc2.KeyPair.ReadablePubkey(), acc0.KeyPair.ReadablePubkey(), acc3.KeyPair.ReadablePubkey(), acc1.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

		// do stat
		s.Head.Number += 2000
		s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
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
		pendingList, _ = json.Marshal([]string{acc6.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey(), acc8.KeyPair.ReadablePubkey(), acc2.KeyPair.ReadablePubkey(), acc0.KeyPair.ReadablePubkey(), acc3.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

		// do stat
		s.Head.Number += 2000
		s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
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
		pendingList, _ = json.Marshal([]string{acc6.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey(), acc1.KeyPair.ReadablePubkey(), acc8.KeyPair.ReadablePubkey(), acc2.KeyPair.ReadablePubkey(), acc0.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

		// do stat
		s.Head.Number += 2000
		s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
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
		pendingList, _ = json.Marshal([]string{acc6.KeyPair.ReadablePubkey(), acc3.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey(), acc1.KeyPair.ReadablePubkey(), acc8.KeyPair.ReadablePubkey(), acc2.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

		// do stat
		s.Head.Number += 2000
		s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
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
		pendingList, _ = json.Marshal([]string{acc6.KeyPair.ReadablePubkey(), acc3.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey(), acc7.KeyPair.ReadablePubkey(), acc1.KeyPair.ReadablePubkey(), acc8.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

		// do stat
		s.Head.Number += 2000
		s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
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
		pendingList, _ = json.Marshal([]string{acc6.KeyPair.ReadablePubkey(), acc3.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey(), acc7.KeyPair.ReadablePubkey(), acc1.KeyPair.ReadablePubkey(), acc8.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))

		// do stat
		s.Head.Number += 2000
		s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
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
		pendingList, _ = json.Marshal([]string{acc6.KeyPair.ReadablePubkey(), acc3.KeyPair.ReadablePubkey(), acc2.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey(), acc7.KeyPair.ReadablePubkey(), acc1.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))
	})
}

func Test_Unregister2(t *testing.T) {
	ilog.Stop()
	Convey("test Unregister2", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareFakeBase(t, s)
		prepareToken(t, s, acc0)
		prepareNewProducerVote(t, s, acc0)
		initProducer(s)

		s.Head.Number = 1
		for _, acc := range testAccounts[6:] {
			r, err := s.Call("vote_producer.iost", "applyRegister", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId", true]`, acc.ID, acc.KeyPair.ReadablePubkey()), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			r, err = s.Call("vote_producer.iost", "approveRegister", fmt.Sprintf(`["%v"]`, acc.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			r, err = s.Call("vote_producer.iost", "logInProducer", fmt.Sprintf(`["%v"]`, acc.ID), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
		}
		// So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-voteInfo", fmt.Sprintf(`%d`, 1))), ShouldEqual, "")
		for idx, acc := range testAccounts {
			r, err := s.Call("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc.ID, (idx+2)*1e7), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", fmt.Sprintf(`%d`, idx))), ShouldEqual, fmt.Sprintf(`"%d"`, (idx+2)*1e7))
		}

		// do stat
		s.Head.Number = 2000
		r, err := s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		// acc	: score			, votes
		// 0	: 2				, 2
		// 1	: 3				, 3
		// 2	: 4				, 4
		// 3	: 5				, 5
		// 4	: 6 - 0.65		, 6
		// 5	: 7 - 0.65		, 7
		// 6	: 8	- 0.65		, 8
		// 7	: 9 - 0.65		, 9
		// 8	: 10 - 0.65		, 10
		// 9	: 11 - 0.65		, 11
		// 0, 3, 1, 4, 5, 2
		currentList, _ := json.Marshal([]string{acc0.KeyPair.ReadablePubkey(), acc3.KeyPair.ReadablePubkey(), acc1.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey(), acc5.KeyPair.ReadablePubkey(), acc2.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
		// 9, 8, 7, 6, 5, 4
		pendingList, _ := json.Marshal([]string{acc9.KeyPair.ReadablePubkey(), acc8.KeyPair.ReadablePubkey(), acc7.KeyPair.ReadablePubkey(), acc6.KeyPair.ReadablePubkey(), acc5.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))
		scores := `{"user_9":"103500000.00000000","user_8":"93500000.00000000","user_7":"83500000.00000000","user_6":"73500000.00000000","user_5":"63500000.00000000","user_4":"53500000.00000000","user_3":"50000000","user_2":"40000000","user_1":"30000000","user_0":"20000000"}`
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-producerScores")), ShouldEqual, scores)

		r, err = s.Call("vote_producer.iost", "applyUnregister", fmt.Sprintf(`["%v"]`, acc9.ID), acc9.ID, acc9.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		// do stat
		s.Head.Number += 2000
		s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
		// acc	: score			, votes
		// 0	: 4				, 2
		// 1	: 6				, 3
		// 2	: 8				, 4
		// 3	: 10			, 5
		// 4	: 12 - 1.911	, 6
		// 5	: 14 - 1.911	, 7
		// 6	: 16 - 1.911	, 8
		// 7	: 18 - 1.911	, 9
		// 8	: 20 - 1.911	, 10
		// 9	: 22 - 1.911	, 11
		// 9, 8, 7, 6, 5, 4
		currentList = pendingList
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
		// 9, 8, 7, 6, 5, 4
		pendingList, _ = json.Marshal([]string{acc9.KeyPair.ReadablePubkey(), acc8.KeyPair.ReadablePubkey(), acc7.KeyPair.ReadablePubkey(), acc6.KeyPair.ReadablePubkey(), acc5.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))
		scores = `{"user_9":"200890000.00000000","user_8":"180890000.00000000","user_7":"160890000.00000000","user_6":"140890000.00000000","user_5":"120890000.00000000","user_4":"100890000.00000000","user_3":"100000000","user_2":"80000000","user_1":"60000000","user_0":"40000000"}`
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-producerScores")), ShouldEqual, scores)

		r, err = s.Call("vote_producer.iost", "approveUnregister", fmt.Sprintf(`["%v"]`, acc9.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)
		r, err = s.Call("vote_producer.iost", "unregister", fmt.Sprintf(`["%v"]`, acc9.ID), acc9.ID, acc9.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		// do stat
		s.Head.Number += 2000
		s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
		// acc	: score				, votes
		// 0	: 6					, 2
		// 1	: 9					, 3
		// 2	: 12				, 4
		// 3	: 15 - 1.69383333	, 5
		// 4	: 18 - 3.60483333	, 6
		// 5	: 21 - 3.60483333	, 7
		// 6	: 24 - 3.60483333	, 8
		// 7	: 27 - 3.60483333	, 9
		// 8	: 30 - 3.60483333	, 10
		// 9 X	: X					, 11
		// 9, 8, 7, 6, 5, 4
		currentList = pendingList
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
		// 8, 7, 6, 5, 4, 3
		pendingList, _ = json.Marshal([]string{acc8.KeyPair.ReadablePubkey(), acc7.KeyPair.ReadablePubkey(), acc6.KeyPair.ReadablePubkey(), acc5.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey(), acc3.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))
		scores = `{"user_8":"263951666.66666666","user_7":"233951666.66666666","user_6":"203951666.66666666","user_5":"173951666.66666666","user_4":"143951666.66666666","user_3":"133061666.66666666","user_2":"120000000","user_1":"90000000","user_0":"60000000"}`
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-producerScores")), ShouldEqual, scores)

		// unregister 8
		r, err = s.Call("vote_producer.iost", "applyUnregister", fmt.Sprintf(`["%v"]`, acc8.ID), acc8.ID, acc8.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)
		r, err = s.Call("vote_producer.iost", "approveUnregister", fmt.Sprintf(`["%v"]`, acc8.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)
		r, err = s.Call("vote_producer.iost", "unregister", fmt.Sprintf(`["%v"]`, acc8.ID), acc8.ID, acc8.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		// unregister 3
		r, err = s.Call("vote_producer.iost", "applyUnregister", fmt.Sprintf(`["%v"]`, acc3.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)
		r, err = s.Call("vote_producer.iost", "approveUnregister", fmt.Sprintf(`["%v"]`, acc3.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)
		r, err = s.Call("vote_producer.iost", "unregister", fmt.Sprintf(`["%v"]`, acc3.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		// do stat
		s.Head.Number += 2000
		s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
		// acc	: score			, votes
		// 0	: 8					, 2
		// 1	: 12 - 2.02258095	, 3
		// 2	: 16 - 2.02258095	, 4
		// 3 X	: X					, 5
		// 4	: 24 - 5.62741428	, 6
		// 5	: 28 - 5.62741428	, 7
		// 6	: 32 - 5.62741428	, 8
		// 7	: 36 - 5.62741428	, 9
		// 8 X	: X					, 10
		// 9 X	: X					, 11
		// 8, 7, 6, 5, 4, 3
		currentList = pendingList
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
		// 7, 6, 5, 4, 2, 1
		pendingList, _ = json.Marshal([]string{acc7.KeyPair.ReadablePubkey(), acc6.KeyPair.ReadablePubkey(), acc5.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey(), acc2.KeyPair.ReadablePubkey(), acc1.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))
		scores = `{"user_7":"303725857.14285713","user_6":"263725857.14285713","user_5":"223725857.14285713","user_4":"183725857.14285713","user_2":"139774190.47619047","user_1":"99774190.47619047","user_0":"80000000"}`
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-producerScores")), ShouldEqual, scores)

		// force unregister all left except for 2 (or acc2.ID)
		for _, acc := range []*TestAccount{acc0, acc1, acc4, acc5, acc6, acc7} {
			r, err = s.Call("vote_producer.iost", "forceUnregister", fmt.Sprintf(`["%v"]`, acc.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "unregister", fmt.Sprintf(`["%v"]`, acc.ID), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
		}

		// do stat
		s.Head.Number += 2000
		s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
		// acc	: score				, votes
		// 0 X	: X					, 2
		// 1 W	: X					, 3
		// 2	: 20 - 3.82032285	, 4
		// 3 X	: X					, 5
		// 4 W	: X					, 6
		// 5 W	: X					, 7
		// 6 W	: X					, 8
		// 7 W	: X					, 9
		// 8 X	: X					, 10
		// 9 X	: X					, 11
		// 7, 6, 5, 4, 2, 1
		currentList = pendingList
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
		// 2, 1, 6, 4, 5, 7
		pendingList, _ = json.Marshal([]string{acc2.KeyPair.ReadablePubkey(), acc1.KeyPair.ReadablePubkey(), acc6.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey(), acc5.KeyPair.ReadablePubkey(), acc7.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))
		scores = `{"user_2":"161796771.42857142"}`
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-producerScores")), ShouldEqual, scores)

		for _, acc := range []*TestAccount{acc3, acc4, acc8, acc9} {
			r, err := s.Call("vote_producer.iost", "applyRegister", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId", true]`, acc.ID, acc.KeyPair.ReadablePubkey()), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			r, err = s.Call("vote_producer.iost", "approveRegister", fmt.Sprintf(`["%v"]`, acc.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			r, err = s.Call("vote_producer.iost", "logInProducer", fmt.Sprintf(`["%v"]`, acc.ID), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
		}

		// do stat
		s.Head.Number += 2000
		s.Call("base.iost", "stat", `[]`, acc0.ID, acc0.KeyPair)
		// acc	: score				, votes
		// 0 X	: X					, 2
		// 1 W	: X					, 3
		// 2	: 24 - 4.86391639	, 4
		// 3 	: 5 - 1.04359354	, 5
		// 4 	: 6 - 1.04359354	, 6
		// 5 W	: X					, 7
		// 6 W	: X					, 8
		// 7 W	: X					, 9
		// 8 	: 10 - 1.04359354	, 10
		// 9 	: 11 - 1.04359354	, 11
		// 2, 1, 6, 4, 5, 7
		currentList = pendingList
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-currentProducerList")), ShouldEqual, string(currentList))
		// 2, 9, 8, 4, 3, 1
		pendingList, _ = json.Marshal([]string{acc2.KeyPair.ReadablePubkey(), acc9.KeyPair.ReadablePubkey(), acc8.KeyPair.ReadablePubkey(), acc4.KeyPair.ReadablePubkey(), acc3.KeyPair.ReadablePubkey(), acc1.KeyPair.ReadablePubkey()})
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-pendingProducerList")), ShouldEqual, string(pendingList))
		scores = `{"user_2":"191360835.99999999","user_9":"99564064.57142857","user_8":"89564064.57142857","user_4":"49564064.57142857","user_3":"39564064.57142857"}`
		So(database.MustUnmarshal(s.Visitor.Get("vote_producer.iost-producerScores")), ShouldEqual, scores)
	})
}
