package protocol

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/iost-official/PrototypeWorks/iosbase"
)

//go:generate mockgen -destination mocks/mock_component.go -package protocol_mock github.com/iost-official/PrototypeWorks/protocol Component

type Phase int

const (
	StartPhase Phase = iota
	PrePreparePhase
	PreparePhase
	CommitPhase
	PanicPhase
	EndPhase
)

const (
	PrePrepareTimeout = 58 * time.Second
	PrepareTimeout    = 1 * time.Second
	CommitTimeout     = 1 * time.Second
)

type Component interface {
	Init(self iosbase.Member, db Database, router Router) error
	Run()
	Stop()
}

func ReplicaFactory(target string, pool iosbase.TxPool) (Component, error) {
	switch target {
	case "pbft":
		rep := ReplicaImpl{
			txPool: pool,
		}
		return &rep, nil
	}
	return nil, fmt.Errorf("target replica not found")
}

/*
Implement of Component replica

Replica runs the PBFT protocol to reach consensus with other replicas,
and make new block to block chain. reward would be given to every replicas
so members of ios network would be willing to devote their power and network
to be a replica.
*/
type ReplicaImpl struct {
	txPool iosbase.TxPool
	block  *iosbase.Block
	db     Database
	net    Router
	iosbase.Member

	phase            Phase
	sblk             *SignedBlock
	AcceptCount      int
	RejectCount      int
	commitCounts     map[string]int
	correctBlockHash []byte

	chView                            chan View             // in
	chPrePrepare, chPrepare, chCommit chan iosbase.Request  // in
	chTxPack                          chan iosbase.Request  // in
	chReply                           chan iosbase.Response // out, in
	ExitSignal                        chan bool             // in

	view  View
	chara Character
}

func (r *ReplicaImpl) Init(self iosbase.Member, db Database, router Router) error {
	r.Member = self
	r.db = db
	r.net = router
	r.txPool = &iosbase.TxPoolImpl{}

	r.ExitSignal = make(chan bool)

	var err error
	r.Member = self
	r.chView, err = db.NewViewSignal()
	if err != nil {
		return err
	}

	filter := Filter{
		AcceptType: []ReqType{ReqSubmitTxPack},
	}
	r.chTxPack, r.chReply, err = router.FilteredChan(filter)
	if err != nil {
		return err
	}

	filter = Filter{
		AcceptType: []ReqType{ReqPrePrepare},
	}
	r.chPrePrepare, _, err = router.FilteredChan(filter)
	if err != nil {
		return err
	}

	filter = Filter{
		AcceptType: []ReqType{ReqPrepare},
	}
	r.chPrepare, _, err = router.FilteredChan(filter)
	if err != nil {
		return err
	}

	filter = Filter{
		AcceptType: []ReqType{ReqCommit},
	}
	r.chCommit, _, err = router.FilteredChan(filter)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReplicaImpl) Run() {
	go r.collectLoop()
	go r.pbftLoop()
}

func (r *ReplicaImpl) Stop() {
	r.ExitSignal <- true
}

func (r *ReplicaImpl) collectLoop() {
	for true {
		select {
		case req := <-r.chTxPack:
			var pool iosbase.TxPool
			pool.Decode(req.Body)
			txs, err := pool.GetSlice()
			if err != nil {
				r.chReply <- syntaxError(req)
			}
			for _, tx := range txs {
				if err := r.db.VerifyTxWithCache(tx, r.txPool); err == nil {
					r.txPool.Add(tx)
				} else {
					r.chReply <- illegalTx(req)
				}
			}
		case <-r.ExitSignal:
			return
		}
	}
}

