package protocol

import (
	"IOS/src/iosbase"
	"fmt"
	"time"
)


//go:generate mockgen -destination mocks/mock_recorder.go -package protocol -source recorder.go


type Recorder interface {
	Init(rd *RuntimeData, nw *NetworkFilter) error
	PublishTx(tx iosbase.Tx) error
	RecorderLoop()
	SendTxPack(member iosbase.Member) error
}

func RecorderFactory(kind string) (Recorder, error) {
	switch kind {
	case "base1.0":
		rep := RecorderImpl{}
		return &rep, nil
	}
	return nil, fmt.Errorf("target recorder not found")
}

type RecorderImpl struct {
	*RuntimeData
	network  *NetworkFilter
	replicas []iosbase.Member

	txPool iosbase.TxPool
}

func (r *RecorderImpl) Init(rd *RuntimeData, nw *NetworkFilter) error {
	r.RuntimeData = rd
	r.network = nw
	return nil
}

func (r *RecorderImpl) PublishTx(tx iosbase.Tx) error {
	if err := r.VerifyTx(tx); err != nil {
		return err
	}
	r.txPool.Add(tx)
	return nil
}

func (r *RecorderImpl) SendTxPack(member iosbase.Member) error {
	req := iosbase.Request{
		From:    r.ID,
		To:      member.ID,
		ReqType: int(ReqSubmitTxPack),
		Body:    r.txPool.Encode(),
		Time:    time.Now().Unix(),
	}

	r.network.Send(req)
	return nil
}

//func (r *RecorderImpl) ApplyNewBlockLoop() {
//	for r.isRunning {
//		// every Period require new block
//		r.blkHashCounter = make(map[string]int)
//		r.blkSourceCache = make(map[string]string)
//
//		view := NewDposView(r.blockChain)
//		r.replicas = append(view.backup, view.primary)
//		for _, m := range r.replicas {
//			req := r.applyBlockHash(m.ID)
//			resChan := r.network.Send(req)
//			if res := <-resChan; res.Code == int(Accepted) {
//				r.onReceiveBlockHash(res.From, res.Description)
//			}
//		}
//		time.Sleep(Period)
//	}
//}
//
//func (r *RecorderImpl) onReceiveBlockHash(senderID string, b58Hash string) {
//	r.blkHashCounter[b58Hash]++
//	if _, ok := r.blkSourceCache[b58Hash]; !ok {
//		r.blkSourceCache[b58Hash] = senderID
//	}
//
//	for key, val := range r.blkHashCounter {
//		if val > len(r.replicas)/3 {
//			r.network.Send(iosbase.Request{
//				From:    r.ID,
//				To:      r.blkSourceCache[key],
//				Time:    time.Now().Unix(),
//				ReqType: int(ReqApplyBlock),
//				Body:    nil,
//			})
//		}
//	}
//}
//
//func (r *RecorderImpl) OnReceiveBlock(block *iosbase.Block) {
//	if r.blkHashCounter[iosbase.Base58Encode(block.Head.Hash())] > len(r.replicas)/3 {
//		r.AdmitBlock(block)
//	}
//}
//
//func (r *RecorderImpl) OnAppliedBlock(ID string) {
//	r.network.Send(iosbase.Request{
//		From:    r.ID,
//		To:      ID,
//		Time:    time.Now().Unix(),
//		ReqType: int(ReqSendBlock),
//		Body:    r.blockChain.Top().Encode(),
//	})
//}
//
//func (r *RecorderImpl) OnAppliedBlockHash(ID string, res chan iosbase.Response) {
//	resp := iosbase.Response{
//		From:        r.ID,
//		To:          ID,
//		Code:        int(Accepted),
//		Description: iosbase.Base58Encode(r.blockChain.Top().Head.Hash()),
//	}
//	res <- resp
//}
//
//func (r *RecorderImpl) applyBlockHash(memberID string) iosbase.Request {
//	bin := iosbase.Binary{}
//	bin.PutInt(r.blockChain.Length())
//	return iosbase.Request{
//		From:    r.ID,
//		To:      memberID,
//		Time:    time.Now().Unix(),
//		ReqType: int(ReqApplyBlockHash),
//		Body:    bin.Bytes(),
//	}
//}

func (r *RecorderImpl) RecorderLoop() {
	for r.isRunning {
		if r.txPool.Size() > 0 {
			for _, m := range r.view.GetBackup() {
				r.SendTxPack(m)
			}
		}
	}
}
