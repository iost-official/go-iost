package consensus

import "BlockChainFramework/src/iosbase"


type State int
const (
	Accepted State = iota
	Reject
	Error

)


func GetPrimary (chain iosbase.BlockChain) (primary iosbase.Member) {
	return primary
}

func GetBackupList (chain iosbase.BlockChain) (backups []iosbase.Member) {
	return backups
}

func GetView (chain iosbase.BlockChain) int {
	return 0
}

type PbftRequest struct {
	sender iosbase.Member
	blk iosbase.Block
	time uint32
}

type PbftResponse struct {
	receiver Replica
	view int
	time uint32
	result State
}

type PrePrepare struct {

}




type Replica struct {
	iosbase.Member
}

func (r * Replica) OnReceive(request PbftRequest) (PbftResponse, error) {
}




type Primary struct {
	iosbase.Member
}

func (p *Primary)