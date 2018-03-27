package protocol

import (
	"IOS/src/iosbase"
	"sort"
)

type View interface {
	GetPrimary() iosbase.Member
	GetBackup() []iosbase.Member
	isPrimary(ID string) bool
	isBackup(ID string) bool
	CommitteeSize() int
	ByzantineTolerance() int
}

type DposView struct {
	primary iosbase.Member
	backup  []iosbase.Member
}

func NewDposView(chain iosbase.BlockChain) DposView {
	var view DposView

	/*
	ruler:
	1. the backups become primary in turn
	2. primary make block, and leave
	3. the best recorder become the newest backup
	4. the rulers in bad situations，simply put empty block，thus the member will walk though as the empty block on the top
	*/

	top := chain.Length()
	emptyCount := 0

	top --
	blk, err := chain.Get(top)
	if err != nil {
		panic(err) // TODO
	}

	for isEmptyBlock(blk) {
		emptyCount ++
		top --
		blk, err = chain.Get(top)
		if err != nil {
			panic(err) // TODO
		}
	}

	txpool, _ := iosbase.FindTxPool(blk.Content)
	txs, _ := txpool.GetSlice()

	candidateMap := make(map[string]int)

	for _, tx := range txs {
		candidateMap[tx.Recorder] ++
	}

	var candidates CandidateSlice

	for key, val := range candidateMap {
		candidates = append(candidates, Candidate{key, val})
	}

	sort.Sort(candidates)

	var winners = make([]iosbase.Member, 1+emptyCount, 1+emptyCount)
	for i, c := range candidates[:1+emptyCount] {
		winners[i] = iosbase.Member{ID: c.ID}
	}

	_, oldbus, _ := parseCoinbaseTx(txs[0])
	view.primary = oldbus[0]
	view.backup = append(oldbus[1+emptyCount:], winners...)

	return view
}

func (v *DposView) GetPrimary() iosbase.Member {
	return v.primary
}

func (v *DposView) GetBackup() []iosbase.Member {
	return v.backup
}

func (v *DposView) isPrimary(ID string) bool {
	if ID == v.primary.ID {
		return true
	} else {
		return false
	}
}

func (v *DposView) isBackup(ID string) bool {
	ans := false
	for _, m := range v.backup {
		if ID == m.ID {
			ans = true
		}
	}
	return ans
}

func (v *DposView) CommitteeSize() int {
	return len(v.backup) + 1
}

func (v *DposView) ByzantineTolerance() int {
	return len(v.backup) / 3
}

func isEmptyBlock(block iosbase.Block) bool {
	return false
}

func parseCoinbaseTx(tx iosbase.Tx) (primary iosbase.Member, backups []iosbase.Member, err error) {
	return iosbase.Member{}, nil, nil
}

type Candidate struct {
	ID   string
	Vote int
}

type CandidateSlice [] Candidate

func (a CandidateSlice) Len() int {
	return len(a)
}
func (a CandidateSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a CandidateSlice) Less(i, j int) bool {
	return a[j].Vote < a[i].Vote
}
