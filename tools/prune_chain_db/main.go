package main

import (
	"fmt"
	"time"

	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func prune_state_db() error {
	from := "/data/iserver/storage/StateDB"
	to := "./storage/StateDB"
	db, err := leveldb.OpenFile(from, nil)
	if err != nil {
		return err
	}
	defer func() {
		db.Close()
	}()

	db2, err := leveldb.OpenFile(to, nil)
	if err != nil {
		return err
	}

	range1 := &util.Range{Start: nil, Limit: []byte("state/b-base.iost-chain_infn")}
	iter1 := db.NewIterator(range1, nil)
	for iter1.Next() {
		db2.Put(iter1.Key(), iter1.Value(), nil)
	}
	iter1.Release()
	err = iter1.Error()
	if err != nil {
		return err
	}

	range2 := &util.Range{Start: []byte("state/b-base.iost-chain_infp"), Limit: nil}
	iter2 := db.NewIterator(range2, nil)
	for iter2.Next() {
		db2.Put(iter2.Key(), iter2.Value(), nil)
	}
	iter2.Release()
	err = iter2.Error()
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second)
	err = db2.Close()
	if err != nil {
		return err
	}
	db2, err = leveldb.OpenFile(to, nil)
	defer func() {
		db2.Close()
	}()
	if err != nil {
		return err
	}
	fmt.Println("prune state done")
	return nil
}

func prune_histoy_db() {
	from := "/data/iserver/storage/BlockChainDB"
	to := "./storage/BlockChainDB"
	chainDB, err := block.NewBlockChain(from)
	if err != nil {
		fmt.Println("cannot load chain", err)
		return
	}
	fmt.Println("start trim from", from, "to", to)
	err = chainDB.(*block.BlockChain).CopyLastNBlockTo(to, 1)
	if err != nil {
		fmt.Println("cannot write chain", err)
		return
	}
	fmt.Println("trim chain done")
}

func main() {
	prune_histoy_db()
	err := prune_state_db()
	if err != nil {
		panic(err)
	}
}
