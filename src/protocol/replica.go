package protocol

import (
	"IOS/src/iosbase"
	"math/rand"
	"time"
	"fmt"
)

type Replica struct {
	character Character
	view      View

	block            iosbase.Block
	prePrepare       *PrePrepare
	prepare          Prepare
	commit           Commit
	AcceptCount      int
	RejectCount      int
	commitCounts     map[string]int
	correctBlockHash []byte
}

func (c *Consensus) onNewView(view View) (Phase, error) {
	c.view = view
	// step 1 determine what character it is
	if c.view.isPrimary(c.ID) {
		c.character = Primary
	} else if c.view.isBackup(c.ID) {
		c.character = Backup
	} else {
		c.character = Idle
	}

	if c.character == Backup {
		return PreparePhase, nil
	} else if c.character == Idle {
		return EndPhase, nil
	}

	// step 2 if it is primary, make a block/pre-prepare package and broadcast it
	c.block = c.makeBlock()
	bBlk := c.block.Encode()

	sig := iosbase.Sign(iosbase.Sha256(bBlk), c.Seckey)

	pre := PrePrepare{sig, c.Pubkey, bBlk}

	bPre, err := pre.Marshal(nil)
	if err != nil {
		return PanicPhase, err
	}

	for _, m := range view.GetBackup() {
		c.send(iosbase.Request{time.Now().Unix(), c.ID, m.ID, int(PrePreparePhase), bPre})
	}

	// step 3, as primary, prepare a Prepare pack which is always true
	c.prepare = c.makePrepare(true)

	return PreparePhase, nil
}

func (c *Consensus) onPrePrepare(prePrepare *PrePrepare) (Phase, error) {
	// 1. verify block syntax
	c.prePrepare = prePrepare
	c.block.Decode(prePrepare.blk)

	// 2. verify if txs contains tx which validated sign conflict
	// if ok, c.prepare is Accept, vise versa
	c.prepare = c.makePrepare(true)

	// 3. save pre-prepare, block, broadcast prepare
	bPare, err := c.prepare.Marshal(nil)
	if err != nil {
		return PrePreparePhase, err
	}

	c.send(iosbase.Request{time.Now().Unix(), c.ID, c.view.GetPrimary().ID, int(PreparePhase), bPare})
	for _, m := range c.view.GetBackup() {
		if m.ID == c.ID {
			continue
		}
		c.send(iosbase.Request{time.Now().Unix(), c.ID, m.ID, int(PreparePhase), bPare})
	}

	c.AcceptCount = 0
	c.RejectCount = 0

	return PreparePhase, nil
}

func (c *Consensus) onPrepare(prepare Prepare) (Phase, error) {
	// 1. verify prepare Sign whether comes from right member
	// 2. then count accept or reject, if
	//    2.1 2t + 1 ac or rj, do as
	//    2.2 t ac vs t rj, or time expired, the system is in failed, currently we push empty block to go to next turn
	if prepare.isAccept {
		c.AcceptCount ++
	} else {
		c.RejectCount ++
	}

	if c.AcceptCount > 1+2*c.view.ByzantineTolerance() {
		c.commit = c.makeCommit()

		bCmt, err := c.commit.Marshal(nil)
		if err != nil {
			return PreparePhase, err
		}
		for _, m := range c.view.GetBackup() {
			if m.ID == c.ID {
				continue
			}
			c.send(iosbase.Request{time.Now().Unix(), c.ID, m.ID, int(PreparePhase), bCmt})
		}
		c.AcceptCount = 0
		c.RejectCount = 0
		c.commitCounts = make(map[string]int, c.view.ByzantineTolerance())
		c.commitCounts[iosbase.Base58Encode(c.commit.blkHash)] = 1
		return CommitPhase, nil
	} else if c.RejectCount > 1+2*c.view.ByzantineTolerance() {
		return c.onPrepareReject()
	} else if c.RejectCount > c.view.ByzantineTolerance() && c.AcceptCount > c.view.ByzantineTolerance() {
		return c.onTimeOut(PreparePhase)
	}
	return PreparePhase, nil
}

func (c *Consensus) onPrepareReject() (Phase, error) {
	c.makeEmptyBlock()
	return StartPhase, nil
}

func (c *Consensus) onCommit(commit Commit) (Phase, error) {

	// 1. verify the sender is correct
	c.commitCounts[iosbase.Base58Encode(commit.blkHash)] ++

	isCritical := 0
	for key, value := range c.commitCounts {
		if value > 1+2*c.view.ByzantineTolerance() {
			c.admitBlock(iosbase.Base58Decode(key))  // if the block is local, ok; otherwise request the correct block
			return StartPhase,nil
		}
		if value > c.view.ByzantineTolerance() {
			isCritical++
			if isCritical > 2 {
				// TODO critical situation
				return c.onTimeOut(CommitPhase)
			}
		}
	}
	return CommitPhase, nil
}

func (c *Consensus) onCommitFailed() (Phase, error) {
	c.makeEmptyBlock()
	return StartPhase, nil
}

func (c *Consensus) onTimeOut(phase Phase) (Phase, error) {
	switch phase {
	case PrePreparePhase :
		// if backup did not receive PPP pack, simply mark reject and goes to prepare phase
		c.prepare = c.makePrepare(false)
		return PreparePhase, nil
	case PreparePhase :
		// time out means offline committee members or divergence > t
		return c.onPrepareReject()
	case CommitPhase :
		//
		return c.onCommitFailed()
	}
	return StartPhase, nil
}

func (c *Consensus) makePrepare(isAccept bool) Prepare {
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

	pareSig := iosbase.Sign(vote, c.Seckey)
	prepare := Prepare{pareSig, c.Pubkey, randomByte, isAccept}

	return prepare
}

func (c *Consensus) makeCommit() Commit {
	var cc Commit
	cc.pubkey = c.Pubkey
	cc.blkHash = iosbase.Sha256(c.block.Encode())
	cc.sig = iosbase.Sign(cc.blkHash, c.Seckey)
	return cc
}

func (c *Consensus) replicaLoop() {
	c.phase = StartPhase
	var req iosbase.Request
	var err error = nil
	c.isRunning = true

	to := time.NewTimer(1 * time.Minute)

	for c.isRunning {

		switch c.phase {
		case StartPhase:
			v := NewDposView(c.blockChain)
			c.phase, err = c.onNewView(&v)
		case PanicPhase:
			return
		case EndPhase:
			return
		}

		if err != nil {
			fmt.Println(err)
		}

		select {
		case <-c.valiChan:
			req = <-c.valiChan

			switch c.phase {
			case PrePreparePhase:
				pp := PrePrepare{}
				pp.Unmarshal(req.Body)
				c.phase, err = c.onPrePrepare(&pp)
			case PreparePhase:
				p := Prepare{}
				p.Unmarshal(req.Body)
				c.phase, err = c.onPrepare(p)
			case CommitPhase:
				cm := Commit{}
				cm.Unmarshal(req.Body)
				c.phase, err = c.onCommit(cm)
			}

			if !to.Stop() {
				<-to.C
			}
			to.Reset(ExpireTime)
		case <-to.C:
			c.phase, err = c.onTimeOut(c.phase)
			if err != nil {
				return
			}
			to.Reset(ExpireTime)
		}
	}
}