func (r *ReplicaImpl) pbftLoop() {
	r.phase = StartPhase
	var req iosbase.Request
	var err error

	switch r.phase {
	case StartPhase:
		select {
		case r.view = <-r.chView:
			r.phase, err = r.onStart()
			if err != nil {
				r.phase = PanicPhase
			}
		case <-r.ExitSignal:
			r.phase = EndPhase
		}
	case PrePreparePhase:
		timeout := time.NewTimer(PrePrepareTimeout)
		select {
		case req = <-r.chPrePrepare:
			r.phase, err = r.onPrePrepare(req)
		case <-r.ExitSignal:
			r.phase = EndPhase
		case <-timeout.C:
			r.phase, err = r.onTimeOut(PrePreparePhase)
		}
	case PreparePhase:
		timeout := time.NewTimer(PrepareTimeout)
		select {
		case req = <-r.chPrepare:
			r.phase, err = r.onPrepare(req)
		case <-r.ExitSignal:
			r.phase = EndPhase
		case <-timeout.C:
			r.phase, err = r.onTimeOut(PreparePhase)
		}
	case CommitPhase:
		timeout := time.NewTimer(CommitTimeout)
		select {
		case req = <-r.chCommit:
			r.phase, err = r.onCommit(req)
		case <-r.ExitSignal:
			r.phase = EndPhase
		case <-timeout.C:
			r.phase, err = r.onTimeOut(CommitPhase)
		}
	case PanicPhase:
		fmt.Println(err)
		fallthrough
	case EndPhase:
		return
	}
}

func (r *ReplicaImpl) onStart() (Phase, error) {
	r.block = nil
	r.AcceptCount = 0
	r.RejectCount = 0
	r.commitCounts = nil
	r.correctBlockHash = nil

	// step 1 determine what character it is
	if r.view.IsPrimary(r.ID) {
		r.chara = Primary
	} else if r.view.IsBackup(r.ID) {
		r.chara = Backup
	} else {
		r.chara = Idle
	}

	if r.chara == Backup {
		return PrePreparePhase, nil
	} else if r.chara == Idle {
		return StartPhase, nil
	}

	// step 2 if it is primary, make a block/pre-prepare package and broadcast it
	var err error
	r.block, err = r.makeBlock()
	if err != nil {
		return PanicPhase, err
	}
	bBlk := r.block.Encode()
	bHH := r.block.Head.Hash()
	sig := iosbase.Sign(iosbase.Sha256(bHH), r.Seckey)
	pre := SignedBlock{
		Sig:         sig,
		Pubkey:      r.Pubkey,
		Blk:         bBlk,
		BlkHeadHash: bHH,
	}

	bPre, err := pre.Marshal(nil)
	if err != nil {
		return PanicPhase, err
	}

	for _, m := range r.view.GetBackup() {
		r.net.Send(iosbase.Request{
			Time:    time.Now().Unix(),
			From:    r.ID,
			To:      m.ID,
			ReqType: int(PrePreparePhase),
			Body:    bPre,
		})
	}

	r.AcceptCount = 0
	r.RejectCount = 0
	return PreparePhase, nil
}

func (r *ReplicaImpl) onPrePrepare(req iosbase.Request) (Phase, error) {

	// 0. verify sender

	var sblk SignedBlock
	_, err := sblk.Unmarshal(req.Body)
	if err != nil {
		r.chReply <- syntaxError(req)
		return PrePreparePhase, nil
	}

	//if !r.view.IsPrimary(iosbase.Base58Encode(iosbase.Hash160(sblk.Pubkey))) ||
	//	!iosbase.VerifySignature(sblk.BlkHeadHash, sblk.Pubkey, sblk.Sig) {
	//	return PrePreparePhase, nil
	//}

	// 1. verify block syntax
	r.sblk = &sblk
	var prepare Prepare
	err = r.block.Decode(sblk.Blk)
	if err != nil {
		r.chReply <- syntaxError(req)
		prepare = r.makePrepare(false)
	} else {
		r.chReply <- accept(req)
		prepare = r.makePrepare(true)
	}

	// 2. verify if txs contains tx which validated sign conflict
	// if ok, prepare is Accept, vise versa
	err = r.db.VerifyBlockWithCache(r.block, r.txPool)
	if err != nil {
		r.chReply <- illegalTx(req)
		prepare = r.makePrepare(false)
	} else {
		r.chReply <- accept(req)
		prepare = r.makePrepare(true)
	}

	err = r.broadcastPrepare(prepare)
	if err != nil {
		return PanicPhase, err
	}

	// initial next phase at end of previous phase
	r.AcceptCount = 0
	r.RejectCount = 0
	return PreparePhase, nil
}

