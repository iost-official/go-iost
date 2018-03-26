package protocol

import (
	"IOS/src/iosbase"
	"math/rand"
	"time"
)

type Validator struct {
	character Character
	view      View

	prePrepare   *PrePrepare
	prepare      *Prepare
	prepareCount int
	commit       *Commit
	commitCount  int
}

func (c *Consensus) onNewView() (Phase, error) {
	// step 1 determine what character it is
	view := NewDposView(c.blockChain)
	c.view = &view
	if c.view.isPrimary(c.ID) {
		c.character = Primary
	} else if c.view.isBackup(c.ID){
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
	bBlk := c.makeBlock().Encode()

	sig := iosbase.Sign(iosbase.Sha256(bBlk), c.Seckey)

	pre := PrePrepare{sig, c.Pubkey, bBlk}

	bPre, err := pre.Marshal(nil)
	if err != nil {
		return PanicPhase, err
	}

	for _, m := range view.backup {
		c.send(iosbase.Request{time.Now().Unix(),c.ID, m.ID, int(PrePreparePhase), bPre})
	}

	// step 3, as primary, prepare a Prepare pack which is always true
	prepare := c.makePrepare(true)
	c.prepare = &prepare

	return PreparePhase, nil
}

func (c *Consensus) onPrePrepare(prePrepare PrePrepare) (Phase, error) {
	// 1. verify block syntax

	// 2. verify if txs contains tx which validated sign conflict
	// if ok, c.prepare is true, vise versa

	// 3. broadcast prepare
	bPare, err := c.prepare.Marshal(nil)
	if err != nil {
		return PrePreparePhase, err
	}


	c.send(iosbase.Request{time.Now().Unix(),c.ID, c.view.GetPrimaryID(), int(PreparePhase), bPare})
	for _, mID := range c.view.GetBackupID() {
		if mID == c.ID {
			continue
		}
		c.send(iosbase.Request{ time.Now().Unix(),c.ID, mID, int(PreparePhase), bPare})
	}
	return PreparePhase, nil

}

func (c *Consensus) onPrepare(prepare Prepare) (Phase, error) {
	// 1. verify prepare Sign, a/ comes from right member, b/ have different random bytes
	// 2. if 2t+1 true, go to commit, otherwise waiting for a new pre-prepare, after 3 pre-prepare go to next view
	return CommitPhase, nil
}

func (c *Consensus) onCommit(commit Commit) (Phase, error) {
	return StartPhase, nil
}

func (c *Consensus) onTimeOut() (Phase, error) {
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
	prepare := Prepare{pareSig, c.Pubkey, isAccept}

	return prepare
}

func (c *Consensus) makeBlock() iosbase.Block {
	return iosbase.Block{}
}

