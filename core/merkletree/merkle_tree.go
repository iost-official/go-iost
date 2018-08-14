package merkletree

import (
	"bytes"
	"crypto/sha256"
	"errors"

	"math"
)

func (m *MerkleTree) Build(data [][]byte) {
	if len(data) == 0 {
		m.HashList[0] = nil
		return
	}
	n := int32(math.Exp2(math.Ceil(math.Log2(float64(len(data))))))
	m.LeafNum = n
	m.HashList = make([][]byte, 2*n)
	copy(m.HashList[n-1:n+int32(len(data))-1], data)
	start := n - 1
	end := n + int32(len(data)) - 2
	for {
		for i := start; i <= end; i = i + 2 {
			var tmpHash [32]byte
			if m.HashList[i+1] == nil {
				tmpHash = sha256.Sum256(append(m.HashList[i], m.HashList[i]...))
			} else {
				if bytes.Compare(m.HashList[i], m.HashList[i+1]) < 0 {
					tmpHash = sha256.Sum256(append(m.HashList[i], m.HashList[i+1]...))
				} else {
					tmpHash = sha256.Sum256(append(m.HashList[i+1], m.HashList[i]...))
				}
			}
			p := (i - 1) / 2
			m.HashList[p] = tmpHash[:]
		}
		start = (start - 1) / 2
		end = (end - 1) / 2
		if start == end {
			break
		}
	}
	m.Hash2Idx = make(map[string]int32)
	for idx, datum := range data {
		m.Hash2Idx[string(datum)] = int32(idx) + n - 1
	}
}

func (m *MerkleTree) RootHash() []byte {
	return m.HashList[0]
}

func (m *MerkleTree) MerklePath(hash []byte) ([][]byte, error) {
	if m.LeafNum == 0 {
		return nil, errors.New("merkletree hasn't built")
	}
	idx, ok := m.Hash2Idx[string(hash)]
	if !ok {
		return nil, errors.New("hash isn't in the tree")
	}
	mp := make([][]byte, int32(math.Log2(float64(m.LeafNum))))
	for i := 0; idx != 0; i++ {
		p := (idx - 1) / 2
		if m.HashList[4*p+3-idx] == nil { // p, 2p+1, 2p+2
			mp[i] = m.HashList[idx]
		} else {
			mp[i] = m.HashList[4*p+3-idx]
		}
		idx = p
	}
	return mp, nil
}

func (m *MerkleTree) MerkleProve(hash []byte, rootHash []byte, mp [][]byte) (bool, error) {
	if hash == nil {
		return false, errors.New("hash input error")
	}
	if rootHash == nil {
		return false, errors.New("rootHash input error")
	}
	for _, p := range mp {
		var tmpHash [32]byte
		if bytes.Compare(hash, p) < 0 {
			tmpHash = sha256.Sum256(append(hash, p...))
		} else {
			tmpHash = sha256.Sum256(append(p, hash...))
		}
		hash = tmpHash[:]
	}
	return bytes.Equal(hash, rootHash), nil
}
