package blockcache

import (
	"strconv"

	"github.com/iost-official/go-iost/ilog"
)

// PICSIZE draw the blockcache, for debug
const PICSIZE int = 1000

// Variable for drawing the blockcache
var pic = makePic()
var picX, picY int

func makePic() [][]string {
	a := make([][]string, 0)
	for i := 0; i < PICSIZE; i++ {
		s := make([]string, PICSIZE)
		a = append(a, s)
	}
	return a
}

func calcTree(root *BlockCacheNode, x int, y int, isLast bool) int {
	if x >= PICSIZE || y >= PICSIZE {
		return 0
	}
	if x > picX {
		picX = x
	}
	if y > picY {
		picY = y
	}
	if y != 0 {
		pic[x][y-1] = "-"
		for i := x; i >= 0; i-- {
			if pic[i][y-2] != " " {
				break
			}
			pic[i][y-2] = "|"
		}
	}
	pic[x][y] = strconv.FormatInt(root.Head.Number, 10)
	if root != nil && len(root.Head.Witness) >= 6 {
		pic[x][y] += "(" + root.Head.Witness[4:6] + ")"
	}
	var width int
	var f bool
	i := 0
	for k := range root.Children {
		if i == len(root.Children)-1 {
			f = true
		}
		if x+width < PICSIZE && y+2 < PICSIZE {
			width = calcTree(k, x+width, y+2, f)
		}
		i++
	}
	if isLast {
		return x + width
	}
	return x + width + 2
}

// DrawTree returns the the graph format of blockcache tree.
func (bcn *BlockCacheNode) DrawTree() string {
	picX, picY = 0, 0
	var ret string
	for i := 0; i < PICSIZE; i++ {
		for j := 0; j < PICSIZE; j++ {
			pic[i][j] = " "
		}
	}
	calcTree(bcn, 0, 0, true)
	if picX > PICSIZE-1 {
		picX = PICSIZE - 1
	}
	if picY > PICSIZE-1 {
		picY = PICSIZE - 1
	}
	for i := 0; i <= picX; i++ {
		l := ""
		for j := 0; j <= picY; j++ {
			l = l + pic[i][j]
		}
		ret += l
	}
	ilog.Info(ret)
	return ret
}

// Draw returns the linkedroot's and singleroot's tree graph.
func (bc *BlockCacheImpl) Draw() string {
	return bc.linkedRoot.DrawTree() + "\n\n" + bc.singleRoot.DrawTree()
}
