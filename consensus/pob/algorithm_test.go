package pob

import (
		"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/common"
		"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"time"
)

func TestConfirmNode(t *testing.T) {
	convey.Convey("Test of Confirm node", t, func() {
		staticProperty.WitnessList = []string{"id0", "id1", "id2", "id3", "id4"}
		staticProperty.NumberOfWitnesses = 5
		rootNode := &blockcache.BlockCacheNode{
			Number:       1,
			Witness:      "id0",
			ConfirmUntil: 0,
		}
		convey.Convey("Normal", func() {
			node := addNode(rootNode, 2, 0, "id1")
			node = addNode(node, 3, 0, "id2")
			node = addNode(node, 4, 0, "id3")
			node = addNode(node, 5, 0, "id4")

			confirmNode := calculateConfirm(node, rootNode)
			convey.So(confirmNode.Number, convey.ShouldEqual, 2)
		})

		convey.Convey("Diconvey.Sordered normal", func() {
			node := addNode(rootNode, 2, 0, "id1")
			node = addNode(node, 3, 0, "id2")
			node = addNode(node, 4, 2, "id0")
			node = addNode(node, 5, 4, "id2")
			node = addNode(node, 6, 3, "id1")
			node = addNode(node, 7, 0, "id3")

			confirmNode := calculateConfirm(node, rootNode)
			convey.So(confirmNode.Number, convey.ShouldEqual, 4)
		})

		convey.Convey("Diconvey.Sordered not enough", func() {
			node := addNode(rootNode, 2, 0, "id1")
			node = addNode(node, 3, 0, "id2")
			node = addNode(node, 4, 0, "id3")
			node = addNode(node, 5, 3, "id4")
			confirmNode := calculateConfirm(node, rootNode)
			convey.So(confirmNode, convey.ShouldBeNil)

			node = addNode(node, 6, 4, "id5")
			confirmNode = calculateConfirm(node, rootNode)
			convey.So(confirmNode, convey.ShouldBeNil)

			node = addNode(node, 7, 2, "id0")
			confirmNode = calculateConfirm(node, rootNode)
			convey.So(confirmNode.Number, convey.ShouldEqual, 4)
		})
	})
}

func TestUpdateWitness(t *testing.T) {
	convey.Convey("Test of Promote Witness", t, func() {
		staticProperty = newStaticProperty(account.Account{}, []string{"id0", "id1", "id2"})
		rootNode := &blockcache.BlockCacheNode{
			Number:                1,
			Witness:               "id0",
			PendingWitnessList:    []string{"id0", "id1", "id2"},
		}
		node := addNode(rootNode, 2, 0, "id1")
		node.PendingWitnessList = []string{"id3", "id2", "id1"}

		lastNode := node
		node = addNode(node, 3, 0, "id2")
		node.PendingWitnessList = lastNode.PendingWitnessList

		lastNode = node
		node = addNode(node, 4, 2, "id0")
		node.PendingWitnessList = lastNode.PendingWitnessList

		confirmNode := calculateConfirm(node, rootNode)
		convey.So(confirmNode.Number, convey.ShouldEqual, 2)
		staticProperty.updateWitnessList(confirmNode.PendingWitnessList)
		convey.So(staticProperty.WitnessList[0], convey.ShouldEqual, "id3")

		node = addNode(rootNode, 2, 0, "id1")
		node.PendingWitnessList = []string{"id3", "id2", "id1"}

		lastNode = node
		node = addNode(node, 3, 0, "id2")
		node.PendingWitnessList = lastNode.PendingWitnessList

		lastNode = node
		node = addNode(node, 4, 3, "id1")
		node.PendingWitnessList = []string{"id2", "id3", "id4"}

		lastNode = node
		node = addNode(node, 5, 4, "id2")
		node.PendingWitnessList = lastNode.PendingWitnessList

		confirmNode = calculateConfirm(node, rootNode)
		convey.So(confirmNode, convey.ShouldBeNil)

		lastNode = node
		node = addNode(node, 6, 2, "id0")
		node.PendingWitnessList = []string{"id5", "id2", "id3"}

		confirmNode = calculateConfirm(node, rootNode)
		convey.So(confirmNode.Number, convey.ShouldEqual, 4)
		staticProperty.updateWitnessList(confirmNode.PendingWitnessList)
		convey.So(staticProperty.WitnessList[0], convey.ShouldEqual, "id2")
	})
}

