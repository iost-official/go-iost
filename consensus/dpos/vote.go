package dpos

import (
	"github.com/iost-official/prototype/core"
	"github.com/iost-official/prototype/iostdb"
	"sort"
	"errors"
)

func (p *DPoS) PerformMaintenance() error {
	//维护过程，主要进行投票结果统计并生成新的witness列表
	// Lock to avoid new updates
	p.blockUpdateLock.Lock()
	defer p.blockUpdateLock.Unlock()

	votes := make(map[string]int)
	// Calculate votes from mem stats
	for _, mem := range iostdb.GetMemList() { //assume GetMemList returns a slice
		votedList := iostdb.GetVotedWitnessList(mem) //assume GetVotedWitnessList returns the voted list of a mem
		for _, witness := range votedList {
			// we don't judge if the witness is in lists, the proc is in stats updating
			if value, ok := votes[witness]; ok {
				votes[witness] = value + 1
			} else {
				votes[witness] = 1
			}
		}
	}
	if len(votes) < p.GlobalStaticProperty.NumberOfWitnesses {
		return errors.New("voted witnesses too few")
	}

	// choose the top NumberOfWitnesses witnesses and update lists
	witnessList := chooseTopN(votes, p.GlobalStaticProperty.NumberOfWitnesses)
	p.GlobalStaticProperty.UpdateWitnessLists(witnessList)

	// assume maintenance interval is defined in core/timestamp
	// assume Add() adds a certain number into timestamp
	p.GlobalDynamicProperty.NextMaintenanceTime.Add(core.MaintenanceInterval)
	return nil
}

type Pair struct {
	Key string
	Value int
}
type PairList []Pair

func (pl PairList) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}
func (pl PairList) Len() int {
	return len(pl)
}
func (pl PairList) Less(i, j int) bool {
	return pl[i].Value < pl[j].Value
}

func chooseTopN(votes map[string]int, num int) []string {
	var voteList PairList
	for k, v := range votes {
		voteList = append(voteList, Pair{k, v})
	}
	sort.Sort(voteList)
	list := make([]string, num)
	for i := 0; i < num; i++ {
		list[i] = voteList[i].Key
	}
	return list
}

