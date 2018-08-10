package merkletree

import (
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/golang/protobuf/proto"
	"errors"
)

func (m *TXRMerkleTree) Build(txrs []*tx.TxReceipt) error {
	m.MT = &MerkleTree{}
	data := make([][]byte, len(txrs))
	m.TX2TXR = make(map[string][]byte)
	for i, txr := range txrs {
		m.TX2TXR[string(txr.TxHash)] = txr.Encode()
		data[i] = m.TX2TXR[string(txr.TxHash)]
	}
	return m.MT.Build(data)
}

func (m *TXRMerkleTree) GetTXR(hash []byte) (*tx.TxReceipt, error) {
	txr := tx.TxReceipt{}
	txrHash, ok := m.TX2TXR[string(hash)]
	if !ok {
		return nil, errors.New("txHash isn't in the tree")
	}
	err := txr.Decode(txrHash)
	if err != nil {
		return nil, err
	}
	return &txr, nil
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

func (m *TXRMerkleTree) Encode() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *TXRMerkleTree) Decode(b []byte) error {
	return proto.Unmarshal(b, m)
}