func TestNodeInfoUpdate(t *testing.T) {
	convey.Convey("Test of node info update", t, func() {
		staticProperty = newStaticProperty(account.Account{ID:"id0"}, []string{"id0", "id1", "id2"})
		rootNode := &blockcache.BlockCacheNode{
			Number:  1,
			Witness: "id0",
			Children: make(map[*blockcache.BlockCacheNode]bool),
		}
		staticProperty.addSlot(1)
		staticProperty.Watermark["id0"] = 2
		convey.Convey("Normal", func() {
			node := addBlock(rootNode, 2, "id1", 2)
			updateNodeInfo(node)
			convey.So(staticProperty.Watermark["id1"], convey.ShouldEqual, 3)
			convey.So(staticProperty.hasSlot(2), convey.ShouldBeTrue)

			node = addBlock(node, 3, "id2", 3)
			updateNodeInfo(node)
			convey.So(staticProperty.Watermark["id2"], convey.ShouldEqual, 4)
			convey.So(staticProperty.hasSlot(3), convey.ShouldBeTrue)

			node = addBlock(node, 4, "id0", 4)
			updateNodeInfo(node)
			convey.So(staticProperty.Watermark["id0"], convey.ShouldEqual, 5)
			convey.So(staticProperty.hasSlot(4),convey.ShouldBeTrue)

			node = calculateConfirm(node, rootNode)
			convey.So(node.Number, convey.ShouldEqual, 2)
		})

		convey.Convey("Slot witness error", func() {
			node := addBlock(rootNode, 2, "id1", 2)
			updateNodeInfo(node)

			node = addBlock(node, 3, "id1", 2)
			updateNodeInfo(node)
			convey.So(staticProperty.hasSlot(2), convey.ShouldBeTrue)
		})

		convey.Convey("Watermark test", func() {
			node := addBlock(rootNode, 2, "id1", 2)
			updateNodeInfo(node)
			convey.So(node.ConfirmUntil, convey.ShouldEqual, 0)
			branchNode := node

			node = addBlock(node, 3, "id2", 3)
			updateNodeInfo(node)

			newNode := addBlock(branchNode, 3, "id0", 4)
			updateNodeInfo(newNode)
			convey.So(newNode.ConfirmUntil, convey.ShouldEqual, 2)
			confirmNode := calculateConfirm(newNode, rootNode)
			convey.So(confirmNode, convey.ShouldBeNil)
			convey.So(staticProperty.Watermark["id0"], convey.ShouldEqual, 4)
			node = addBlock(node, 4, "id1", 5)
			updateNodeInfo(node)
			convey.So(node.ConfirmUntil, convey.ShouldEqual, 3)

			node = addBlock(node, 5, "id0", 7)
			updateNodeInfo(node)
			convey.So(node.ConfirmUntil, convey.ShouldEqual, 4)
			confirmNode = calculateConfirm(node, rootNode)
			convey.So(confirmNode, convey.ShouldBeNil)

			node = addBlock(node, 6, "id2", 9)
			updateNodeInfo(node)
			confirmNode = calculateConfirm(node, rootNode)
			convey.So(confirmNode.Number, convey.ShouldEqual, 4)
		})
	})
}

