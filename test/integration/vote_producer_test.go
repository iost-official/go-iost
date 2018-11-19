package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/ilog"
	. "github.com/iost-official/go-iost/verifier"
	. "github.com/smartystreets/goconvey/convey"
)

func prepareProducerVote(t *testing.T, s *Simulator, kp *account.KeyPair) {
	// deploy vote.iost
	setNonNativeContract(s, "vote.iost", "vote_common.js", ContractPath)
	s.Call("vote.iost", "init", `[]`, kp.ID, kp)

	// deploy vote_producer.iost
	setNonNativeContract(s, "vote_producer.iost", "vote.js", ContractPath)

	s.Call("token.iost", "issue", fmt.Sprintf(`["%v", "%v", "%v"]`, "iost", "vote_producer.iost", "1000"), kp.ID, kp)

	r, err := s.Call("vote_producer.iost", "init", `[]`, kp.ID, kp)
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
		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(s)
		prepareToken(t, s, kp)
		prepareProducerVote(t, s, kp)

		So(s.Visitor.Get("vote.iost-current_id"), ShouldEqual, `s"1"`)
		So(s.Visitor.Get("vote_producer.iost-voteId"), ShouldEqual, `s"1"`)
		Convey("test init producer", func() {
			for i := 0; i < 12; i += 2 {
				s.Call("vote_producer.iost", "InitProducer", fmt.Sprintf(`["%v", "%v"]`, testID[i], testID[i]), kp.ID, kp)
			}
			So(s.Visitor.Get("vote_producer.iost-pendingProducerList"), ShouldEqual,
				`s["IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn","IOST54ETA3q5eC8jAoEpfRAToiuc6Fjs5oqEahzghWkmEYs9S9CMKd"`+
					`,"IOST558jUpQvBD7F3WTKpnDAWg6HwKrfFiZ7AqhPFf4QSrmjdmBGeY","IOST7GmPn8xC1RESMRS6a62RmBcCdwKbKvk2ZpxZpcXdUPoJdapnnh"`+
					`,"IOST7ZGQL4k85v4wAxWngmow7JcX4QFQ4mtLNjgvRrEnEuCkGSBEHN","IOST7ZNDWeh8pHytAZdpgvp7vMpjZSSe5mUUKxDm6AXPsbdgDMAYhs"]`)

			So(s.Visitor.MKeys("vote.iost-v-1"), ShouldResemble, []string{testID[0], testID[2], testID[4], testID[6], testID[8], testID[10]})
		})
	})
}

func Test_RegisterProducer(t *testing.T) {
	ilog.Stop()
	Convey("test RegisterProducer", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0
		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(s)
		prepareToken(t, s, kp)
		prepareProducerVote(t, s, kp)
		for i := 0; i < 12; i += 2 {
			s.Call("vote_producer.iost", "InitProducer", fmt.Sprintf(`["%v", "%v"]`, testID[i], testID[i]), kp.ID, kp)
		}

		Convey("test register/unregister", func() {
			kp6, _ := account.NewKeyPair(common.Base58Decode(testID[13]), crypto.Secp256k1)
			s.Call("vote_producer.iost", "RegisterProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, kp6.ID, testID[12]), kp6.ID, kp6)
			So(s.Visitor.MGet("vote_producer.iost-producerTable", testID[12]), ShouldEqual, `s{"pubkey":"IOST59uMX3Y4ab5dcq8p1wMXodANccJcj2efbcDThtkw6egvcni5L9","loc":"loc","url":"url","netId":"netId","online":false,"registerFee":"200000000","score":"0"}`)
			So(s.Visitor.TokenBalance("iost", kp6.ID), ShouldEqual, int64(1800000000*1e8))
			So(s.Visitor.MGet("vote.iost-v-1", kp6.ID), ShouldEqual, `s["0",false,-1]`)

			s.Call("vote_producer.iost", "UnregisterProducer", fmt.Sprintf(`["%v"]`, kp6.ID), kp6.ID, kp6)
			So(s.Visitor.MHas("vote_producer.iost-producerTable", kp6.ID), ShouldEqual, false)
			So(s.Visitor.MGet("vote.iost-v-1", kp6.ID), ShouldEqual, `s["0",true,-1]`)
		})
	})
}

