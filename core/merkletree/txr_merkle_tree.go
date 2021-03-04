package merkletree

import (
	"encoding/hex"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/v3/core/tx"
)

// Build return the merkle tree
func (m *TXRMerkleTree) Build(txrs []*tx.TxReceipt) {
	m.Mt = &MerkleTree{}
	data := make([][]byte, len(txrs))
	m.Tx2Txr = make(map[string][]byte)
	for i, txr := range txrs {
		k := hex.EncodeToString(txr.TxHash)
		m.Tx2Txr[k] = txr.Hash()
		data[i] = m.Tx2Txr[k]
	}
	m.Mt.Build(data)
}

// RootHash return root of merkle tree
func (m *TXRMerkleTree) RootHash() []byte {
	return m.Mt.RootHash()
}

// MerklePath return path of the merkle tree
func (m *TXRMerkleTree) MerklePath(hash []byte) ([][]byte, error) {
	return m.Mt.MerklePath(hash)
}

// MerkleProve return prove of the merkle tree
func (m *TXRMerkleTree) MerkleProve(hash []byte, rootHash []byte, mp [][]byte) (bool, error) {
	//return m.Mt.MerkleProve(hash, rootHash, mp)
	return false, nil
}

// Encode is marshal of the merkle tree
func (m *TXRMerkleTree) Encode() ([]byte, error) {
	return proto.Marshal(m)
}

// Decode is unmarshal of the merkle tree
func (m *TXRMerkleTree) Decode(b []byte) error {
	return proto.Unmarshal(b, m)
}
