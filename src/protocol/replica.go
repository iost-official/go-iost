package protocol

import (
	"IOS/src/iosbase"
	"fmt"
	"math/rand"
	"time"
)

//go:generate mockgen -destination replica_mock_test.go -package protocol -source replica.go

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
	PrePrepareTimeout = 30 * time.Second
	PrepareTimeout    = 10 * time.Second
	CommitTimeout     = 20 * time.Second
	PBFTPeriod        = 1 * time.Minute
)

type Replica interface {
	Init(db Database, router Router) error
	Run()
	Stop()
}

func ReplicaFactory(kind string) (Replica, error) {
	switch kind {
	case "base1.0":
		rep := ReplicaImpl{}
		return &rep, nil
	}
	return nil, fmt.Errorf("target replica not found")
}

type ReplicaImpl struct {
	txPool iosbase.TxPool
	block  *iosbase.Block
	Database

	phase            Phase
	prePrepare       *SignedBlock
	prepare          Prepare
	commit           Commit
	AcceptCount      int
	RejectCount      int
	commitCounts     map[string]int
	correctBlockHash []byte

	iosbase.Member
	chView                            chan View             // in
	chPrePrepare, chPrepare, chCommit chan iosbase.Request  // in
	chTxPack                          chan iosbase.Request  // in
	chSend                            chan iosbase.Request  // out
	chReply, chRes                    chan iosbase.Response // out, in
	ExitSignal                        chan bool             // in

	view  View
	chara Character
}

func (r *ReplicaImpl) Init(db Database, router Router) error {
	var err error
	r.Member, err = db.GetIdentity()
	if err != nil {
		return err
	}
	r.chView, err = db.NewViewSignal()
	if err != nil {
		return err
	}

	filter := Filter{
		AcceptType: []ReqType{ReqSubmitTxPack},
	}
	r.chTxPack, err = router.FilteredReqChan(filter)
	if err != nil {
		return err
	}
	r.chSend, r.chRes, err = router.SendChan()
	if err != nil {
		return err
	}
	r.chReply, err = router.ReplyChan()
	if err != nil {
		return err
	}
	filter = Filter{
		AcceptType: []ReqType{ReqPrePrepare},
	}
	r.chPrePrepare, err = router.FilteredReqChan(filter)
	if err != nil {
		return err
	}

	filter = Filter{
		AcceptType: []ReqType{ReqPrepare},
	}
	r.chPrepare, err = router.FilteredReqChan(filter)
	if err != nil {
		return err
	}

	filter = Filter{
		AcceptType: []ReqType{ReqCommit},
	}
	r.chCommit, err = router.FilteredReqChan(filter)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReplicaImpl) Run() {
	go r.collectLoop()
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
				if err := r.verifyTxWithCache(tx); err == nil {
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
		r.prePrepare = nil
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
		r.AcceptCount = 0
		r.RejectCount = 0

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
	r.block = r.makeBlock()
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
		r.chSend <- iosbase.Request{
			Time:    time.Now().Unix(),
			From:    r.ID,
			To:      m.ID,
			ReqType: int(PrePreparePhase),
			Body:    bPre}
	}

	// step 3, as primary, prepare a Prepare pack which is always true
	r.prepare = r.makePrepare(true)

	return PreparePhase, nil
}

func (r *ReplicaImpl) Stop() {
	r.ExitSignal <- true
}

func (r *ReplicaImpl) onPrePrepare(req iosbase.Request) (Phase, error) {

	// 0. verify sender

	var sblk SignedBlock
	_, err := sblk.Unmarshal(req.Body)
	if err != nil {
		return PanicPhase, err
	}

	if !r.view.IsPrimary(iosbase.Base58Encode(iosbase.Hash160(sblk.Pubkey))) ||
		!iosbase.VerifySignature(sblk.BlkHeadHash, sblk.Pubkey, sblk.Sig) {
		return PrePreparePhase, nil
	}

	// 1. verify block syntax
	r.prePrepare = &sblk
	err = r.block.Decode(sblk.Blk)
	if err != nil {
		r.chReply <- syntaxError(req)
		r.prepare = r.makePrepare(true)
	} else {
		r.chReply <- accept(req)
		r.prepare = r.makePrepare(false)
	}

	// 2. verify if txs contains tx which validated sign conflict
	// if ok, r.prepare is Accept, vise versa
	if err := r.verifyBlockWithCache(r.block); err != nil {
		r.chReply <- illegalTx(req)
		r.prepare = r.makePrepare(true)
	} else {
		r.chReply <- accept(req)
		r.prepare = r.makePrepare(false)
	}

	// 3. save pre-prepare, block, broadcast prepare
	bPare, err := r.prepare.Marshal(nil)
	if err != nil {
		return PanicPhase, err
	}

	r.chSend <- iosbase.Request{
		Time:    time.Now().Unix(),
		From:    r.ID,
		To:      r.view.GetPrimary().ID,
		ReqType: int(PreparePhase),
		Body:    bPare}

	for _, m := range r.view.GetBackup() {
		if m.ID == r.ID {
			continue
		}
		r.chSend <- iosbase.Request{
			Time:    time.Now().Unix(),
			From:    r.ID,
			To:      r.view.GetPrimary().ID,
			ReqType: int(PreparePhase),
			Body:    bPare}
	}

	return PreparePhase, nil
}

