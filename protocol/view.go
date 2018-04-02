package protocol

import (
	"sort"
	"github.com/iost-official/PrototypeWorks/iosbase"
	"fmt"
)

//go:generate mockgen -destination view_mock_test.go -package protocol -source view.go

type View interface {
	Init(chain iosbase.BlockChain)

	GetPrimary() iosbase.Member
	GetBackup() []iosbase.Member
	IsPrimary(ID string) bool
	IsBackup(ID string) bool
	CommitteeSize() int
	ByzantineTolerance() int
}


func ViewFactory(target string) (View, error){
	switch target {
	case "dpos":
		return &DposView{}, nil
	}
	return nil, fmt.Errorf("target view not found")
}


type DposView struct {
	primary iosbase.Member
	backup  []iosbase.Member
}

func (v *DposView)Init(chain iosbase.BlockChain) {

	/*
		ruler:
		1. the backups become primary in turn
		2. primary make block, and leave
		3. the best recorder become the newest backup
		4. the rulers in bad situations，simply put empty block，thus the member will walk though as the empty block on the top
	*/

	top := chain.Length()
	emptyCount := 0

	top--
	blk, err := chain.Get(top)
	if err != nil {
		panic(err) // TODO remove panic
	}

	for isEmptyBlock(blk) {
		emptyCount++
		top--
		blk, err = chain.Get(top)
		if err != nil {
			panic(err) // TODO remove panic
		}
	}

	txpool, _ := iosbase.FindTxPool(blk.Content)
	txs, _ := txpool.GetSlice()

	candidateMap := make(map[string]int)

	for _, tx := range txs {
		candidateMap[tx.Recorder]++
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
	v.primary = oldbus[0]
	v.backup = append(oldbus[1+emptyCount:], winners...)
}

func (v *DposView) GetPrimary() iosbase.Member {
	return v.primary
}

func (v *DposView) GetBackup() []iosbase.Member {
	return v.backup
}

func (v *DposView) IsPrimary(ID string) bool {
	if ID == v.primary.ID {
		return true
	} else {
		return false
	}
}

func (v *DposView) IsBackup(ID string) bool {
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
	return block.Content == nil
}

func parseCoinbaseTx(tx iosbase.Tx) (primary iosbase.Member, backups []iosbase.Member, err error) {
	return iosbase.Member{}, nil, nil
}

type Candidate struct {
	ID   string
	Vote int
}

type CandidateSlice []Candidate

func (a CandidateSlice) Len() int {
	return len(a)
}
func (a CandidateSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a CandidateSlice) Less(i, j int) bool {
	return a[j].Vote < a[i].Vote
}
