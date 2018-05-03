package dpos

import (
	"errors"
	"sort"

	"bytes"
	"github.com/iost-official/prototype/core"
	"strings"
	"time"
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/core/message"
)

const (
	// 测试用常量
	ReqTypeVoteTest = 100
	// 维护周期的长度，以小时为单位
	MaintenanceInterval = 24
)

func (p *DPoS) VoteForWitness(voter account.Account, witnessId string, voteType bool) {
	//应该生成一个交易并发送，测试版本中简单设置为广播一个消息，后续再对接
	var reqString string
	if voteType {
		reqString = "Vote For " + voter.GetId() + " " + witnessId
	} else {
		reqString = "Vote Against " + voter.GetId() + " " + witnessId
	}
	req := message.Message{
		Time: time.Now().Unix(),
		From: voter.GetId(),
		//To:      p.DPoSSuperMember,
		To:      "ALL",
		ReqType: ReqTypeVoteTest,
		Body:    []byte(reqString),
	}
	p.Router.Send(req)
}

func (p *DPoS) WitnessJoin(witness core.Member) {
	//应该生成一个交易并发送，测试版本中简单设置为广播一个消息，后续再对接
	reqString := "Join " + witness.GetId()
	req := core.Request{
		Time: time.Now().Unix(),
		From: witness.GetId(),
		//To:      p.DPoSSuperMember,
		To:      "ALL",
		ReqType: ReqTypeVoteTest,
		Body:    []byte(reqString),
	}
	p.Router.Send(req)
}

func (p *DPoS) WitnessQuit(witness core.Member) {
	//应该生成一个交易并发送，测试版本中简单设置为广播一个消息，后续再对接
	reqString := "Quit " + witness.GetId()
	req := core.Request{
		Time: time.Now().Unix(),
		From: witness.GetId(),
		//To:      p.DPoSSuperMember,
		To:      "ALL",
		ReqType: ReqTypeVoteTest,
		Body:    []byte(reqString),
	}
	p.Router.Send(req)
}

// 测试用函数：p2p收到ReqTypeVoteTest后调用，将消息加入到info的缓存中
// 在生成块时，将infoCache中的内容序列化后直接加入info，清空infoCache
func (p *DPoS) AddWitnessMsg(req core.Request) {
	if req.ReqType != ReqTypeVoteTest {
		return
	}
	for _, request := range p.infoCache {
		if bytes.Equal(request, req.Body) {
			return
		}
	}
	p.infoCache = append(p.infoCache, req.Body)
}

// 测试用函数：当块被确认，解码info中的相关消息更新投票状态
func (p *DPoS) ProcessWitnessTx(req []byte) error {
	reqStrings := strings.Split(string(req), " ")
	switch reqStrings[0] {
	case "Join":
		witness := reqStrings[1]
		return p.AddPendingWitness(witness)
	case "Quit":
		witness := reqStrings[1]
		return p.DeletePendingWitness(witness)
	case "Vote":
		if reqStrings[1] == "For" {
			return p.addVote(reqStrings[2], reqStrings[3])
		} else if reqStrings[1] == "Against" {
			return p.deleteVote(reqStrings[2], reqStrings[3])
		} else {
			return errors.New("illegal vote msg")
		}
	default:
		return errors.New("illegal msg")
	}
}

// 测试用函数，加入投票状态，正式版本中应该在运行合约时维护
func (p *DPoS) addVote(voter string, witness string) error {
	if votedList, ok := p.votedStats[voter]; ok {
		for _, wit := range votedList {
			if wit == witness {
				return errors.New("already voted")
			}
		}
		p.votedStats[voter] = append(votedList, witness)
	} else {
		p.votedStats[voter] = []string{witness}
	}
	return nil
}

// 测试用函数，删除投票状态，正式版本中应该在运行合约时维护
func (p *DPoS) deleteVote(voter string, witness string) error {
	if votedList, ok := p.votedStats[voter]; ok {
		i := 0
		for _, wit := range votedList {
			if wit == witness {
				p.votedStats[voter] = append(votedList[:i], votedList[i+1:]...)
				return nil
			}
			i++
		}
		return errors.New("never voted")
	} else {
		return errors.New("voter error")
	}
}

func (p *DPoS) PerformMaintenance() error {
	//Maintenance过程，主要进行投票结果统计并生成新的witness列表
	votes := make(map[string]int)
	// 测试用写法，原本应该从core.statspool中读取
	for _, votedList := range p.votedStats {
		for _, witness := range votedList {
			if inList(witness, p.WitnessList) || inList(witness, p.PendingWitnessList) {
				if value, ok := votes[witness]; ok {
					votes[witness] = value + 1
				} else {
					votes[witness] = 1
				}
			}
		}
	}
	if len(votes) < p.GlobalStaticProperty.NumberOfWitnesses {
		return errors.New("voted witnesses too few")
	}

	// choose the top NumberOfWitnesses witnesses and update lists
	witnessList := chooseTopN(votes, p.GlobalStaticProperty.NumberOfWitnesses)
	p.GlobalStaticProperty.UpdateWitnessLists(witnessList)

	// assume Add() adds a certain number into timestamp
	p.GlobalDynamicProperty.NextMaintenanceTime.AddHour(MaintenanceInterval)
	return nil
}

type Pair struct {
	Key   string
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
