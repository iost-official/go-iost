package pob

import (
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/consensus/common"

	"encoding/binary"
	"errors"
	"time"

	"fmt"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
	"github.com/iost-official/Go-IOS-Protocol/db"
)

var (
	ErrWitness     = errors.New("wrong witness")
	ErrPubkey      = errors.New("wrong pubkey")
	ErrSignature   = errors.New("wrong signature")
	ErrSlotWitness = errors.New("witness slot duplicate")
	ErrTxTooOld    = errors.New("tx too old")
	ErrTxDup       = errors.New("duplicate tx")
	ErrTxSignature = errors.New("tx wrong signature")
)

func genBlock(account account.Account, node *blockcache.BlockCacheNode, txPool txpool.TxPool, db *db.MVCCDB) *block.Block {
	lastBlock := node.Block
	parentHash := lastBlock.HeadHash()
	blk := block.Block{
		Head: block.BlockHead{
			Version:    0,
			ParentHash: parentHash,
			Number:     lastBlock.Head.Number + 1,
			Witness:    account.ID,
			Time:       common.GetCurrentTimestamp().Slot,
		},
		Txs:      []*tx.Tx{},
		Receipts: []*tx.TxReceipt{},
	}
	txCnt := 1000
	limitTime := time.NewTicker((common.SlotLength/3 * time.Second))
	txsList, _ := txPool.PendingTxs(txCnt)
	txPoolSize.Set(float64(len(txsList)))
	if len(txsList) != 0 {
		consensus_common.VerifyTxBegin(lastBlock, db)
	ForEnd:
		for _, t := range txsList {
			select {
			case <-limitTime.C:
				break ForEnd
			default:
				if receipt, err := consensus_common.VerifyTx(t); err == nil {
					db.Commit()
					blk.Txs = append(blk.Txs, t)
					blk.Receipts = append(blk.Receipts, receipt)
				} else {
					db.Rollback()
				}
			}
		}
	}
	blk.Head.TxsHash = blk.CalculateTxsHash()
	blk.Head.MerkleHash = blk.CalculateMerkleHash()
	headInfo := generateHeadInfo(blk.Head)
	sig, _ := common.Sign(common.Secp256k1, headInfo, account.Seckey)
	blk.Head.Signature = sig.Encode()
	err := blk.CalculateHeadHash()
	if err != nil {
		fmt.Println(err)
	}
	db.Tag(string(blk.HeadHash()))

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

func verifyBasics(blk *block.Block) error {
	// verify block witness
	if witnessOfTime(common.Timestamp{Slot: blk.Head.Time}) != blk.Head.Witness {
		return ErrWitness
	}

	headInfo := generateHeadInfo(blk.Head)
	var signature common.Signature
	signature.Decode(blk.Head.Signature)

	if blk.Head.Witness != account.GetIdByPubkey(signature.Pubkey) {
		return ErrPubkey
	}

	// verify block witness signature
	if !common.VerifySignature(headInfo, signature) {
		return ErrSignature
	}

	// block produced by itself: do not verify the rest parts
	if blk.Head.Witness == staticProperty.ID {
		return nil
	}

	// verify slot map
	if staticProperty.hasSlotWitness(uint64(blk.Head.Time), blk.Head.Witness) {
		return ErrSlotWitness
	}

	return nil
}

func verifyBlock(blk *block.Block, parent *block.Block, lib *block.Block, txPool txpool.TxPool, db *db.MVCCDB) error {
	err := consensus_common.VerifyBlockHead(blk, parent, lib)
	if err != nil {
		return err
	}
	for _, tx := range blk.Txs {
		if dynamicProperty.slotToTimestamp(blk.Head.Time).ToUnixSec()-tx.Time/1e9 > 60 {
			return ErrTxTooOld
		}
		exist, _ := txPool.ExistTxs(tx.Hash(), parent)
		if exist == txpool.FoundChain {
			return ErrTxDup
		} else if exist != txpool.FoundPending {
			if err := tx.VerifySelf(); err != nil {
				return ErrTxSignature
			}
		}
	}
	err = consensus_common.VerifyBlockWithVM(blk, db)
	if err != nil {
		return err
	}
	return nil
}

func updateNodeInfo(node *blockcache.BlockCacheNode) {
	node.Number = uint64(node.Block.Head.Number)
	node.Witness = node.Block.Head.Witness
	if number, has := staticProperty.Watermark[node.Witness]; has {
		node.ConfirmUntil = number
		if node.Number >= number {
			staticProperty.Watermark[node.Witness] = node.Number + 1
		}
	} else {
		node.ConfirmUntil = 0
		staticProperty.Watermark[node.Witness] = node.Number + 1
	}
	staticProperty.addSlotWitness(uint64(node.Block.Head.Time), node.Witness)
}

func updatePendingWitness(node *blockcache.BlockCacheNode, db *db.MVCCDB) []string {
	// TODO how to decode witness list from db?
	//newList, err := db.Get("state", "witnessList")
	var err error
	if err == nil {
		//node.PendingWitnessList = newList
		node.LastWitnessListNumber = node.Number
	} else {
		node.PendingWitnessList = node.Parent.PendingWitnessList
		node.LastWitnessListNumber = node.Parent.LastWitnessListNumber
	}
	return nil
}

func calculateConfirm2(node *blockcache.BlockCacheNode, root *blockcache.BlockCacheNode) *blockcache.BlockCacheNode {
	confirmNumber := staticProperty.NumberOfWitnesses*2/3 + 1
	startNumber := node.Number
	libNumber := root.Number
	votedWitnesses := make([][]string, startNumber-libNumber)
	var j uint64 = 0
	for node != root {
		for i := 0; uint64(i) < startNumber-node.ConfirmUntil+1; i++ {
			votedWitnesses[i] = append(votedWitnesses[0], node.Witness)
		}
		if len(votedWitnesses[j]) == confirmNumber {
			return node
		}
		j++
		node = node.Parent
	}
	return nil
}
func calculateConfirm(node *blockcache.BlockCacheNode, root *blockcache.BlockCacheNode) *blockcache.BlockCacheNode {
	confirmNumber := staticProperty.NumberOfWitnesses*2/3 + 1
	startNumber := node.Number
	libNumber := root.Number
	confirmMap := make(map[string]int)
	votedWitnesses := make([][]string, startNumber-libNumber+1)
	for node != root {
		if node.ConfirmUntil <= node.Number {
			if num, err := confirmMap[node.Witness]; err {
				confirmMap[node.Witness] = 1
			} else {
				confirmMap[node.Witness] = num + 1
			}
			index := int64(node.ConfirmUntil) - int64(libNumber)
			if index > 0 {
				votedWitnesses[index] = append(votedWitnesses[index], node.Witness)
			}
		}
		if len(confirmMap) >= confirmNumber {
			staticProperty.delSlotWitness(libNumber, node.Number)
			return node
		}
		i := node.Number - libNumber
		if votedWitnesses[i] != nil {
			for _, witness := range votedWitnesses[i] {
				confirmMap[witness]--
				if confirmMap[witness] == 0 {
					delete(confirmMap, witness)
				}
			}
		}
		node = node.Parent
	}
	return nil
}

func promoteWitness(node *blockcache.BlockCacheNode, confirmed *blockcache.BlockCacheNode) {
	// update the last pending witness list that has been confirmed
	for node != confirmed && node.LastWitnessListNumber > confirmed.Number {
		node = node.Parent
	}
	if node.PendingWitnessList != nil {
		staticProperty.updateWitnessList(node.PendingWitnessList)
	}
}