func Test_LogInOut(t *testing.T) {
	ilog.Stop()
	Convey("test RegisterProducer", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0
		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(s)
		prepareToken(t, s, kp)
		prepareProducerVote(t, s, kp)
		for i := 0; i < 12; i += 2 {
			s.Call("vote_producer.iost", "InitProducer", fmt.Sprintf(`["%v", "%v"]`, testID[i], testID[i]), kp.ID, kp)
		}

		Convey("test login/logout", func() {
			kp6, _ := account.NewKeyPair(common.Base58Decode(testID[13]), crypto.Secp256k1)
			s.Call("vote_producer.iost", "RegisterProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, kp6.ID, testID[12]), kp6.ID, kp6)

			s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, kp6.ID), kp6.ID, kp6)
			So(s.Visitor.MGet("vote_producer.iost-producerTable", testID[12]), ShouldEqual, `s{"pubkey":"IOST59uMX3Y4ab5dcq8p1wMXodANccJcj2efbcDThtkw6egvcni5L9","loc":"loc","url":"url","netId":"netId","online":true,"registerFee":"200000000","score":"0"}`)

			s.Call("vote_producer.iost", "LogOutProducer", fmt.Sprintf(`["%v"]`, kp6.ID), kp6.ID, kp6)
			So(s.Visitor.MGet("vote_producer.iost-producerTable", testID[12]), ShouldEqual, `s{"pubkey":"IOST59uMX3Y4ab5dcq8p1wMXodANccJcj2efbcDThtkw6egvcni5L9","loc":"loc","url":"url","netId":"netId","online":false,"registerFee":"200000000","score":"0"}`)

			r, _ := s.Call("vote_producer.iost", "LogOutProducer", fmt.Sprintf(`["%v"]`, kp.ID), kp.ID, kp)
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
		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(s)
		prepareToken(t, s, kp)
		prepareProducerVote(t, s, kp)
		for i := 0; i < 12; i += 2 {
			r, err := s.Call("vote_producer.iost", "InitProducer", fmt.Sprintf(`["%v", "%v"]`, testID[i], testID[i]), kp.ID, kp)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
		}

		Convey("test vote/unvote", func() {
			s.Head.Number = 1
			kp6, _ := account.NewKeyPair(common.Base58Decode(testID[13]), crypto.Secp256k1)
			kp7, _ := account.NewKeyPair(common.Base58Decode(testID[15]), crypto.Secp256k1)
			kp8, _ := account.NewKeyPair(common.Base58Decode(testID[17]), crypto.Secp256k1)
			r, err := s.Call("vote_producer.iost", "RegisterProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, kp6.ID, testID[12]), kp6.ID, kp6)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "RegisterProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, kp7.ID, testID[14]), kp7.ID, kp7)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "RegisterProducer", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, kp8.ID, testID[16]), kp8.ID, kp8)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, kp6.ID), kp6.ID, kp6)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, kp7.ID), kp7.ID, kp7)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, kp8.ID), kp8.ID, kp8)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.Visitor.MGet("vote.iost-v-1", kp6.ID), ShouldEqual, `s["0",false,-1]`)
			So(s.Visitor.MGet("vote.iost-v-1", kp7.ID), ShouldEqual, `s["0",false,-1]`)
			So(s.Visitor.MGet("vote.iost-v-1", kp8.ID), ShouldEqual, `s["0",false,-1]`)

			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, kp.ID, kp6.ID, "100000000"), kp.ID, kp)
			So(s.Visitor.MGet("vote.iost-v-1", testID[12]), ShouldEqual, `s["100000000",false,-1]`)

			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, kp7.ID, kp6.ID, "100000000"), kp7.ID, kp7)
			So(s.Visitor.MGet("vote.iost-v-1", testID[12]), ShouldEqual, `s["200000000",false,-1]`)
			So(s.Visitor.MHas("vote.iost-p-1", testID[12]), ShouldEqual, false)

			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, kp8.ID, kp6.ID, "100000000"), kp8.ID, kp8)
			So(s.Visitor.MGet("vote.iost-v-1", testID[12]), ShouldEqual, `s["300000000",false,-1]`)
			So(s.Visitor.MGet("vote.iost-p-1", testID[12]), ShouldEqual, `s"300000000"`)

			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, kp.ID, kp7.ID, "215000000"), kp.ID, kp)
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, kp.ID, kp8.ID, "220000000"), kp.ID, kp)

			r, err = s.Call("vote_producer.iost", "GetProducer", fmt.Sprintf(`["%v"]`, kp6.ID), kp6.ID, kp6)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(r.Returns[0], ShouldEqual, `["{\"pubkey\":\"IOST59uMX3Y4ab5dcq8p1wMXodANccJcj2efbcDThtkw6egvcni5L9\",\"loc\":\"loc\",\"url\":\"url\",\"netId\":\"netId\",\"online\":true,\"registerFee\":\"200000000\",\"score\":\"0\",\"voteInfo\":{\"votes\":\"300000000\",\"deleted\":false,\"clearTime\":-1}}"]`)

			r, err = s.Call("vote_producer.iost", "GetVote", fmt.Sprintf(`["%v"]`, kp.ID), kp.ID, kp)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(r.Returns[0], ShouldEqual, `["[{\"option\":\"IOST59uMX3Y4ab5dcq8p1wMXodANccJcj2efbcDThtkw6egvcni5L9\",\"votes\":\"100000000\",\"voteTime\":1,\"clearedVotes\":\"0\"},{\"option\":\"IOST8mFxe4kq9XciDtURFZJ8E76B8UssBgRVFA5gZN9HF5kLUVZ1BB\",\"votes\":\"215000000\",\"voteTime\":1,\"clearedVotes\":\"0\"},{\"option\":\"IOST7uqa5UQPVT9ongTv6KmqDYKdVYSx4DV2reui4nuC5mm5vBt3D9\",\"votes\":\"220000000\",\"voteTime\":1,\"clearedVotes\":\"0\"}]"]`)

			// do stat
			s.Head.Number = 200
			s.Call("vote_producer.iost", "Stat", `[]`, kp.ID, kp)
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
			currentList, _ := json.Marshal([]string{testID[0], testID[6], testID[2], testID[8], testID[10], testID[4]})
			So(s.Visitor.Get("vote_producer.iost-currentProducerList"), ShouldEqual, "s"+string(currentList))
			// 6, 0, 3, 1, 4, 5
			pendingList, _ := json.Marshal([]string{testID[12], testID[0], testID[6], testID[2], testID[8], testID[10]})
			So(s.Visitor.Get("vote_producer.iost-pendingProducerList"), ShouldEqual, "s"+string(pendingList))

			// do stat
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, testID[0], testID[2], "240000000"), kp.ID, kp)
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, testID[0], testID[4], "230000000"), kp.ID, kp)
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, testID[0], testID[6], "260000000"), kp.ID, kp)
			s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, testID[0], testID[8], "250000000"), kp.ID, kp)
			s.Head.Number += 200
			s.Call("vote_producer.iost", "Stat", `[]`, kp.ID, kp)
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
			So(s.Visitor.Get("vote_producer.iost-currentProducerList"), ShouldEqual, "s"+string(currentList))
			// 6, 2, 0, 3, 1, 4
			pendingList, _ = json.Marshal([]string{testID[12], testID[4], testID[0], testID[6], testID[2], testID[8]})
			So(s.Visitor.Get("vote_producer.iost-pendingProducerList"), ShouldEqual, "s"+string(pendingList))

			// do stat
			s.Call("vote_producer.iost", "Unvote", fmt.Sprintf(`["%v", "%v", "%v"]`, testID[16], testID[12], "60000000"), kp8.ID, kp8)
			s.Head.Number += 200
			s.Call("vote_producer.iost", "Stat", `[]`, kp.ID, kp)
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
			So(s.Visitor.Get("vote_producer.iost-currentProducerList"), ShouldEqual, "s"+string(currentList))
			// 6, 8, 2, 0, 3, 1
			pendingList, _ = json.Marshal([]string{testID[12], testID[16], testID[4], testID[0], testID[6], testID[2]})
			So(s.Visitor.Get("vote_producer.iost-pendingProducerList"), ShouldEqual, "s"+string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("vote_producer.iost", "Stat", `[]`, kp.ID, kp)
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
			So(s.Visitor.Get("vote_producer.iost-currentProducerList"), ShouldEqual, "s"+string(currentList))
			// 6, 4, 8, 2, 0, 3
			pendingList, _ = json.Marshal([]string{testID[12], testID[8], testID[16], testID[4], testID[0], testID[6]})
			So(s.Visitor.Get("vote_producer.iost-pendingProducerList"), ShouldEqual, "s"+string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("vote_producer.iost", "Stat", `[]`, kp.ID, kp)
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
			So(s.Visitor.Get("vote_producer.iost-currentProducerList"), ShouldEqual, "s"+string(currentList))
			// 6, 4, 1, 8, 2, 0
			pendingList, _ = json.Marshal([]string{testID[12], testID[8], testID[2], testID[16], testID[4], testID[0]})
			So(s.Visitor.Get("vote_producer.iost-pendingProducerList"), ShouldEqual, "s"+string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("vote_producer.iost", "Stat", `[]`, kp.ID, kp)
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
			So(s.Visitor.Get("vote_producer.iost-currentProducerList"), ShouldEqual, "s"+string(currentList))
			// 6, 3, 4, 1, 8, 2
			pendingList, _ = json.Marshal([]string{testID[12], testID[6], testID[8], testID[2], testID[16], testID[4]})
			So(s.Visitor.Get("vote_producer.iost-pendingProducerList"), ShouldEqual, "s"+string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("vote_producer.iost", "Stat", `[]`, kp.ID, kp)
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
			So(s.Visitor.Get("vote_producer.iost-currentProducerList"), ShouldEqual, "s"+string(currentList))
			// 6, 3, 4, 7, 1, 8
			pendingList, _ = json.Marshal([]string{testID[12], testID[6], testID[8], testID[14], testID[2], testID[16]})
			So(s.Visitor.Get("vote_producer.iost-pendingProducerList"), ShouldEqual, "s"+string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("vote_producer.iost", "Stat", `[]`, kp.ID, kp)
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
			So(s.Visitor.Get("vote_producer.iost-currentProducerList"), ShouldEqual, "s"+string(currentList))
			// 6, 3, 4, 7, 1, 8
			pendingList, _ = json.Marshal([]string{testID[12], testID[6], testID[8], testID[14], testID[2], testID[16]})
			So(s.Visitor.Get("vote_producer.iost-pendingProducerList"), ShouldEqual, "s"+string(pendingList))

			// do stat
			s.Head.Number += 200
			s.Call("vote_producer.iost", "Stat", `[]`, kp.ID, kp)
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
			So(s.Visitor.Get("vote_producer.iost-currentProducerList"), ShouldEqual, "s"+string(currentList))
			// 6, 3, 2, 4, 7, 1
			pendingList, _ = json.Marshal([]string{testID[12], testID[6], testID[4], testID[8], testID[14], testID[2]})
			So(s.Visitor.Get("vote_producer.iost-pendingProducerList"), ShouldEqual, "s"+string(pendingList))
		})
	})
}