func TestVerifyBlockHead(t *testing.T) {
	convey.Convey("Test of verify block head", t, func() {
		parentBlk := &block.Block{
			Head: block.BlockHead{
				Number: 3,
				Time:   common.GetCurrentTimestamp().Slot - 1,
			},
		}
		chainTop := &block.Block{
			Head: block.BlockHead{
				Number: 1,
				Time:   common.GetCurrentTimestamp().Slot - 4,
			},
		}
		hash := parentBlk.HeadHash()
		blk := &block.Block{
			Head: block.BlockHead{
				ParentHash: hash,
				Number:     4,
				Time:       common.GetCurrentTimestamp().Slot,
				TxsHash:    common.Sha256([]byte{}),
				MerkleHash: []byte{},
			},
		}
		convey.Convey("Pass", func() {
			err := verifyBlockHead(blk, parentBlk, chainTop)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Wrong time", func() {
			blk.Head.Time = common.GetCurrentTimestamp().Slot - 5
			err := verifyBlockHead(blk, parentBlk, chainTop)
			convey.So(err, convey.ShouldEqual, ErrOldBlk)
			blk.Head.Time = common.GetCurrentTimestamp().Slot + 2
			err = verifyBlockHead(blk, parentBlk, chainTop)
			convey.So(err, convey.ShouldEqual, ErrFutureBlk)
		})

		convey.Convey("Wrong parent", func() {
			blk.Head.ParentHash = []byte("fake hash")
			err := verifyBlockHead(blk, parentBlk, chainTop)
			convey.So(err, convey.ShouldEqual, ErrParentHash)
		})

		convey.Convey("Wrong number", func() {
			blk.Head.Number = 5
			err := verifyBlockHead(blk, parentBlk, chainTop)
			convey.So(err, convey.ShouldEqual, ErrNumber)
		})

		convey.Convey("Wrong tx hash", func() {
			tx0 := tx.NewTx(nil, nil, 1000, 1, 300)
			blk.Txs = append(blk.Txs, &tx0)
			blk.Head.TxsHash = blk.CalculateTxsHash()
			err := verifyBlockHead(blk, parentBlk, chainTop)
			convey.So(err, convey.ShouldBeNil)
			blk.Head.TxsHash = []byte("fake hash")
			err = verifyBlockHead(blk, parentBlk, chainTop)
			convey.So(err, convey.ShouldEqual, ErrTxHash)
		})
	})
}

