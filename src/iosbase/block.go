package iosbase

type Block interface {
	SelfHash() []byte
	SuperHash() []byte
	ContentHash() []byte

	Verify(chain *BlockChain, pool *StatePool) bool
	Bytes() []byte
}
