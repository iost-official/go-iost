package blockcache

import (
	"fmt"
	"strconv"

	"github.com/xlab/treeprint"
)

// Draw returns the linkedroot's and singleroot's tree graph.
func (bc *BlockCacheImpl) Draw() string {
	nmLen := 0
	bc.number2node.Range(func(k, v interface{}) bool {
		nmLen++
		return true
	})

	hmLen := 0
	bc.hash2node.Range(func(k, v interface{}) bool {
		hmLen++
		return true
	})

	leafLen := len(bc.leaf)

	mapInfo := fmt.Sprintf("nmLen: %v, hmLen: %v, leafLen: %v", nmLen, hmLen, leafLen)

	linkedTree := treeprint.New()
	bc.LinkedRoot().drawChildren(linkedTree)

	result := mapInfo + "\n" + linkedTree.String() + "\n"
	for _, vbcn := range bc.singleRoot {
		singleTree := treeprint.New()
		vbcn.drawChildren(singleTree)
		result = result + singleTree.String() + "\n"
	}
	return result
}

func (bcn *BlockCacheNode) drawChildren(root treeprint.Tree) {
	for c := range bcn.Children {
		pattern := strconv.Itoa(int(c.Head.Number))
		if c.Head.Witness != "" {
			pattern += "(" + c.Head.Witness[4:6] + ")"
		}
		root.AddNode(pattern)
		c.drawChildren(root.FindLastNode())
	}
}
