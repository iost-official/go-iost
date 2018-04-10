package pow

import (
	"github.com/iost-official/prototype/core"
	"encoding/binary"
)

type Pow struct {
	Recorder

}

func (p * Pow) Init() {

}



func parseInfo (head core.BlockHead)(difficulty, nonce uint64) {
	difficulty = binary.BigEndian.Uint64(head.Info[0:8])
	nonce = binary.BigEndian.Uint64(head.Info[8:16])
	return
}

func makeInfo (dif, nonce uint64) []byte {
	b1 := make([]byte, 8)
	binary.BigEndian.PutUint64(b1, dif)
	b2 := make([]byte, 8)
	binary.BigEndian.PutUint64(b2, nonce)

	return append(b1, b2...)
}