package synchronizer

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/golang/mock/gomock"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/genesis"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/mocks"
	"github.com/iost-official/go-iost/db/mocks"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
	"github.com/iost-official/go-iost/p2p/mocks"
	"github.com/iost-official/go-iost/vm/database"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDownloadController(t *testing.T) {
	t.Skip()
	Convey("Test DownloadController", t, func() {
		dHash := make(chan string, 10)
		dPID := make(chan p2p.PeerID, 10)
		fpFunc := func(hash string, p interface{}) bool {
			return false
		}
		mFunc := func(hash string, p interface{}, peerID interface{}) (bool, bool) {
			dHash <- hash
			dPID <- peerID.(p2p.PeerID)
			return true, false
		}
		dc, err := NewDownloadController(fpFunc, mFunc)
		dc.Start()

		So(err, ShouldBeNil)
		Convey("Check OnRecvHash", func() {
			dc.CreateMission("111", 10, "aaa")
			dc.CreateMission("222", 10, "aaa")
			var hash string
			var PID p2p.PeerID
			hashes := make(map[string]bool, 0)
			pids := make(map[p2p.PeerID]bool, 0)
			hash = <-dHash
			PID = <-dPID
			hashes[hash] = true
			pids[PID] = true
			hash = <-dHash
			PID = <-dPID
			hashes[hash] = true
			pids[PID] = true

			_, ok := hashes["111"]
			So(ok, ShouldEqual, true)
			_, ok = hashes["222"]
			So(ok, ShouldEqual, true)
			So(len(pids), ShouldEqual, 1)
		})
		Convey("Stop DownloadLoop", func() {
			dc.Stop()
		})
	})
}

func TestSynchronizer(t *testing.T) {
	ilog.Stop()
	ctl := NewController(t)
	b0 := &block.Block{
		Head: &block.BlockHead{
			Version:    0,
			ParentHash: []byte("nothing"),
			Witness:    "w0",
			Number:     0,
		},
	}
	b0.CalculateHeadHash()
	tpl := "[\"a1\",\"a2\",\"a3\",\"a4\",\"a5\"]"
	statedb := db_mock.NewMockMVCCDB(ctl)
	statedb.EXPECT().Get("state", "b-vote_producer.iost-"+"pendingBlockNumber").AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal("5"), nil
	})
	statedb.EXPECT().Get("state", "b-vote_producer.iost-"+"pendingProducerList").AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(tpl), nil
	})
	statedb.EXPECT().Get("snapshot", "blockHead").AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		bhJson, _ := json.Marshal(b0.Head)
		return string(bhJson), nil
	})
	statedb.EXPECT().Get("state", Any()).AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(`{"loc":"11","url":"22","netId":"33","online":true,"score":0,"votes":0}`), nil
	})
	statedb.EXPECT().Fork().AnyTimes().Return(statedb)
	statedb.EXPECT().Checkout(Any()).AnyTimes().Return(true)
	baseVariable := core_mock.NewMockBaseVariable(ctl)
	baseVariable.EXPECT().StateDB().AnyTimes().Return(statedb)

	config := common.Config{
		DB: &common.DBConfig{
			LdbPath: "DB/",
		},
	}
	baseVariable.EXPECT().Config().AnyTimes().Return(&config)
	baseVariable.EXPECT().BlockChain().AnyTimes().Return(nil)
	baseVariable.EXPECT().Continuous().AnyTimes().Return(0)
	baseVariable.EXPECT().Mode().AnyTimes().Return(global.ModeNormal)
	Convey("Test Synchronizer", t, func() {
		genesis.FakeBv(baseVariable)

		blockcache.CleanBlockCacheWAL()
		blockCache, err := blockcache.NewBlockCache(baseVariable)
		So(err, ShouldBeNil)
		mockController := gomock.NewController(t)
		mockP2PService := p2p_mock.NewMockService(mockController)
		channel := make(chan p2p.IncomingMessage, 1024)
		mockP2PService.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(channel).AnyTimes()
		mockP2PService.EXPECT().Register(gomock.Any(), gomock.Any()).Return(channel)
		mockP2PService.EXPECT().Broadcast(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(a interface{}, b interface{}, c interface{}) {
			channel <- *p2p.NewIncomingMessage("abc", a.([]byte), b.(p2p.MessageType))
		}).AnyTimes()
		sy, err := NewSynchronizer(baseVariable, blockCache, mockP2PService) //mock
		sy.Start()
		So(err, ShouldBeNil)
		err = sy.syncBlocks(1, 15)
		So(err, ShouldBeNil)
		time.Sleep(200 * time.Millisecond)
		blockcache.CleanBlockCacheWAL()

		os.RemoveAll("DB/")
	})
}
