package discover

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"strings"

	"github.com/iost-official/prototype/params"
)

func TestGenNodeId(t *testing.T) {
	Convey("Test of discover node\n", t, func() {
		So(len(GenNodeId()), ShouldEqual, 64)
		node, err := ParseNode(params.TestnetBootnodes[0])
		So(err, ShouldBeNil)
		So(node.TCP, ShouldEqual, uint16(30302))
		So(node.IP.String(), ShouldEqual, "127.0.0.1")
		So(node.ID, ShouldEqual, "84a8ecbeeb6d3f676da1b261c35c7cd15ae17f32b659a6f5ce7be2d60f6c16f9")
		//	xor distance
		dis := xorDistance("0056847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610", "0056847afe9799739b3d9677972a3b58ef4d92ae0b48d32d9c67dc9a302bfc76")
		So(dis, ShouldEqual, 278)
		dis = xorDistance("0056847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610", "0156847afe9799739b3d9677972a3b58ef4d92ae0b48d32d9c67dc9a302bfc76")
		So(dis, ShouldEqual, 15)
		dis = xorDistance("0056847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610", "0556847afe9799739b3d9677972a3b58ef4d92ae0b48d32d9c67dc9a302bfc76")
		So(dis, ShouldEqual, 13)

		nodeIds := []string{
			"0156847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610",
			"0256847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610",
			"0356847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610",
			"0456847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610",
			"0556847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610",
			"0656847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610",
			"0756847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610",
			"0856847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610",
			"0956847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610",
			"1056847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610",
			"1156847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610",
		}
		ns := make([]*Node, 0)
		for _, v := range nodeIds {
			n := &Node{ID: NodeID(v)}
			ns = append(ns, n)
		}
		neighbours := ns[0].FindNeighbours(ns)
		neighboursStr := make([]string, 0)
		for _, n := range neighbours {
			neighboursStr = append(neighboursStr, n.String())
		}
		result := strings.Join(neighboursStr, ",")

		So(result, ShouldNotContainSubstring, nodeIds[1])
		So(result, ShouldNotContainSubstring, nodeIds[0])
		So(result, ShouldContainSubstring, nodeIds[3])

	})
}
