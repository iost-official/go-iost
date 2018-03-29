package protocol

import (
	"IOS/src/iosbase"
	"fmt"
	"math/rand"
	"time"
)

type Replica interface {
	Init(rd *RuntimeData, network *NetworkFilter, recorder Recorder)
	ReplicaLoop()
	OnRequest(request iosbase.Request)
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
	*RuntimeData

	network  *NetworkFilter
	recorder Recorder

	block            *iosbase.Block
	prePrepare       *PrePrepare
	prepare          Prepare
	commit           Commit
	AcceptCount      int
	RejectCount      int
	commitCounts     map[string]int
	correctBlockHash []byte
	reqChan          chan iosbase.Request
}

func (r *ReplicaImpl) Init(rd *RuntimeData, network *NetworkFilter, recorder Recorder) {
	r.RuntimeData = rd
	r.recorder = recorder
	r.network = network

	r.reqChan = make(chan iosbase.Request)

}

func (r *ReplicaImpl) onNewView(view View) (Phase, error) {

	r.block = nil
	r.AcceptCount = 0
	r.RejectCount = 0
	r.commitCounts = nil
	r.correctBlockHash = nil

	r.view = view
	// step 1 determine what character it is
	if r.view.isPrimary(r.ID) {
		r.character = Primary
	} else if r.view.isBackup(r.ID) {
		r.character = Backup
	} else {
		r.character = Idle
	}

	if r.character == Backup {
		return PreparePhase, nil
	} else if r.character == Idle {
		return EndPhase, nil
	}

	// step 2 if it is primary, make a block/pre-prepare package and broadcast it
	r.block = r.recorder.MakeBlock()
	bBlk := r.block.Encode()
	bHH := r.block.Head.Hash()
	sig := iosbase.Sign(iosbase.Sha256(bBlk), r.Seckey)
	pre := PrePrepare{
		sig:         sig,
		pubkey:      r.Pubkey,
		blk:         bBlk,
		blkHeadHash: bHH,
	}

	bPre, err := pre.Marshal(nil)
	if err != nil {
		return PanicPhase, err
	}

	for _, m := range view.GetBackup() {
		r.network.send(iosbase.Request{
			Time:    time.Now().Unix(),
			From:    r.ID,
			To:      m.ID,
			ReqType: int(PrePreparePhase),
			Body:    bPre})
	}

	// step 3, as primary, prepare a Prepare pack which is always true
	r.prepare = r.makePrepare(true)

	return PreparePhase, nil
}

func (r *ReplicaImpl) onPrePrepare(prePrepare *PrePrepare) (Phase, error) {

	// 1. verify block syntax
	r.prePrepare = prePrepare
	r.block.Decode(prePrepare.blk)

	// 2. verify if txs contains tx which validated sign conflict
	// if ok, r.prepare is Accept, vise versa
	if err := r.recorder.VerifyBlock(r.block); err != nil {
		r.prepare = r.makePrepare(true)
	} else {
		r.prepare = r.makePrepare(false)
	}

	// 3. save pre-prepare, block, broadcast prepare
	bPare, err := r.prepare.Marshal(nil)
	if err != nil {
		return PrePreparePhase, err
	}

	r.network.send(iosbase.Request{
		Time: time.Now().Unix(),
		From: r.ID, To: r.view.GetPrimary().ID,
		ReqType: int(PreparePhase),
		Body:    bPare})

	for _, m := range r.view.GetBackup() {
		if m.ID == r.ID {
			continue
		}
		r.network.send(iosbase.Request{
			Time: time.Now().Unix(),
			From: r.ID, To: m.ID,
			ReqType: int(PreparePhase),
			Body:    bPare})
	}

	r.AcceptCount = 0
	r.RejectCount = 0

	return PreparePhase, nil
}

func (r *ReplicaImpl) onPrepare(prepare Prepare) (Phase, error) {
	// count accept and reject numbers, if
	//    1. ac or rj > 2t + 1 , do as it is
	//    2. ac > t && rj > t, or time expired, the system is in failed and no consensus can reach, so put empty in it
	if prepare.isAccept {
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
			r.network.send(iosbase.Request{
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
		r.commitCounts[iosbase.Base58Encode(r.commit.blkHeadHash)] = 1
		return CommitPhase, nil
	} else if r.RejectCount > 1+2*r.view.ByzantineTolerance() {
		return r.onPrepareReject()
	} else if r.RejectCount > r.view.ByzantineTolerance() && r.AcceptCount > r.view.ByzantineTolerance() {
		return r.onTimeOut(PreparePhase)
	}
	return PreparePhase, nil
}

func (r *ReplicaImpl) onPrepareReject() (Phase, error) {
	r.recorder.AdmitEmptyBlock()
	return StartPhase, nil
}

func (r *ReplicaImpl) onCommit(commit Commit) (Phase, error) {

	r.commitCounts[iosbase.Base58Encode(commit.blkHeadHash)]++

	divergence := 0
	for key, value := range r.commitCounts {
		if value > 1+2*r.view.ByzantineTolerance() {
			// if the block is local, ok; otherwise request the correct block
			if key == iosbase.Base58Encode(r.block.Head.Hash()) {
				r.recorder.AdmitBlock(r.block)
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
	r.recorder.AdmitEmptyBlock()
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
	cc.pubkey = r.Pubkey
	cc.blkHeadHash = r.block.Head.Hash()
	cc.sig = iosbase.Sign(cc.blkHeadHash, r.Seckey)
	return cc
}

func (r *ReplicaImpl) ReplicaLoop() {
	r.phase = StartPhase
	var req iosbase.Request
	var err error = nil
	r.isRunning = true

	to := time.NewTimer(1 * time.Minute)

	for r.isRunning {

		switch r.phase {
		case StartPhase:
			v := NewDposView(r.blockChain)
			r.phase, err = r.onNewView(&v)
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
				pp := PrePrepare{}
				pp.Unmarshal(req.Body)
				// verify sender's identity
				if iosbase.Base58Encode(pp.pubkey) != req.From ||
					!iosbase.VerifySignature(pp.blkHeadHash, pp.pubkey, pp.sig) {
					err = fmt.Errorf("fake package")
				} else {
					r.phase, err = r.onPrePrepare(&pp)
				}
			case PreparePhase:
				p := Prepare{}
				p.Unmarshal(req.Body)
				// verify prepare Sign whether comes from right member
				if iosbase.Base58Encode(p.pubkey) != req.From ||
					!iosbase.VerifySignature(p.rand, p.pubkey, p.sig) {
					err = fmt.Errorf("fake package")
				} else {
					r.phase, err = r.onPrepare(p)
				}
			case CommitPhase:
				cm := Commit{}
				cm.Unmarshal(req.Body)
				if iosbase.Base58Encode(cm.pubkey) != req.From ||
					!iosbase.VerifySignature(cm.blkHeadHash, cm.pubkey, cm.sig) {
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
		}
	}
}

func (r *ReplicaImpl) OnRequest(request iosbase.Request) {
	r.reqChan <- request
}
