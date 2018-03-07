package pob

import "BlockChainFramework/src/iosbase"

type Vote struct {
	blk * iosbase.Block
	isAccept bool
	sig []byte
	pubkey []byte
}

type Validator struct {
	self *Client
}

func (v *Validator) VoteFor(isAccept bool, blk *iosbase.Block) (vote Vote) {
	vote.isAccept = isAccept
	vote.blk = blk
	var content []byte
	if isAccept {
		content = []byte{255}
	} else {
		content = []byte{0}
	}
	content = append(content, (*blk).SelfHash()...)
	vote.sig = iosbase.Sign(content, v.self.seckey)
	vote.pubkey = v.self.pubkey
	return
}

func (v *Validator) OnReceiveBlock(blk iosbase.Block) {

}

func (v *Validator) OnReceiveVote(vote Vote) {

}

