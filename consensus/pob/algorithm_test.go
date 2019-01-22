package pob

import (
	"testing"
	"time"

	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/core/txpool/mock"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/verifier"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/native"
	"github.com/smartystreets/goconvey/convey"
)

var testID = []string{
	"IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn", "EhNiaU4DzUmjCrvynV3gaUeuj2VjB1v2DCmbGD5U2nSE",
	"IOST558jUpQvBD7F3WTKpnDAWg6HwKrfFiZ7AqhPFf4QSrmjdmBGeY", "8dJ9YKovJ5E7hkebAQaScaG1BA8snRUHPUbVcArcTVq6",
}

func MakeTx(act *tx.Action) (*tx.Tx, error) {
	trx := tx.NewTx([]*tx.Action{act}, nil, 10000, 1, 10000000, 0, 0)

	ac, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
	if err != nil {
		return nil, err
	}
	trx, err = tx.SignTx(trx, ac.ReadablePubkey(), []*account.KeyPair{ac})
	if err != nil {
		return nil, err
	}
	return trx, nil
}

func BenchmarkGenerateBlock(b *testing.B) { // 296275 = 0.3ms(0tx), 466353591 = 466ms(3000tx)
	account, _ := account.NewKeyPair(nil, crypto.Secp256k1)
	topBlock := &block.Block{
		Head: &block.BlockHead{
			ParentHash: []byte("abc"),
			Number:     10,
			Witness:    "witness",
			Time:       123456,
		},
	}
	topBlock.CalculateHeadHash()
	mockController := gomock.NewController(nil)
	stateDB, err := db.NewMVCCDB("./StateDB")
	if err != nil {
		b.Fatal(err)
	}
	defer stateDB.Close()
	vi := database.NewVisitor(0, stateDB)
	vi.SetTokenBalance("iost", testID[0], 100000000000000000)
	vi.SetContract(native.SystemABI())
	vi.Commit()
	stateDB.Commit(string(topBlock.HeadHash()))
	mockTxPool := txpool_mock.NewMockTxPool(mockController)
	pendingTx := txpool.NewSortedTxMap()
	for i := 0; i < 40000; i++ {
		act := tx.NewAction("system.iost", "Transfer", fmt.Sprintf(`["%v","%v",%v]`, testID[0], testID[2], "100"))
		trx, _ := MakeTx(act)
		pendingTx.Add(trx)
	}
	mockTxPool.EXPECT().PendingTx().Return(pendingTx, &blockcache.BlockCacheNode{Block: topBlock}).AnyTimes()
	mockTxPool.EXPECT().DelTxList(gomock.Any()).AnyTimes()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		generateBlock(account, mockTxPool, stateDB, time.Millisecond*1000)
	}
	b.StopTimer()
}

func BenchmarkVerifyBlockWithVM(b *testing.B) { // 296275 = 0.3ms(0tx), 466353591 = 466ms(3000tx)
	account, _ := account.NewKeyPair(nil, crypto.Secp256k1)
	topBlock := &block.Block{
		Head: &block.BlockHead{
			ParentHash: []byte("abc"),
			Number:     10,
			Witness:    "witness",
			Time:       123456,
		},
	}
	topBlock.CalculateHeadHash()
	mockController := gomock.NewController(nil)
	stateDB, err := db.NewMVCCDB("./StateDB")
	if err != nil {
		b.Fatal(err)
	}
	defer stateDB.Close()
	vi := database.NewVisitor(0, stateDB)
	vi.SetTokenBalance("iost", testID[0], 100000000000000000)
	vi.SetContract(native.SystemABI())
	vi.Commit()
	stateDB.Commit(string(topBlock.HeadHash()))
	mockTxPool := txpool_mock.NewMockTxPool(mockController)
	pendingTx := txpool.NewSortedTxMap()
	for i := 0; i < 30000; i++ {
		act := tx.NewAction("system.iost", "Transfer", fmt.Sprintf(`["%v","%v",%v]`, testID[0], testID[2], "100"))
		trx, _ := MakeTx(act)
		pendingTx.Add(trx)
	}
	mockTxPool.EXPECT().PendingTx().Return(pendingTx, &blockcache.BlockCacheNode{Block: topBlock}).AnyTimes()
	mockTxPool.EXPECT().DelTxList(gomock.Any()).AnyTimes()
	blk, _ := generateBlock(account, mockTxPool, stateDB, time.Millisecond*1000)

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		v := verifier.Verifier{}
		v.Verify(blk, topBlock, stateDB, &verifier.Config{
			Mode:        0,
			Timeout:     time.Millisecond * 1000,
			TxTimeLimit: time.Millisecond * 100,
		})
	}
	b.StopTimer()
}

