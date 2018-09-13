package merkletree

import (
	"encoding/binary"
	"errors"
	"sync"

	"github.com/iost-official/Go-IOS-Protocol/db"
)

type TXRMerkleTreeDB struct {
	txrMerkleTreeDB *db.LDB
}

var TXRMTDB TXRMerkleTreeDB

var once sync.Once

func Uint64ToByte(n uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)
	return b
}

func Init(LevelDBPath string) error {
	var err error
	once.Do(func() {
		levelDB, tempErr := db.NewLDB(LevelDBPath+"TXRMerkleTreeDB", 0, 0)
		if tempErr != nil {
			err = errors.New("fail to init TXRMerkleTreeDB")
		}
		TXRMTDB = TXRMerkleTreeDB{levelDB}
	})
	return err
}

func (mdb *TXRMerkleTreeDB) Put(m *TXRMerkleTree, blockNum uint64) error {
	mByte, err := m.Encode()
	if err != nil {
		return errors.New("fail to encode TXRMerkleTree")
	}
	err = mdb.txrMerkleTreeDB.Put(Uint64ToByte(blockNum), mByte)
	if err != nil {
		return errors.New("fail to put TXRMerkleTree")
	}
	return nil
}

func (mdb *TXRMerkleTreeDB) Get(blockNum uint64) (*TXRMerkleTree, error) {
	mByte, err := mdb.txrMerkleTreeDB.Get(Uint64ToByte(blockNum))
	m := TXRMerkleTree{}
	if err != nil {
		return nil, errors.New("fail to get TXRMerkleTree")
	}
	err = m.Decode(mByte)
	if err != nil {
		return nil, errors.New("fail to decode TXRMerkleTree")
	}
	return &m, nil
}
