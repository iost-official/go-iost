package protocol

import (
	"IOS/src/iosbase"
	"fmt"
	"time"
)

type Recorder interface {
	Init(rd *RuntimeData, nw *NetworkFilter, bc iosbase.BlockChain, sp iosbase.StatePool) error
	PublishTx(tx iosbase.Tx) error
	RecorderLoop()
	AdmitBlock(block *iosbase.Block)
	AdmitEmptyBlock()
	MakeBlock() *iosbase.Block
	VerifyBlock(block *iosbase.Block) error
	OnReceiveBlock(block *iosbase.Block)
	OnAppliedBlock(ID string)
	OnAppliedBlockHash(ID string, res chan iosbase.Response)
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

	blkHashCache   map[string]int
	blkSourceCache map[string]string
}

func (r *RecorderImpl) Init(rd *RuntimeData, nw *NetworkFilter, bc iosbase.BlockChain, sp iosbase.StatePool) error {
	r.RuntimeData = rd
	r.blockChain = bc
	r.statePool = sp
	return nil
}

func (r *RecorderImpl) verifyTx(tx iosbase.Tx) error {
	// here only existence of Tx inputs will be verified
	for _, in := range tx.Inputs {
		if _, err := r.statePool.Find(in.StateHash); err == nil {
			return fmt.Errorf("some input not found")
		}
	}
	return nil
}

func (r *RecorderImpl) VerifyBlock(block *iosbase.Block) error {
	if !iosbase.Equal(block.SuperHash, r.blockChain.Top().HeadHash()) {
		return fmt.Errorf("unrelated block")
	}
	var txpool iosbase.TxPool
	err := txpool.Decode(block.Content)
	if err != nil {
		return err
	}
	txs, err := txpool.GetSlice()
	if err != nil {
		return err
	}
	for _, tx := range txs {
		if err := r.verifyTx(tx); err != nil {
			return err
		}
	}
	return nil
}

func (r *RecorderImpl) MakeBlock() *iosbase.Block {

	head := iosbase.BlockHead{
		Version:   Version,
		SuperHash: r.blockChain.Top().HeadHash(),
		TreeHash:  r.txPool.Hash(),
		Time:      time.Now().Unix(),
	}
	r.blockChain.Push(iosbase.Block{
		Version:   Version,
		SuperHash: r.blockChain.Top().HeadHash(),
		Head:      head,
		Content:   r.txPool.Encode(),
	})
	return &iosbase.Block{}
}

func (r *RecorderImpl) AdmitEmptyBlock() {
	baseBlk, err := r.blockChain.Get(r.blockChain.Length())
	if err != nil {
		panic(err)
	}
	head := iosbase.BlockHead{
		Version:   Version,
		SuperHash: baseBlk.Head.Hash(),
		TreeHash:  make([]byte, 32),
		Time:      time.Now().Unix(),
	}
	r.blockChain.Push(iosbase.Block{
		Version:   Version,
		SuperHash: baseBlk.Head.Hash(),
		Head:      head,
		Content:   nil,
	})
}

func (r *RecorderImpl) AdmitBlock(block *iosbase.Block) {
	r.blockChain.Push(*block)
}

func (r *RecorderImpl) PublishTx(tx iosbase.Tx) error {
	if err := r.verifyTx(tx); err != nil {
		return err
	}
	r.txPool.Add(tx)
	return nil
}

func (r *RecorderImpl) RecorderLoop() {
	for r.isRunning {
		// every Period require new block
		r.blkHashCache = make(map[string]int)
		r.blkSourceCache = make(map[string]string)

		view := NewDposView(r.blockChain)
		r.replicas = append(view.backup, view.primary)
		for _, m := range r.replicas {
			req := r.applyBlockHash(m.ID)
			resChan := r.network.send(req)
			if res := <-resChan; res.Code == int(Accepted) {
				r.onReceiveBlockHash(res.From, res.Description)
			}
		}
		time.Sleep(Period)
	}
}

func (r *RecorderImpl) onReceiveBlockHash(senderID string, b58Hash string) {
	r.blkHashCache[b58Hash]++
	if _, ok := r.blkSourceCache[b58Hash]; !ok {
		r.blkSourceCache[b58Hash] = senderID
	}

	for key, val := range r.blkHashCache {
		if val > len(r.replicas)/3 {
			r.network.send(iosbase.Request{
				From:    r.ID,
				To:      r.blkSourceCache[key],
				Time:    time.Now().Unix(),
				ReqType: int(ReqApplyBlock),
				Body:    nil,
			})
		}
	}
}

func (r *RecorderImpl) OnReceiveBlock(block *iosbase.Block) {
	if r.blkHashCache[iosbase.Base58Encode(block.Head.Hash())] > len(r.replicas)/3 {
		r.AdmitBlock(block)
	}
}

func (r *RecorderImpl) OnAppliedBlock(ID string) {
	r.network.send(iosbase.Request{
		From:    r.ID,
		To:      ID,
		Time:    time.Now().Unix(),
		ReqType: int(ReqSendBlock),
		Body:    r.blockChain.Top().Encode(),
	})
}

func (r *RecorderImpl) OnAppliedBlockHash(ID string, res chan iosbase.Response) {
	resp := iosbase.Response{
		From:        r.ID,
		To:          ID,
		Code:        int(Accepted),
		Description: iosbase.Base58Encode(r.blockChain.Top().Head.Hash()),
	}
	res <- resp
}

func (r *RecorderImpl) applyBlockHash(memberID string) iosbase.Request {
	bin := iosbase.Binary{}
	bin.PutInt(r.blockChain.Length())
	return iosbase.Request{
		From:    r.ID,
		To:      memberID,
		Time:    time.Now().Unix(),
		ReqType: int(ReqApplyBlockHash),
		Body:    bin.Bytes(),
	}
}