func TestConfirmNode(t *testing.T) {
	convey.Convey("Test of Confirm node", t, func() {

		acc, _ := account.NewKeyPair(nil, crypto.Secp256k1)
		staticProperty = newStaticProperty(acc, []string{"id0", "id1", "id2", "id3", "id4"})

		rootNode := &blockcache.BlockCacheNode{
			Block: &block.Block{
				Head: &block.BlockHead{
					Number:  1,
					Witness: "id0",
				},
			},
			ConfirmUntil: 0,
		}
		convey.Convey("Normal", func() {
			node := addNode(rootNode, 2, 0, "id1")
			node = addNode(node, 3, 0, "id2")
			node = addNode(node, 4, 0, "id3")
			node = addNode(node, 5, 0, "id4")

			confirmNode := calculateConfirm(node, rootNode)
			convey.So(confirmNode.Head.Number, convey.ShouldEqual, 2)
		})

		convey.Convey("Diconvey.Sordered normal", func() {
			node := addNode(rootNode, 2, 0, "id1")
			node = addNode(node, 3, 0, "id2")
			node = addNode(node, 4, 2, "id0")
			node = addNode(node, 5, 4, "id2")
			node = addNode(node, 6, 3, "id1")
			node = addNode(node, 7, 0, "id3")

			confirmNode := calculateConfirm(node, rootNode)
			convey.So(confirmNode.Head.Number, convey.ShouldEqual, 4)
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
			convey.So(confirmNode.Head.Number, convey.ShouldEqual, 4)
		})
	})
}

func TestNodeInfoUpdate(t *testing.T) {
	convey.Convey("Test of node info update", t, func() {
		kp, _ := account.NewKeyPair(nil, crypto.Ed25519)
		k := kp.ReadablePubkey()
		staticProperty = newStaticProperty(kp, []string{k, "id1", "id2"})
		rootNode := &blockcache.BlockCacheNode{
			Block: &block.Block{
				Head: &block.BlockHead{
					Number:  1,
					Witness: k,
				},
			},
			Children: make(map[*blockcache.BlockCacheNode]bool),
		}
		staticProperty.Watermark[k] = 2
		convey.Convey("Normal", func() {
			node := addBlock(rootNode, 2, "id1", 2)
			updateWaterMark(node)
			convey.So(staticProperty.Watermark["id1"], convey.ShouldEqual, 3)

			node = addBlock(node, 3, "id2", 3)
			updateWaterMark(node)
			convey.So(staticProperty.Watermark["id2"], convey.ShouldEqual, 4)

			node = addBlock(node, 4, k, 4)
			updateWaterMark(node)
			convey.So(staticProperty.Watermark[k], convey.ShouldEqual, 5)

			node = calculateConfirm(node, rootNode)
			convey.So(node.Head.Number, convey.ShouldEqual, 2)
		})

		convey.Convey("Slot witness error", func() {
			node := addBlock(rootNode, 2, "id1", 2)
			updateWaterMark(node)

			node = addBlock(node, 3, "id1", 2)
			updateWaterMark(node)
		})

		convey.Convey("Watermark test", func() {
			node := addBlock(rootNode, 2, "id1", 2)
			updateWaterMark(node)
			convey.So(node.ConfirmUntil, convey.ShouldEqual, 0)
			branchNode := node

			node = addBlock(node, 3, "id2", 3)
			updateWaterMark(node)

			newNode := addBlock(branchNode, 3, k, 4)
			updateWaterMark(newNode)
			convey.So(newNode.ConfirmUntil, convey.ShouldEqual, 2)
			confirmNode := calculateConfirm(newNode, rootNode)
			convey.So(confirmNode, convey.ShouldBeNil)
			convey.So(staticProperty.Watermark[k], convey.ShouldEqual, 4)
			node = addBlock(node, 4, "id1", 5)
			updateWaterMark(node)
			convey.So(node.ConfirmUntil, convey.ShouldEqual, 3)

			node = addBlock(node, 5, k, 7)
			updateWaterMark(node)
			convey.So(node.ConfirmUntil, convey.ShouldEqual, 4)
			confirmNode = calculateConfirm(node, rootNode)
			convey.So(confirmNode, convey.ShouldBeNil)

			node = addBlock(node, 6, "id2", 9)
			updateWaterMark(node)
			confirmNode = calculateConfirm(node, rootNode)
			convey.So(confirmNode.Head.Number, convey.ShouldEqual, 4)
		})
	})
}

