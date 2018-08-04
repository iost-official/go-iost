package pob

import (
	"testing"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfirmNode(t *testing.T) {
	staticProp.WitnessList = []string{"id0", "id1", "id2", "id3", "id4"}
	staticProp.NumberOfWitnesses = 5
	bc := blockcache.NewBlockCache()
	// Root of linked tree is confirmed
	bc.LinkedTree = &blockcache.BlockCacheNode {
		Number:  1,
		Witness: "id0",
		ConfirmUntil: 0,
	}

	Convey("Test of Confirm node", t, func() {
		Convey("Normal", func() {
			node :=addNode(bc.LinkedTree, 2, 0, "id1")
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

func addNode(parent *blockcache.BlockCacheNode, number uint64, confirm uint64, witness string) *blockcache.BlockCacheNode {
	node := &blockcache.BlockCacheNode{
		Parent: parent,
		Number: number,
		ConfirmUntil: confirm,
		Witness: witness,
	}
	parent.Children = append(parent.Children, node)
	return node
}