package merkletree

import (
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
)

func (m *TXRMerkleTree) Build(txrs []*tx.TxReceipt) {
	m.Mt = &MerkleTree{}
	data := make([][]byte, len(txrs))
	m.Tx2Txr = make(map[string][]byte)
	for i, txr := range txrs {
		m.Tx2Txr[string(txr.TxHash)] = txr.Encode()
		data[i] = m.Tx2Txr[string(txr.TxHash)]
	}
	m.Mt.Build(data)
}

func (m *TXRMerkleTree) GetTXR(hash []byte) (*tx.TxReceipt, error) {
	txr := tx.TxReceipt{}
	txrHash, ok := m.Tx2Txr[string(hash)]
	if !ok {
		return nil, errors.New("txHash isn't in the tree")
	}
	err := txr.Decode(txrHash)
	if err != nil {
		return nil, err
	}
	return &txr, nil
}
func (m *TXRMerkleTree) RootHash() []byte {
	return m.Mt.RootHash()
}

func (m *TXRMerkleTree) MerklePath(hash []byte) ([][]byte, error) {
	return m.Mt.MerklePath(hash)
}

func (m *TXRMerkleTree) MerkleProve(hash []byte, rootHash []byte, mp [][]byte) (bool, error) {
	return m.Mt.MerkleProve(hash, rootHash, mp)
}

func (m *TXRMerkleTree) Encode() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *TXRMerkleTree) Decode(b []byte) error {
	return proto.Unmarshal(b, m)
}