func TestVerifyBasics(t *testing.T) {
	convey.Convey("Test of verifyBasics", t, func() {
		secKey := common.Sha3([]byte("secKey of id0"))
		account0, _ := account.NewKeyPair(secKey, crypto.Secp256k1)
		secKey = common.Sha3([]byte("secKey of id1"))
		account1, _ := account.NewKeyPair(secKey, crypto.Secp256k1)
		staticProperty = newStaticProperty(account1, []string{account0.ReadablePubkey(), account1.ReadablePubkey(), "id2"})
		convey.Convey("Normal (self block)", func() {
			blk := &block.Block{
				Head: &block.BlockHead{
					Time:    1,
					Witness: account1.ReadablePubkey(),
				},
			}
			blk.CalculateHeadHash()
			//info := generateHeadInfo(blk.Head)
			hash := blk.HeadHash()
			blk.Sign = account1.Sign(hash)
			err := verifyBasics(blk, blk.Sign)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Normal (other's block)", func() {
			blk := &block.Block{
				Head: &block.BlockHead{
					Time:    0,
					Witness: account0.ReadablePubkey(),
				},
			}
			blk.CalculateHeadHash()
			hash := blk.HeadHash()
			blk.Sign = account0.Sign(hash)
			err := verifyBasics(blk, blk.Sign)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Wrong witness/pubkey/signature", func() {
			blk := &block.Block{
				Head: &block.BlockHead{
					Time:    1,
					Witness: account0.ReadablePubkey(),
				},
			}
			blk.CalculateHeadHash()
			//err := verifyBasics(blk.Head, blk.Sign)
			//convey.So(err, convey.ShouldEqual, errWitness)

			blk.Head.Witness = account1.ReadablePubkey()
			hash := blk.HeadHash()
			blk.Sign = account0.Sign(hash)
			err := verifyBasics(blk, blk.Sign)
			convey.So(err, convey.ShouldEqual, errSignature)
		})
		/*
			convey.Convey("Slot witness duplicate", func() {
				blk := &block.Block{
					Head: &block.BlockHead{
						Time:    0,
						Witness: account0.ID,
					},
				}
				blk.CalculateHeadHash()
				hash, _ := blk.HeadHash()
				blk.Sign = account0.Sign(crypto.Secp256k1, hash)
				err := verifyBasics(blk.Head, blk.Sign)
				convey.So(err, convey.ShouldBeNil)

				staticProperty.addSlot(0)
				blk = &block.Block{
					Head: &block.BlockHead{
						Time:    0,
						Witness: account0.ID,
					},
				}
			blk.CalculateHeadHash()
				hash, _ = blk.HeadHash()
				blk.Sign = account0.Sign(crypto.Secp256k1, hash)
				err = verifyBasics(blk.Head, blk.Sign)
				convey.So(err, convey.ShouldEqual, errSlot)
			})
		*/
	})
}

func TestVerifyBlock(t *testing.T) {
	convey.Convey("Test of verify block", t, func() {
		secKey := common.Sha3([]byte("secKey of id0"))
		account0, _ := account.NewKeyPair(secKey, crypto.Secp256k1)
		secKey = common.Sha3([]byte("sec of id1"))
		account1, _ := account.NewKeyPair(secKey, crypto.Secp256k1)
		secKey = common.Sha3([]byte("sec of id2"))
		account2, _ := account.NewKeyPair(secKey, crypto.Secp256k1)
		staticProperty = newStaticProperty(account0, []string{account0.ReadablePubkey(), account1.ReadablePubkey(), account2.ReadablePubkey()})
		rootTime := time.Now().UnixNano()
		rootBlk := &block.Block{
			Head: &block.BlockHead{
				Number:  1,
				Time:    rootTime,
				Witness: witnessOfSlot(rootTime),
			},
		}
		tx0 := &tx.Tx{
			Time: time.Now().UnixNano(),
			Actions: []*tx.Action{{
				Contract:   "contract1",
				ActionName: "actionname1",
				Data:       "{\"num\": 1, \"message\": \"contract1\"}",
			}},
			Signers: []string{account1.ReadablePubkey()},
		}
		rcpt0 := &tx.TxReceipt{
			TxHash: tx0.Hash(),
		}
		curTime := time.Now().UnixNano()
		hash := rootBlk.HeadHash()
		witness := witnessOfSlot(curTime)
		blk := &block.Block{
			Head: &block.BlockHead{
				Number:     2,
				ParentHash: hash,
				Time:       curTime,
				Witness:    witnessOfSlot(curTime),
			},
			Txs:      []*tx.Tx{},
			Receipts: []*tx.TxReceipt{},
		}
		blk.Head.TxMerkleHash = blk.CalculateTxMerkleHash()
		blk.Head.TxReceiptMerkleHash = blk.CalculateTxReceiptMerkleHash()
		info := blk.HeadHash()
		var sig *crypto.Signature
		if witness == account0.ReadablePubkey() {
			sig = account0.Sign(info)
		} else if witness == account1.ReadablePubkey() {
			sig = account1.Sign(info)
		} else {
			sig = account2.Sign(info)
		}
		blk.Sign = sig
		//convey.Convey("Normal (no txs)", func() {
		//	err := verifyBlock(blk, rootBlk, rootBlk, nil, nil)
		//	convey.So(err, convey.ShouldBeNil)
		//})

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
		Block: &block.Block{
			Head: &block.BlockHead{
				Number:  number,
				Witness: witness,
			},
		},
		ConfirmUntil: confirm,
	}
	node.SetParent(parent)
	return node
}

func addBlock(parent *blockcache.BlockCacheNode, number int64, witness string, ts int64) *blockcache.BlockCacheNode {
	blk := &block.Block{
		Head: &block.BlockHead{
			Number:  number,
			Witness: witness,
			Time:    ts,
		},
	}
	return blockcache.NewBCN(parent, blk)
}
