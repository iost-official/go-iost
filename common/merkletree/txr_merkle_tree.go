package merkletree

import (
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"fmt"
	"encoding/hex"
)

func (m *TXRMerkleTree) Build(txrs []tx.TxReceipt) error {
	m.MT = &MerkleTree{}
	data := make([][]byte, len(txrs))
	m.TX2TXR = make(map[string][]byte)
	for i, txr := range txrs {
		m.TX2TXR[string(txr.TxHash)] = txr.Encode()
		fmt.Printf("%s\n", hex.EncodeToString(txr.Encode()))
		data[i] = m.TX2TXR[string(txr.TxHash)]
	}
	return m.MT.Build(data)
}

func (m *TXRMerkleTree) GetTXR(hash []byte) (tx.TxReceipt, error) {
	txr := tx.TxReceipt{}
	txr_hash := m.TX2TXR[string(hash)]
	fmt.Printf("%s", hex.EncodeToString(txr_hash))
	//if err != nil {
	//	return txr, err
	//}
	return txr, nil
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