func (r *ReplicaImpl) onPrepare(req iosbase.Request) (Phase, error) {

	var prepare Prepare
	err := prepare.parse(req.Body)
	if err != nil {
		r.chReply <- syntaxError(req)
		return PreparePhase, err
	}

	realSender := iosbase.Base58Encode(iosbase.Hash160(prepare.Pubkey))
	if !r.view.IsBackup(realSender) ||
		!iosbase.VerifySignature(prepare.Rand, prepare.Pubkey, prepare.Sig) {
		r.chReply <- authorityError(req)
		return PreparePhase, nil
	}

	if time.Now().Unix()-iosbase.GetInt64(prepare.Rand, 1) > 60 {
		r.chReply <- authorityError(req)
		return PreparePhase, nil
	}

	// count accept and reject numbers, if
	//    1. ac or rj > 2t + 1 , do as it is
	//    2. ac > t && rj > t, or time expired, the system is in failed and no consensus can reach, so put empty in it
	if prepare.IsAccept {
		r.AcceptCount++
	} else {
		r.RejectCount++
	}

	if r.AcceptCount > 2*r.view.ByzantineTolerance() { // the prepare of Primary should be true, so ac + 1
		commit := r.makeCommit()

		bCmt, err := commit.Marshal(nil)
		if err != nil {
			return PreparePhase, err
		}
		for _, m := range r.view.GetBackup() {
			r.net.Send(iosbase.Request{
				Time:    time.Now().Unix(),
				From:    r.ID,
				To:      m.ID,
				ReqType: int(PreparePhase),
				Body:    bCmt,
			})
		}
		r.AcceptCount = 0
		r.RejectCount = 0
		r.commitCounts = make(map[string]int, r.view.ByzantineTolerance())
		r.commitCounts[iosbase.Base58Encode(commit.BlkHeadHash)] = 1
		return CommitPhase, nil
	} else if r.RejectCount > 1+2*r.view.ByzantineTolerance() {
		return r.onPrepareReject()
	} else if r.RejectCount > r.view.ByzantineTolerance() && r.AcceptCount > r.view.ByzantineTolerance() {
		return r.onTimeOut(PreparePhase)
	}

	return PreparePhase, nil
}

func (r *ReplicaImpl) onPrepareReject() (Phase, error) {
	err := r.admitEmptyBlock()
	return StartPhase, err
}

func (r *ReplicaImpl) onCommit(req iosbase.Request) (Phase, error) {

	var commit Commit
	_, err := commit.Unmarshal(req.Body)
	if err != nil {
		r.chReply <- syntaxError(req)
		return CommitPhase, err
	}

	// TODO: ensure sender's identity in iosbase.network
	realSender := iosbase.Base58Encode(iosbase.Hash160(commit.Pubkey))
	if !(r.view.IsBackup(realSender) || r.view.IsPrimary(realSender)) ||
		!iosbase.VerifySignature(commit.BlkHeadHash, commit.Pubkey, commit.Sig) {
		r.chReply <- authorityError(req)
		return CommitPhase, nil
	}

	r.commitCounts[iosbase.Base58Encode(commit.BlkHeadHash)]++

	divergence := 0
	for key, value := range r.commitCounts {
		if value > 1+2*r.view.ByzantineTolerance() {
			// if the block is local, ok; otherwise request the correct block
			if key == iosbase.Base58Encode(r.block.Head.Hash()) {
				r.admitBlock(r.block)
			} else {

			}
			return StartPhase, nil
		}
		if value > r.view.ByzantineTolerance() {
			divergence++
			if divergence > 2 {
				return r.onCommitFailed()
			}
		}
	}
	return CommitPhase, nil
}

func (r *ReplicaImpl) onCommitFailed() (Phase, error) {
	r.admitEmptyBlock()
	return StartPhase, nil
}

func (r *ReplicaImpl) onTimeOut(phase Phase) (Phase, error) {
	switch phase {
	case PrePreparePhase:
		// if backup did not receive PPP pack, simply mark reject and goes to prepare phase
		prepare := r.makePrepare(false)
		r.broadcastPrepare(prepare)
		return PreparePhase, nil
	case PreparePhase:
		// time out means offline committee members or divergence > t
		return r.onPrepareReject()
	case CommitPhase:
		//
		return r.onCommitFailed()
	}
	return StartPhase, nil
}

