package pob

import (
	. "github.com/iost-official/Go-IOS-Protocol/account"
	. "github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	. "github.com/iost-official/Go-IOS-Protocol/new_consensus/common"

	"errors"
	"fmt"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"

	"encoding/binary"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/vm"
	"github.com/iost-official/Go-IOS-Protocol/vm/lua"
	"time"
	"github.com/iost-official/Go-IOS-Protocol/new_vm"
)

func genGenesis(initTime int64) (*block.Block, error) {

	main := lua.NewMethod(vm.Public, "", 0, 0)

	var code string
	for k, v := range GenesisAccount {
		code += fmt.Sprintf("@PutHM iost %v f%v\n", k, v)
	}

	lc := lua.NewContract(vm.ContractInfo{Prefix: "", GasLimit: 0, Price: 0, Publisher: ""}, code, main)

	tx := Tx{
		Time:     0,
		Nonce:    0,
		Contract: &lc,
	}

	genesis := &block.Block{
		Head: block.BlockHead{
			Version: 0,
			Number:  0,
			Time:    initTime,
		},
		Txs:      make([]Tx, 0),
		Receipts: make([]TxReceipt, 0),
	}
	genesis.Txs = append(genesis.Txs, tx)
	return genesis, nil
}

func genBlock(acc Account, node *blockcache.BlockCacheNode, db *db.MVCCDB) *block.Block {
	lastBlk := node.Block
	blk := block.Block{
		Head: block.BlockHead{
			Version:    0,
			ParentHash: lastBlk.HeadHash(),
			Number:     lastBlk.Head.Number + 1,
			Witness:    acc.ID,
			Time:       GetCurrentTimestamp().Slot,
		},
		Txs:      []Tx{},
		Receipts: []TxReceipt{},
	}

	txCnt := 1000
	limitTime := time.NewTicker(((SlotLength/3 - 1) + 1) * time.Second)
	if txpool.TxPoolS != nil {
		tx, err := txpool.TxPoolS.PendingTransactions(txCnt)
		if err == nil {
			txPoolSize.Set(float64(txpool.TxPoolS.TransactionNum()))

			if len(tx) != 0 {
				//  is this necessary?
				db.Fork()
				VerifyTxBegin(lastBlk, db)
			ForEnd:
				for _, t := range tx {
					select {
					case <-limitTime.C:
						break ForEnd
					default:
						if len(blk.Txs) >= txCnt {
							break ForEnd
						}
						if receipt, err := VerifyTx(t); err == nil {
							// same question on commit
							db.Commit()
							blk.Txs = append(blk.Txs, *t)
							blk.Receipts = append(blk.Receipts, receipt)
						} else {
							db.Rollback()
						}
					}
				}
			}
		}
	}

	blk.Head.TxsHash = blk.CalculateTxsHash()
	blk.Head.MerkleHash = blk.CalculateMerkleHash()
	headInfo := generateHeadInfo(blk.Head)
	sig, _ := common.Sign(common.Secp256k1, headInfo, acc.Seckey)
	blk.Head.Signature = sig.Encode()
	// commit here?
	db.Commit(blk.HeadHash())

	generatedBlockCount.Inc()

	return &blk
}

func generateHeadInfo(head block.BlockHead) []byte {
	var info, numberInfo, versionInfo []byte
	info = make([]byte, 8)
	versionInfo = make([]byte, 4)
	numberInfo = make([]byte, 4)
	binary.BigEndian.PutUint64(info, uint64(head.Time))
	binary.BigEndian.PutUint32(versionInfo, uint32(head.Version))
	binary.BigEndian.PutUint32(numberInfo, uint32(head.Number))
	info = append(info, versionInfo...)
	info = append(info, numberInfo...)
	info = append(info, head.ParentHash...)
	info = append(info, head.TxsHash...)
	info = append(info, head.MerkleHash...)
	info = append(info, head.Info...)
	return common.Sha256(info)
}

func verifyBasics(blk *block.Block, parent *block.Block) error {
	// verify block head
	if err := VerifyBlockHead(blk, parent); err != nil {
		return err
	}

	// verify block witness
	if witnessOfTime(Timestamp{Slot: blk.Head.Time}) != blk.Head.Witness {
		return errors.New("wrong witness")
	}

	headInfo := generateHeadInfo(blk.Head)
	var signature common.Signature
	signature.Decode(blk.Head.Signature)

	if blk.Head.Witness != common.Base58Encode(signature.Pubkey) {
		return errors.New("wrong pubkey")
	}

	// verify block witness signature
	if !common.VerifySignature(headInfo, signature) {
		return errors.New("wrong signature")
	}

	// verify slot map
	if staticProp.hasSlotWitness(uint64(blk.Head.Time), blk.Head.Witness) {
		return errors.New("witness slot duplicate")
	}

	// verify exist txs
	if err := txpool.TxPoolS.ExistTxs(blk.HeadHash(), blk); err {
		return errors.New("duplicate txs")
	}

	return nil
}

func verifyBlockTxs(blk *block.Block, db *db.MVCCDB) error {
	// verify txs
	err := VerifyBlock(blk, db)
	if err != nil {
		return err
	}
	return nil
}

func updateNodeInfo(node *blockcache.BlockCacheNode) {
	node.Number = node.Block.Head.Number
	node.Witness = node.Block.Head.Witness

	// watermark
	node.ConfirmUntil = staticProp.Watermark[node.Witness]
	staticProp.Watermark[node.Witness] = node.Number + 1

	// slot map
	staticProp.addSlotWitness(uint64(node.Block.Head.Time), node.Witness)
}

func updateWitness(node *blockcache.BlockCacheNode, db *db.MVCCDB) []string {
	// pending witness
	newList := db.Get("witnessList")

	if newList != nil {
		node.PendingWitnessList = newList
		node.LastWitnessListNumber = node.Number
	} else {
		node.PendingWitnessList = node.Parent.PendingWitnessList
		node.LastWitnessListNumber = node.Parent.LastWitnessListNumber
	}
}

func calculateConfirm(node *blockcache.BlockCacheNode) *blockcache.BlockCacheNode {
	// return the last number that confirmed
	confirmNumber := staticProp.NumberOfWitnesses*2/3 + 1
	totLen := node.Number - block.Chain.Length()
	confirmMap := make(map[string]int)
	confirmUntil := make([][]string, totLen)
	i := 0
	for node != nil {
		if node.ConfirmUntil < node.Number {
			if num, err := confirmMap[node.Witness]; err {
				confirmMap[node.Witness] = 1
			} else {
				confirmMap[node.Witness] = num + 1
			}
		}
		index := node.Number - node.ConfirmUntil
		confirmUntil[index] = append(confirmUntil[index], node.Witness)
		if len(confirmMap) >= confirmNumber {
			staticProp.delSlotWitness(block.Chain.Length(), node.Number)
			return node
		}
		if confirmUntil[i] != nil {
			for j := range confirmUntil[i] {
				confirmMap[confirmUntil[i][j]]--
			}
		}
		node = node.Parent
	}
}

func promoteWitness(node *blockcache.BlockCacheNode, confirmed uint64) {
	// update the last pending witness list that has been confirmed
	for node != nil && node.LastWitnessListNumber > confirmed {
		node = node.Parent
	}
	if node != nil {
		staticProp.updateWitnessList(node.PendingWitnessList)
	}
}
