package pob

import (
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
)

func TestConfirmNode(t *testing.T) {
	Convey("Test of Confirm node", t, func() {

		staticProp.WitnessList = []string{"id0", "id1", "id2", "id3", "id4"}
		staticProp.NumberOfWitnesses = 5
		bc := blockcache.NewBlockCache(&db.MVCCDB{})
		// Root of linked tree is confirmed
		bc.LinkedTree = &blockcache.BlockCacheNode{
			Number:       1,
			Witness:      "id0",
			ConfirmUntil: 0,
		}
		Convey("Normal", func() {
			node := addNode(bc.LinkedTree, 2, 0, "id1")
			node = addNode(node, 3, 0, "id2")
			node = addNode(node, 4, 0, "id3")
			node = addNode(node, 5, 0, "id4")

			confirmNode := calculateConfirm(node, bc.LinkedTree)
			So(confirmNode.Number, ShouldEqual, 2)
		})

		Convey("Disordered normal", func() {
			node := addNode(bc.LinkedTree, 2, 0, "id1")
			node = addNode(node, 3, 0, "id2")
			node = addNode(node, 4, 2, "id0")
			node = addNode(node, 5, 4, "id2")
			node = addNode(node, 6, 3, "id1")
			node = addNode(node, 7, 0, "id3")

			confirmNode := calculateConfirm(node, bc.LinkedTree)
			So(confirmNode.Number, ShouldEqual, 4)
		})

		Convey("Disordered not enough", func() {
			node := addNode(bc.LinkedTree, 2, 0, "id1")
			node = addNode(node, 3, 0, "id2")
			node = addNode(node, 4, 0, "id3")
			node = addNode(node, 5, 3, "id4")
			confirmNode := calculateConfirm(node, bc.LinkedTree)
			So(confirmNode, ShouldBeNil)

			node = addNode(node, 6, 4, "id5")
			confirmNode = calculateConfirm(node, bc.LinkedTree)
			So(confirmNode, ShouldBeNil)

			node = addNode(node, 7, 2, "id0")
			confirmNode = calculateConfirm(node, bc.LinkedTree)
			So(confirmNode.Number, ShouldEqual, 4)
		})
	})

}

func TestPromoteWitness(t *testing.T) {
	Convey("Test of Promote Witness", t, func() {
		staticProp.WitnessList = []string{"id0", "id1", "id2"}
		staticProp.NumberOfWitnesses = 3
		bc := blockcache.NewBlockCache(&db.MVCCDB{})
		bc.LinkedTree = &blockcache.BlockCacheNode{
			Number:                1,
			Witness:               "id0",
			PendingWitnessList:    []string{"id0", "id1", "id2"},
			LastWitnessListNumber: 1,
		}
		Convey("Normal", func() {
			node := addNode(bc.LinkedTree, 2, 0, "id1")
			node.PendingWitnessList = []string{"id3", "id2", "id1"}
			node.LastWitnessListNumber = 2

			lastNode := node
			node = addNode(node, 3, 0, "id2")
			node.PendingWitnessList = lastNode.PendingWitnessList
			node.LastWitnessListNumber = lastNode.LastWitnessListNumber

			lastNode = node
			node = addNode(node, 4, 2, "id0")
			node.PendingWitnessList = lastNode.PendingWitnessList
			node.LastWitnessListNumber = lastNode.LastWitnessListNumber

			confirmNode := calculateConfirm(node, bc.LinkedTree)
			So(confirmNode.Number, ShouldEqual, 2)
			promoteWitness(node, confirmNode)
			So(staticProp.WitnessList[0], ShouldEqual, "id3")
		})

		Convey("Promote Newest", func() {
			node := addNode(bc.LinkedTree, 2, 0, "id1")
			node.PendingWitnessList = []string{"id3", "id2", "id1"}
			node.LastWitnessListNumber = 2

			lastNode := node
			node = addNode(node, 3, 0, "id2")
			node.PendingWitnessList = lastNode.PendingWitnessList
			node.LastWitnessListNumber = lastNode.LastWitnessListNumber

			lastNode = node
			node = addNode(node, 4, 3, "id1")
			node.PendingWitnessList = []string{"id2", "id3", "id4"}
			node.LastWitnessListNumber = 4

			lastNode = node
			node = addNode(node, 5, 4, "id2")
			node.PendingWitnessList = lastNode.PendingWitnessList
			node.LastWitnessListNumber = lastNode.LastWitnessListNumber

			confirmNode := calculateConfirm(node, bc.LinkedTree)
			So(confirmNode, ShouldBeNil)

			lastNode = node
			node = addNode(node, 6, 2, "id0")
			node.PendingWitnessList = []string{"id5", "id2", "id3"}
			node.LastWitnessListNumber = 6

			confirmNode = calculateConfirm(node, bc.LinkedTree)
			So(confirmNode.Number, ShouldEqual, 4)
			promoteWitness(node, confirmNode)
			So(staticProp.WitnessList[0], ShouldEqual, "id2")
		})
	})
}

