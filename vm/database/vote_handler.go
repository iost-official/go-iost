package database

import (
	"github.com/bitly/go-simplejson"
	"github.com/iost-official/go-iost/common"
)

// VoteProducerContractName name of vote producer contract
const VoteProducerContractName = "vote_producer.iost"

// VoteContractName name of common vote contract
const VoteContractName = "vote.iost"

// VoteHandler easy to get info of vote.iost and vote_producer.iost
type VoteHandler struct {
	BasicHandler
	MapHandler
}

// AccountVoteInfo ...
type AccountVoteInfo struct {
	Option       string
	Votes        string
	ClearedVotes string
}

// GetAccountVoteInfo ...
func (v *VoteHandler) GetAccountVoteInfo(account string) []*AccountVoteInfo {
	idVal := v.Get(VoteProducerContractName + "-voteId")
	voteID, ok := Unmarshal(idVal).(string)
	if !ok {
		return nil
	}
	userVoteVal := v.MGet(VoteContractName+"-u_"+voteID, account)
	userVoteStr, ok := Unmarshal(userVoteVal).(string)
	if !ok {
		return nil
	}
	userVote, err := simplejson.NewJson([]byte(userVoteStr))
	if err != nil {
		return nil
	}
	result := []*AccountVoteInfo{}
	for pro := range userVote.MustMap() {
		votes, err := common.NewFixed(userVote.Get(pro).GetIndex(0).MustString(), 8)
		if err != nil {
			return nil
		}
		cleared, err := common.NewFixed(userVote.Get(pro).GetIndex(2).MustString(), 8)
		if err != nil {
			return nil
		}
		result = append(result, &AccountVoteInfo{
			Option:       pro,
			Votes:        votes.Sub(cleared).ToString(),
			ClearedVotes: cleared.ToString(),
		})
	}
	return result
}
