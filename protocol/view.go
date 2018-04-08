package protocol

import (
	"fmt"
	"sort"

	"github.com/iost-official/Go-IOS-Protocol/iosbase"
)

//go:generate mockgen -destination mocks/mock_view.go -package protocol_mock github.com/iost-official/Go-IOS-Protocol/protocol View

/*
Information of PBFT committee members
*/
type View interface {
	Init(chain iosbase.BlockChain)

	GetPrimary() iosbase.Member
	GetBackup() []iosbase.Member
	IsPrimary(ID string) bool
	IsBackup(ID string) bool
	CommitteeSize() int
	ByzantineTolerance() int
}

func ViewFactory(target string) (View, error) {
	switch target {
	case "pob":
		return &PobView{}, nil
	}
	return nil, fmt.Errorf("target view not found")
}

/*
view determined by the preview block's recorder, the best come into committee, and the preview
primary leaves
*/
type PobView struct {
	primary iosbase.Member
	backup  []iosbase.Member
}

func (v *PobView) Init(chain iosbase.BlockChain) {

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

	var txpool iosbase.TxPool
	txpool.Decode(blk.Content)
	txs, _ := txpool.GetSlice()

	candidateMap := make(map[string]int)

	for _, tx := range txs {
		candidateMap[tx.Recorder]++
	}

	var candidates candidateSlice

	for key, val := range candidateMap {
		candidates = append(candidates, candidate{key, val})
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

func (v *PobView) GetPrimary() iosbase.Member {
	return v.primary
}

func (v *PobView) GetBackup() []iosbase.Member {
	return v.backup
}

func (v *PobView) IsPrimary(ID string) bool {
	if ID == v.primary.ID {
		return true
	} else {
		return false
	}
}

func (v *PobView) IsBackup(ID string) bool {
	ans := false
	for _, m := range v.backup {
		if ID == m.ID {
			ans = true
		}
	}
	return ans
}

func (v *PobView) CommitteeSize() int {
	return len(v.backup) + 1
}

func (v *PobView) ByzantineTolerance() int {
	return len(v.backup) / 3
}

func isEmptyBlock(block *iosbase.Block) bool {
	return block.Content == nil
}

func parseCoinbaseTx(tx iosbase.Tx) (primary iosbase.Member, backups []iosbase.Member, err error) {
	return iosbase.Member{}, nil, nil
}

// struct for sorting members
type candidate struct {
	ID   string
	Vote int
}

type candidateSlice []candidate

func (a candidateSlice) Len() int {
	return len(a)
}
func (a candidateSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a candidateSlice) Less(i, j int) bool {
	return a[j].Vote < a[i].Vote
}
