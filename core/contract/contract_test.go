package contract

import (
	"testing"
)

func TestCodec(t *testing.T) {
	c := Contract{
		Code: "codes",
		Info: &Info{
			Lang:    "javascript",
			Version: "1.0.0",
			Abi: []*ABI{
				{
					Name: "abi1",
				},
			},
		},
	}
	buf := c.Encode()
	var d Contract
	d.Decode(buf)
	if d.String() != c.String() {
		t.Fatal(d.String())
	}
}
