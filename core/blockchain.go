package core

//go:generate mockgen -destination mocks/mock_blockchain.go -package core_mock github.com/iost-official/prototype/core BlockChain

// Block chain
type BlockChain interface {
	Push(block *Block) error // 加入block，检查block是否合法在consensus内实现以解耦合
	Length() int
	Top() *Block // 语法糖

	Iterator() BlockChainIterator
}

type BlockChainIterator interface {
	Next() *Block // 返回下一个块
}

//type BlockChainImpl struct {
//	db     *iostdb.LDBDatabase
//	redis  redis.Conn
//	length int
//}
//
//const (
//	DBPath   = "savedata/"
//	IndexKey = "block_chain_index"
//)
//
//func (bc *BlockChainImpl) Get(layer int) (*Block, error) {
//
//	if layer < 0 || layer >= bc.length {
//		return nil, fmt.Errorf("index exceed")
//	}
//
//	headHash, err := redis.Bytes(bc.redis.Do("LINDEX", IndexKey, layer))
//	if err != nil {
//		return nil, err
//	}
//
//	blk, err := bc.db.Get(headHash)
//	if err != nil {
//		return nil, err
//	}
//	var block Block
//	block.Decode(blk)
//	return &block, nil
//}
//
//func (bc *BlockChainImpl) Push(block *Block) error {
//	err := bc.db.Put(block.HeadHash(), block.Encode())
//	if err != nil {
//		return err
//	}
//
//	_, err = bc.redis.Do("RPUSH", IndexKey, block.HeadHash())
//	bc.length++
//	return nil
//}
//
//func (bc *BlockChainImpl) Length() int {
//	return bc.length
//}
//
//func (bc *BlockChainImpl) Top() *Block {
//	blk, err := bc.Get(bc.length - 1)
//	if err != nil {
//		panic(err)
//	}
//	return blk
//}
//
//func (bc *BlockChainImpl) Init() error {
//	var err error
//	bc.db, err = iostdb.NewLDBDatabase(DBPath, 1, 1)
//	if err != nil {
//		return err
//	}
//
//	bc.redis, err = redis.Dial(Conn, DBAddr)
//	if err != nil {
//		return err
//	}
//
//	len, err := redis.Int(bc.redis.Do("llen", "BC_index"))
//	if err != nil {
//		return err
//	}
//	bc.length = len
//	return nil
//}
//
//func (bc *BlockChainImpl) Close() error {
//	bc.db.Close()
//	return nil
//}