// depreciated ====================

func (r *ReplicaImpl) Launch() {
	go r.loop()
}

func (r *ReplicaImpl) onPrepare(prepare Prepare) (Phase, error) {
	// count accept and reject numbers, if
	//    1. ac or rj > 2t + 1 , do as it is
	//    2. ac > t && rj > t, or time expired, the system is in failed and no consensus can reach, so put empty in it
	if prepare.IsAccept {
		r.AcceptCount++
	} else {
		r.RejectCount++
	}

	if r.AcceptCount > 1+2*r.view.ByzantineTolerance() {
		r.commit = r.makeCommit()

		bCmt, err := r.commit.Marshal(nil)
		if err != nil {
			return PreparePhase, err
		}
		for _, m := range r.view.GetBackup() {
			if m.ID == r.ID {
				continue
			}
			r.network.Send(iosbase.Request{
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
		r.commitCounts[iosbase.Base58Encode(r.commit.BlkHeadHash)] = 1
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

func (r *ReplicaImpl) onCommit(commit Commit) (Phase, error) {

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
		r.prepare = r.makePrepare(false)
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

func (r *ReplicaImpl) makePrepare(isAccept bool) Prepare {
	random := rand.NewSource(time.Now().Unix())
	bin := iosbase.Binary{}
	randomByte := bin.PutULong(uint64(random.Int63())).Bytes()

	var vote []byte
	if isAccept {
		vote = []byte{0x00}
	} else {
		vote = []byte{0xFF}
	}
	vote = append(vote, randomByte...)

	pareSig := iosbase.Sign(vote, r.Seckey)
	prepare := Prepare{pareSig, r.Pubkey, randomByte, isAccept}

	return prepare
}

func (r *ReplicaImpl) makeCommit() Commit {
	var cc Commit
	cc.Pubkey = r.Pubkey
	cc.BlkHeadHash = r.block.Head.Hash()
	cc.Sig = iosbase.Sign(cc.BlkHeadHash, r.Seckey)
	return cc
}

func (r *ReplicaImpl) loop() { // TODO : BUG and too many things to be done in one function
	r.phase = StartPhase
	var req iosbase.Request
	var err error = nil

	to := time.NewTimer(1 * time.Minute)

	for true {
		switch r.phase {
		case StartPhase:
			r.phase, err = r.onStart()
		case PanicPhase:
			return
		case EndPhase:
			return
		}

		if err != nil {
			fmt.Println(err)
		}

		select {
		case <-r.reqChan:
			req = <-r.reqChan

			switch r.phase {
			case PrePreparePhase:
				pp := SignedBlock{}
				pp.Unmarshal(req.Body)
				// verify sender's identity
				if iosbase.Base58Encode(pp.Pubkey) != req.From ||
					!iosbase.VerifySignature(pp.BlkHeadHash, pp.Pubkey, pp.Sig) {
					err = fmt.Errorf("fake package")
				} else {
					r.phase, err = r.onPrePrepare(&pp)
				}
			case PreparePhase:
				p := Prepare{}
				p.Unmarshal(req.Body)
				// verify prepare Sign whether comes from right member
				if iosbase.Base58Encode(p.Pubkey) != req.From ||
					!iosbase.VerifySignature(p.Rand, p.Pubkey, p.Sig) {
					err = fmt.Errorf("fake package")
				} else {
					r.phase, err = r.onPrepare(p)
				}
			case CommitPhase:
				cm := Commit{}
				cm.Unmarshal(req.Body)
				if iosbase.Base58Encode(cm.Pubkey) != req.From ||
					!iosbase.VerifySignature(cm.BlkHeadHash, cm.Pubkey, cm.Sig) {
					err = fmt.Errorf("fake package")
				} else {
					r.phase, err = r.onCommit(cm)

				}
			}

			if !to.Stop() {
				<-to.C
			}
			to.Reset(ExpireTime)
		case <-to.C:
			r.phase, err = r.onTimeOut(r.phase)
			if err != nil {
				return
			}
			to.Reset(ExpireTime)
		case <-r.ExitSignal:
			return
		}
	}
}

func (r *ReplicaImpl) OnRequest(req iosbase.Request) iosbase.Response {
	switch r.phase {
	case StartPhase:
		return invalidPhase(req)
	case PrePreparePhase:
		if req.ReqType == int(ReqPrePrepare) {
			r.reqChan <- req
			return accept(req)
		} else {
			return invalidPhase(req)
		}
	case PreparePhase:
		if req.ReqType == int(ReqPrepare) {
			r.reqChan <- req
			return accept(req)

		} else {
			return invalidPhase(req)
		}
	case CommitPhase:
		if req.ReqType == int(ReqCommit) {
			r.reqChan <- req
			return accept(req)

		} else {
			return invalidPhase(req)
		}
	case PanicPhase:
		return internalError(req)
	case EndPhase:
		return invalidPhase(req)
	default:
		return internalError(req)
	}
}

func (r *ReplicaImpl) makeBlock() *iosbase.Block {
	blockHead := iosbase.BlockHead{
		Version:   Version,
		SuperHash: r.blockChain.Top().HeadHash(),
		TreeHash:  r.txPool.Hash(),
		Time:      time.Now().Unix(),
	}

	block := iosbase.Block{
		Version:   Version,
		SuperHash: r.blockChain.Top().HeadHash(),
		Head:      blockHead,
		Content:   r.txPool.Encode(),
	}

	return &block
}

func (r *ReplicaImpl) verifyBlockWithCache(block *iosbase.Block) error {
	var blkTxPool iosbase.TxPool
	blkTxPool.Decode(block.Content)

	txs, _ := blkTxPool.GetSlice()
	for i, tx := range txs {
		if i == 0 { // verify coinbase tx
			continue
		}
		err := r.verifyTxWithCache(tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ReplicaImpl) admitBlock(block *iosbase.Block) error {
	//r.blockChain.Push(*block)
	r.network.BroadcastToMembers(iosbase.Request{
		From:    r.ID,
		To:      "",
		ReqType: int(ReqNewBlock),
		Body:    block.Encode(),
	})
	return nil
}

func (r *ReplicaImpl) admitEmptyBlock() error {
	baseBlk, err := r.blockChain.Get(r.blockChain.Length())
	if err != nil {
		return err
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
	return nil
}

func (r *ReplicaImpl) OnTxPack(pool iosbase.TxPool) error {

	return nil
}

// reject duplicated Tx, which might come from corruption actions
func (r *ReplicaImpl) verifyTxWithCache(tx iosbase.Tx) error {
	err := r.RuntimeData.VerifyTx(tx)
	if err != nil {
		return err
	}
	txs, _ := r.txPool.GetSlice()
	for _, existedTx := range txs {
		if iosbase.Equal(existedTx.Hash(), tx.Hash()) {
			return fmt.Errorf("has included")
		}
		if txConflict(existedTx, tx) {
			r.txPool.Del(existedTx) // TODO : BUG, if there are three txs with different recorder
			return fmt.Errorf("conflicted")
		} else if sliceIntersect(existedTx.Inputs, tx.Inputs) {
			return fmt.Errorf("conflicted")
		}
	}
	return nil
}

func sliceEqualI(a, b []iosbase.TxInput) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if !iosbase.Equal(a[i].Hash(), b[i].Hash()) {
			return false
		}
	}
	return true
}

func sliceEqualS(a, b []iosbase.State) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if !iosbase.Equal(a[i].Hash(), b[i].Hash()) {
			return false
		}
	}
	return true
}

func sliceIntersect(a []iosbase.TxInput, b []iosbase.TxInput) bool {
	for _, ina := range a {
		for _, inb := range b {
			if iosbase.Equal(ina.Hash(), inb.Hash()) {
				return true
			}
		}
	}
	return false
}

func txConflict(a, b iosbase.Tx) bool {
	if sliceEqualI(a.Inputs, b.Inputs) &&
		sliceEqualS(a.Outputs, b.Outputs) &&
		a.Recorder != b.Recorder {
		return true
	} else {
		return false
	}
}
