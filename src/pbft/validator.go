package pbft

import (
	"IOS/src/iosbase"
	//"fmt"
	"IOS/src/network"
	"math/rand"
	"time"
)

type ValidatorImpl struct {
	iosbase.Member
	iosbase.Replica

	character Character
	view      *View

	prePrepare   *PrePrepare
	prepare      *Prepare
	prepareCount int
	commit       *Commit
	commitCount  int
}

func (r *ValidatorImpl) OnNewView() (Phase, error) {
	// step 1 determine what character it is
	view := NewView(&r.BlockChain)
	r.view = &view
	if view.GetPrimaryID() == r.ID {
		r.character = Primary
	} else {
		r.character = Idle
		for _, rep := range r.view.GetBackupID() {
			if rep == r.ID {
				r.character = Backup
			}
		}
	}

	if r.character == Backup {
		return PreparePhase, nil
	} else if r.character == Idle {
		return EndPhase, nil
	}

	// step 2 if it is primary, make a block/pre-prepare package and broadcast it
	bBlk := r.MakeBlock().Bytes()

	sig := iosbase.Sign(iosbase.Sha256(bBlk), r.Seckey)

	pre := PrePrepare{sig, r.Pubkey, bBlk}

	bPre, err := pre.Marshal(nil)
	if err != nil {
		return PanicPhase, err
	}

	pn := network.PbftNetwork{}
	for _, m := range view.backup {
		pn.SendSync(network.PbftRequest{r.ID, m.ID, network.PBFT_PrePrepare, bPre, time.Now().Unix()})
	}

	// step 3, as primary, prepare a Prepare pack which is always true
	prepare := r.MakePrepare(true)
	r.prepare = &prepare

	return PreparePhase, nil
}

func (r *ValidatorImpl) OnPrePrepare(prePrepare PrePrepare) (Phase, error) {
	// 1. verify block syntax
	// 2. verify if txs contains tx which validated sign conflict
	// if ok, r.prepare is true, vise versa

	// 3. broadcast prepare
	bPare, err := r.prepare.Marshal(nil)
	if err != nil {
		return PrePreparePhase, err
	}

	pn := network.PbftNetwork{}

	pn.SendSync(network.PbftRequest{r.ID, r.view.GetPrimaryID(), network.PBFT_Prepare, bPare, time.Now().Unix()})
	for _, mID := range r.view.GetBackupID() {
		if mID == r.ID {
			continue
		}
		pn.SendSync(network.PbftRequest{r.ID, mID, network.PBFT_Prepare, bPare, time.Now().Unix()})
	}
	return PreparePhase, nil

}

func (r *ValidatorImpl) OnPrepare(prepare Prepare) (Phase, error) {
	// 1. verify prepare Sign, a/ comes from right member, b/ have different random bytes
	// 2. if 2t+1 true, go to commit, otherwise waiting for a new pre-prepare, after 3 pre-prepare go to next view
	return CommitPhase, nil
}

func (r *ValidatorImpl) OnCommit(commit Commit) (Phase, error) {
	return StartPhase, nil
}

func (r *ValidatorImpl) MakePrepare(isAccept bool) Prepare {
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
	prepare := Prepare{pareSig, r.Pubkey, isAccept}

	return prepare
}

func (r *ValidatorImpl) MakeBlock() iosbase.Block {
	return nil
}