func TestVerifyBasics(t *testing.T) {
	convey.Convey("Test of verifyBasics", t, func() {
		secKey := common.Sha256([]byte("secKey of id0"))
		account0, _ := account.NewAccount(secKey)
		secKey = common.Sha256([]byte("secKey of id1"))
		account1, _ := account.NewAccount(secKey)
		staticProperty = newStaticProperty(account1, []string{account0.ID, account1.ID, "id2"})
		convey.Convey("Normal (self block)", func() {
			blk := &block.Block{
				Head: block.BlockHead{
					Time: 1,
					Witness: account1.ID,
				},
			}
			info := generateHeadInfo(blk.Head)
			sig, _ := common.Sign(common.Secp256k1, info, account1.Seckey)
			blk.Head.Signature = sig.Encode()
			err := verifyBasics(blk)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Normal (other's block)", func() {
			blk := &block.Block{
				Head: block.BlockHead{
					Time: 0,
					Witness: account0.ID,
				},
			}
			info := generateHeadInfo(blk.Head)
			sig, _ := common.Sign(common.Secp256k1, info, account0.Seckey)
			blk.Head.Signature = sig.Encode()

			err := verifyBasics(blk)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Wrong witness/pubkey/signature", func() {
			blk := &block.Block{
				Head: block.BlockHead{
					Time: 1,
					Witness: account0.ID,
				},
			}
			err := verifyBasics(blk)
			convey.So(err, convey.ShouldEqual, ErrWitness)

			blk.Head.Witness = account1.ID
			info := generateHeadInfo(blk.Head)
			sig, _ := common.Sign(common.Secp256k1, info, account0.Seckey)
			blk.Head.Signature = sig.Encode()
			err = verifyBasics(blk)
			convey.So(err, convey.ShouldEqual, ErrPubkey)

			info = generateHeadInfo(blk.Head)
			sig, _ = common.Sign(common.Secp256k1, info, account1.Seckey)
			blk.Head.Signature = sig.Encode()
			blk.Head.Info = []byte("fake info")
			err = verifyBasics(blk)
			convey.So(err, convey.ShouldEqual, ErrSignature)
		})

		convey.Convey("Slot witness duplicate", func() {
			blk := &block.Block{
				Head: block.BlockHead{
					Time: 0,
					Witness: account0.ID,
					Info: []byte("first one"),
				},
			}
			info := generateHeadInfo(blk.Head)
			sig, _ := common.Sign(common.Secp256k1, info, account0.Seckey)
			blk.Head.Signature = sig.Encode()
			err := verifyBasics(blk)
			convey.So(err, convey.ShouldBeNil)

			staticProperty.addSlot(0)
			blk = &block.Block{
				Head: block.BlockHead{
					Time: 0,
					Witness: account0.ID,
					Info: []byte("second one"),
				},
			}
			info = generateHeadInfo(blk.Head)
			sig, _ = common.Sign(common.Secp256k1, info, account0.Seckey)
			blk.Head.Signature = sig.Encode()
			err = verifyBasics(blk)
			convey.So(err, convey.ShouldEqual, ErrSlot)
		})
	})
}

func TestVerifyBlock(t *testing.T) {
	convey.Convey("Test of verify block", t, func() {
		secKey := common.Sha256([]byte("secKey of id0"))
		account0, _ := account.NewAccount(secKey)
		secKey = common.Sha256([]byte("sec of id1"))
		account1, _ := account.NewAccount(secKey)
		secKey = common.Sha256([]byte("sec of id2"))
		account2, _ := account.NewAccount(secKey)
		staticProperty = newStaticProperty(account0, []string{account0.ID, account1.ID, account2.ID})
		rootTime := common.GetCurrentTimestamp().Slot - 1
		rootBlk := &block.Block{
			Head: block.BlockHead{
				Number: 1,
				Time: rootTime,
				Witness: witnessOfSlot(rootTime),
			},
		}
		tx0 := &tx.Tx{
			Time: time.Now().UnixNano(),
			Actions:[]tx.Action{{
				Contract:"contract1",
				ActionName:"actionname1",
				Data:"{\"num\": 1, \"message\": \"contract1\"}",
			}},
			Signers:[][]byte{account1.Pubkey},
		}
		rcpt0 := &tx.TxReceipt{
			TxHash: tx0.Hash(),
		}
		curTime := common.GetCurrentTimestamp().Slot
		hash := rootBlk.HeadHash()
		witness := witnessOfSlot(curTime)
		blk := &block.Block{
			Head: block.BlockHead{
				Number: 2,
				ParentHash: hash,
				Time:curTime,
				Witness: witnessOfSlot(curTime),
			},
			Txs: []*tx.Tx{},
			Receipts: []*tx.TxReceipt{},
		}
		blk.Head.TxsHash = blk.CalculateTxsHash()
		blk.Head.MerkleHash = blk.CalculateMerkleHash()
		info := generateHeadInfo(blk.Head)
		var sig common.Signature
		if witness == account0.ID {
			sig, _ = common.Sign(common.Secp256k1, info, account0.Seckey)
		} else if witness == account1.ID {
			sig, _ = common.Sign(common.Secp256k1, info, account1.Seckey)
		} else {
			sig, _ = common.Sign(common.Secp256k1, info, account2.Seckey)
		}
		blk.Head.Signature = sig.Encode()
		convey.Convey("Normal (no txs)", func() {
			err := verifyBlock(blk, rootBlk, rootBlk, nil, nil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Normal (with txs)", func() {
			blk.Txs = append(blk.Txs, tx0)
			blk.Receipts = append(blk.Receipts, rcpt0)
			//Use mock
			//txPool, _ := txpool.NewTxPoolImpl()
			//db, _ := db.NewMVCCDB()
			//err := verifyBlock(blk, rootBlk, rootBlk, txPool, db)
			//convey.So(err, convey.ShouldBeNil)
		})
	})
}

func addNode(parent *blockcache.BlockCacheNode, number int64, confirm int64, witness string) *blockcache.BlockCacheNode {
	node := &blockcache.BlockCacheNode{
		Parent:       parent,
		Number:       number,
		ConfirmUntil: confirm,
		Witness:      witness,
	}
	return node
}

func addBlock(parent *blockcache.BlockCacheNode, number int64, witness string, ts int64) *blockcache.BlockCacheNode {
	blk := &block.Block{
		Head: block.BlockHead{
			Number:  number,
			Witness: witness,
			Time:    ts,
		},
	}
	return blockcache.NewBCN(parent, blk)
}