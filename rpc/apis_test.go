package rpc

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/bouk/monkey"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/event"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/p2p"
)

type MockApisSubscribeServer struct {
	Apis_SubscribeServer
	count int
}

func (s *MockApisSubscribeServer) Send(req *SubscribeRes) error {
	s.count++
	if req.Ev.Topic != event.Event_TransactionResult || req.Ev.Data != "test1" {
		return errors.New("unexpected event topic or data. ev = " + req.Ev.String())
	}
	return nil
}

func TestGRPCServer_ExecTx(t *testing.T) {
	t.Skip("The test need an iserver to run. So fix it later")
	server := "localhost:30002"
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer conn.Close()
	client := NewApisClient(conn)
	rootAccount, err := account.NewKeyPair(common.Base58Decode("1rANSfcRzr4HkhbUFZ7L1Zp69JZZHiDDq5v7dNSbbEqeU4jxy3fszV4HGiaLQEyqVpS1dKT9g7zCVRxBVzuiUzB"), crypto.Ed25519)
	if err != nil {
		t.Fatalf(err.Error())
	}
	newAccount, err := account.NewKeyPair(nil, crypto.Ed25519)
	if err != nil {
		t.Fatalf(err.Error())
	}
	dataString := fmt.Sprintf(`["%v", "%v", %d]`, rootAccount.ID, newAccount.ID, 100000000)
	action := tx.NewAction("iost.system", "Transfer", dataString)
	trx := tx.NewTx([]*tx.Action{action}, make([]string, 0),
		10000, 1, time.Now().Add(time.Second*time.Duration(3)).UnixNano(), 0)
	stx, err := tx.SignTx(trx, rootAccount.ID, []*account.KeyPair{rootAccount})

	resp, err := client.ExecTx(context.Background(), &TxReq{Tx: stx.ToPb()})
	if err != nil {
		t.Fatalf(err.Error())
	}
	if resp.TxReceipt.GasUsage != 303 {
		t.Fatalf("gas used %d. should be 303", resp.TxReceipt.GasUsage)
	}

}

func TestRpcServer_Subscribe(t *testing.T) {
	monkey.Patch(NewRPCServer, func(tp txpool.TxPool, bcache blockcache.BlockCache, _global global.BaseVariable, p2pService p2p.Service) *GRPCServer {
		return &GRPCServer{}
	})

	s := NewRPCServer(nil, nil, nil, nil)
	ec := event.GetEventCollectorInstance()
	req := &SubscribeReq{Topics: []event.Event_Topic{event.Event_TransactionResult}}
	res := MockApisSubscribeServer{
		count: 0,
	}
	go func() {
		err := s.Subscribe(req, &res)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if res.count > 5000 {
			t.Fatalf("should send <= 5000 events. got %d", res.count)
		} else {
			t.Logf("send %d events.", res.count)
		}
	}()

	for i := 0; i < 5000; i++ {
		ec.Post(event.NewEvent(event.Event_TransactionResult, "test1"))
		time.Sleep(time.Microsecond * 50)
	}
}
