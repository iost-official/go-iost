package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func pruneStateDb(from, to string) error {
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

func pruneHistoryDb(from, to string, keepLastNBlock int) error {
	chainDB, err := block.NewBlockChain(from)
	if err != nil {
		fmt.Println("cannot load chain", err)
		return err
	}
	fmt.Println("start prune blockchain history: src", from, "dst", to)
	err = chainDB.(*block.BlockChain).CopyLastNBlockTo(to, (int64)(keepLastNBlock))
	if err != nil {
		fmt.Println("cannot write chain", err)
		return err
	}
	fmt.Println("prune blockchain history done")
	return nil
}

func ensureDir(fileName string) {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			panic(merr)
		}
	}
}

func main() {
	var from = flag.String("from", "/data/iserver/storage", "")
	var to = flag.String("to", "./storage", "")
	// TODO: keep block after block N
	var keepLastNBlock = flag.Int("keep", 10000, "")
	var pruneState = flag.Bool("pruneState", false, "")
	flag.Parse()

	ensureDir(*to)

	err := pruneHistoryDb(*from+"/BlockChainDB", *to+"/BlockChainDB", *keepLastNBlock)
	if err != nil {
		panic(err)
	}
	if *pruneState {
		err := pruneStateDb(*from+"/StateDB", *to+"/StateDB")
		if err != nil {
			panic(err)
		}
	}
}
