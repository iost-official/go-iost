package discover

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"net"
)

func TestGenNodeId(t *testing.T) {
	Convey("Test of discover node\n", t, func() {
		So(len(GenNodeID()), ShouldEqual, 64)
		node, err := ParseNode("84a8ecbeeb6d3f676da1b261c35c7cd15ae17f32b659a6f5ce7be2d60f6c16f9@18.219.254.124:30304")
		So(err, ShouldBeNil)
		So(node.TCP, ShouldEqual, uint16(30304))
		So(node.IP.String(), ShouldEqual, "18.219.254.124")
		So(node.ID, ShouldEqual, "84a8ecbeeb6d3f676da1b261c35c7cd15ae17f32b659a6f5ce7be2d60f6c16f9")

	})
}

func TestNode_xorDistance(t *testing.T) {
	Convey("", t, func() {
		dis := xorDistance("0056847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610", "0056847afe9799739b3d9677972a3b58ef4d92ae0b48d32d9c67dc9a302bfc76")
		So(dis, ShouldEqual, 278)
		dis = xorDistance("0056847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610", "0156847afe9799739b3d9677972a3b58ef4d92ae0b48d32d9c67dc9a302bfc76")
		So(dis, ShouldEqual, 15)
		dis = xorDistance("0056847afe9799739b3d9677972a3b58ef609ba78332428f85ed2534d0b49610", "0556847afe9799739b3d9677972a3b58ef4d92ae0b48d32d9c67dc9a302bfc76")
		So(dis, ShouldEqual, 13)
	})
}

func TestNode_FindNeighbours(t *testing.T) {
	Convey("find neighbours", t, func() {
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
	})
}

func TestNewNode(t *testing.T) {
	Convey("", t, func() {
		var id NodeID
		id = "84a8ecbeeb6d3f676da1b261c35c7cd15ae17f32b659a6f5ce7be2d60f6c16f9"
		ip := net.ParseIP("18.219.254.124")
		n := NewNode(id, ip, 30304, 30304)
		So(n.TCP, ShouldEqual, 30304)
		So(n.IP.String(), ShouldEqual, "18.219.254.124")
	})
}

func TestNode_Incomplete(t *testing.T) {
	Convey("", t, func() {
		node, err := ParseNode("84a8ecbeeb6d3f676da1b261c35c7cd15ae17f32b659a6f5ce7be2d60f6c16f9@18.219.254.124:30304")
		So(err, ShouldBeNil)
		So(node.Incomplete(), ShouldBeFalse)
	})
}

func TestNode_Addr(t *testing.T) {
	Convey("test addr", t, func() {
		var id NodeID
		id = "84a8ecbeeb6d3f676da1b261c35c7cd15ae17f32b659a6f5ce7be2d60f6c16f9"
		ip := net.ParseIP("18.219.254.124")
		node := NewNode(id, ip, 30304, 30304)
		So(node.Addr(), ShouldEqual, "18.219.254.124:30304")
	})
}

func TestNode_String(t *testing.T) {
	Convey("test parse node and string", t, func() {
		nodeStr := "84a8ecbeeb6d3f676da1b261c35c7cd15ae17f32b659a6f5ce7be2d60f6c16f9@18.219.254.124:30304"
		node, err := ParseNode(nodeStr)
		So(err, ShouldBeNil)
		So(node.String(), ShouldEqual, nodeStr)
	})
}
