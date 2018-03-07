package pob

import "BlockChainFramework/src/iosbase"

type Client struct {
	blockChain *iosbase.BlockChain

	statePool iosbase.StatePool

	id []byte
	pubkey []byte
	seckey []byte

	clients map[Client]ClientState
	validators map[Validator]ClientState
	leader * Leader
}

func (client *Client) OnReceiveBlock(block iosbase.Block) {

}

func (client *Client) MainLoop() {

}