func (r *ReplicaImpl) broadcastPrepare(prepare Prepare) error {
	// 3. broadcast prepare
	bPare, err := prepare.Marshal(nil)
	if err != nil {
		return err
	}

	r.net.Send(iosbase.Request{
		Time:    time.Now().Unix(),
		From:    r.ID,
		To:      r.view.GetPrimary().ID,
		ReqType: int(PreparePhase),
		Body:    bPare})

	for _, m := range r.view.GetBackup() {
		r.net.Send(iosbase.Request{
			Time:    time.Now().Unix(),
			From:    r.ID,
			To:      m.ID,
			ReqType: int(PreparePhase),
			Body:    bPare})
	}
	return nil
}

func (r *ReplicaImpl) makePrepare(isAccept bool) Prepare {
	rand.Seed(time.Now().UnixNano())
	tb := iosbase.NewBinary()
	tb.PutULong(uint64(time.Now().Unix()))
	bin := iosbase.NewBinary()
	for i := 0; i < 3; i++ {
		bin.PutULong(rand.Uint64())
	}
	randomByte := bin.Bytes()

	var vote []byte
	if isAccept {
		vote = []byte{0x00}
	} else {
		vote = []byte{0xFF}
	}
	vote = append(vote, tb.Bytes()...)
	vote = append(vote, randomByte[9:]...)

	pareSig := iosbase.Sign(vote, r.Seckey)
	prepare := Prepare{
		prepareRaw: prepareRaw{pareSig, r.Pubkey, vote},
		IsAccept:   isAccept,
	}

	return prepare
}

func (r *ReplicaImpl) makeCommit() Commit {
	var cc Commit
	cc.Pubkey = r.Pubkey
	cc.BlkHeadHash = r.block.Head.Hash()
	cc.Sig = iosbase.Sign(cc.BlkHeadHash, r.Seckey)
	return cc
}

func (r *ReplicaImpl) makeBlock() (*iosbase.Block, error) {

	blockChain, err := r.db.GetBlockChain()
	if err != nil {
		return nil, err
	}

	blockHead := iosbase.BlockHead{
		Version:   Version,
		SuperHash: blockChain.Top().HeadHash(),
		TreeHash:  r.txPool.Hash(),
		Time:      time.Now().Unix(),
	}

	block := iosbase.Block{
		Version:   Version,
		SuperHash: blockChain.Top().HeadHash(),
		Head:      blockHead,
		Content:   r.txPool.Encode(),
	}

	return &block, nil
}

func (r *ReplicaImpl) admitBlock(block *iosbase.Block) error {
	//r.blockChain.Push(*block)
	r.net.Broadcast(iosbase.Request{
		From:    r.ID,
		To:      "",
		ReqType: int(ReqNewBlock),
		Body:    block.Encode(),
	})
	return nil
}

func (r *ReplicaImpl) admitEmptyBlock() error {
	bc, err := r.db.GetBlockChain()
	if err != nil {
		return err
	}
	baseBlk := bc.Top()

	head := iosbase.BlockHead{
		Version:   Version,
		SuperHash: baseBlk.Head.Hash(),
		TreeHash:  make([]byte, 32),
		Time:      time.Now().Unix(),
	}

	blk := iosbase.Block{
		Version:   Version,
		SuperHash: baseBlk.Head.Hash(),
		Head:      head,
		Content:   nil,
	}
	r.net.Broadcast(iosbase.Request{
		From:    r.ID,
		To:      "",
		ReqType: int(ReqNewBlock),
		Body:    blk.Encode(),
	})
	return nil
}

/*
Struct to transport vote of prepare phase.

signed to 0x00(vote for accept) or 0xFF(vote for reject) + timestamp + 23 bytes random code.
Timestamp should bigger than preview block's timestamp.
*/
type Prepare struct {
	prepareRaw
	IsAccept bool
}

func (p *Prepare) parse(b []byte) error {
	p.prepareRaw.Unmarshal(b)
	if p.Rand[0] == 0x00 {
		p.IsAccept = true
		return nil
	} else if p.Rand[0] == 0xFF {
		p.IsAccept = false
		return nil
	} else {
		return fmt.Errorf("syntax error")
	}
}
