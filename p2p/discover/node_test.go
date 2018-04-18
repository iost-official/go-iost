package discover

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddr2Node(t *testing.T) {
	addr := "127.0.0.1:20202"
	node, err := Addr2Node(addr)
	if err != nil {
		t.Errorf("addr2Node got err%+v", err)
	}
	assert.Equal(t, addr, string(node.IP)+":"+strconv.Itoa(int(node.UDP)))
}