func TestNodeInfoUpdate(t *testing.T) {
	Convey("Test of node info update", t, func() {

		staticProp = newGlobalStaticProperty(account.Account{"id0",[]byte{}, []byte{}}, []string{"id0", "id1", "id2"})
		bc := blockcache.NewBlockCache(&db.MVCCDB{})
		bc.LinkedTree = &blockcache.BlockCacheNode{
			Number:  1,
			Witness: "id0",
		}
		staticProp.addSlotWitness(1,"id0")
		staticProp.Watermark["id0"] = 2
		Convey("Normal", func() {
			node := addBlock(bc.LinkedTree, 2, "id1", 2)
			updateNodeInfo(node)
			So(staticProp.Watermark["id1"], ShouldEqual, 3)
			So(staticProp.hasSlotWitness(2,"id1"), ShouldBeTrue)

			node = addBlock(node, 3, "id2", 3)
			updateNodeInfo(node)
			So(staticProp.Watermark["id2"], ShouldEqual, 4)
			So(staticProp.hasSlotWitness(3,"id2"), ShouldBeTrue)

			node = addBlock(node, 4, "id0", 4)
			updateNodeInfo(node)
			So(staticProp.Watermark["id0"], ShouldEqual, 5)
			So(staticProp.hasSlotWitness(4,"id0"), ShouldBeTrue)

			node = calculateConfirm(node, bc.LinkedTree)
			So(node.Number, ShouldEqual, 2)
		})

		Convey("Slot witness error", func() {
			node := addBlock(bc.LinkedTree, 2, "id1", 2)
			updateNodeInfo(node)

			node = addBlock(node, 3, "id1", 2)
			updateNodeInfo(node)
			So(staticProp.hasSlotWitness(2, "id1"), ShouldBeTrue)
		})

		Convey("Watermark test", func() {
			node := addBlock(bc.LinkedTree, 2, "id1", 2)
			updateNodeInfo(node)
			So(node.ConfirmUntil, ShouldEqual, 0)
			branchNode := node

			node = addBlock(node, 3, "id2", 3)
			updateNodeInfo(node)

			newNode := addBlock(branchNode, 3, "id0", 4)
			updateNodeInfo(newNode)
			So(newNode.ConfirmUntil, ShouldEqual, 2)
			confirmNode := calculateConfirm(newNode, bc.LinkedTree)
			So(confirmNode, ShouldBeNil)
			So(staticProp.Watermark["id0"], ShouldEqual, 4)

			node = addBlock(node, 4, "id1", 5)
			updateNodeInfo(node)
			So(node.ConfirmUntil, ShouldEqual, 3)

			node = addBlock(node, 5, "id0", 7)
			updateNodeInfo(node)
			So(node.ConfirmUntil, ShouldEqual, 4)
			confirmNode = calculateConfirm(node, bc.LinkedTree)
			So(confirmNode, ShouldBeNil)

			node = addBlock(node, 6, "id2", 9)
			updateNodeInfo(node)
			confirmNode = calculateConfirm(node, bc.LinkedTree)
			So(confirmNode.Number, ShouldEqual, 4)
		})
	})
}

func addNode(parent *blockcache.BlockCacheNode, number uint64, confirm uint64, witness string) *blockcache.BlockCacheNode {
	node := &blockcache.BlockCacheNode{
		Parent:       parent,
		Number:       number,
		ConfirmUntil: confirm,
		Witness:      witness,
	}
	return node
}

func addBlock(parent *blockcache.BlockCacheNode, number uint64, witness string, ts int64) *blockcache.BlockCacheNode {
	blk := &block.Block{
		Head: block.BlockHead{
			Number:  int64(number),
			Witness: witness,
			Time:    ts,
		},
	}
	node := &blockcache.BlockCacheNode{
		Parent: parent,
		Block:  blk,
	}
	return node
}