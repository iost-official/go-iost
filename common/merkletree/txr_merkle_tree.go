package merkletree

import (
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
)

type TXRMerkleTree struct {
	MT		MerkleTree
	TX2TXR	map[string]tx.TxReceipt
}

func (m *TXRMerkleTree) Build(txrs []tx.TxReceipt) error {
	data := make([][]byte, len(txrs))
	m.TX2TXR = make(map[string]tx.TxReceipt)
	for i, txr := range txrs {
		m.TX2TXR[string(txr.TxHash)] = txr
		data[i] = txr.Hash()
	}
	m.MT.Build(data)
	return nil
}

func (m *TXRMerkleTree) RootHash() ([]byte, error) {
	return m.MT.RootHash()
}

func (m *TXRMerkleTree) MerklePath(hash []byte) ([][]byte, error) {
	return m.MT.MerklePath(hash)
}

func (m *TXRMerkleTree) MerkleProve(hash []byte, rootHash []byte, mp [][]byte) (bool, error) {
	return m.MT.MerkleProve(hash, rootHash, mp)
}


