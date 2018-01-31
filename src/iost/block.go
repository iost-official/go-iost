package iost

type IBlock interface {
	Serializable
	SelfHash() []byte
	SuperHash() []byte
	ContentHash() []byte
}


// 以下是btc的block实现

type BtcHeader struct {
	ver        uint32
	superHash  []byte
	merkleHash []byte
	timeStamp  uint32
	difficulty uint32
	nonce      uint32
}

func (h *BtcHeader) Bytes() []byte {
	var s Binary
	s.putUInt(h.ver).putBytes(h.superHash).putBytes(h.merkleHash)
	s.putUInt(h.timeStamp).putUInt(h.difficulty).putUInt(h.nonce)
	return s.Bytes()
}

type BtcBlock struct {
	size        uint32
	header      BtcHeader
	tradeAmount uint32
	trades      []Transaction
}

func (b *BtcBlock) Bytes() []byte {
	var tmp Binary
	tmp.putUInt(b.size)
	tmp.put(&(b.header))
	tmp.putUInt(b.tradeAmount)
	for i := 0; i < int(b.tradeAmount); i ++ {
		tmp.putBytes(b.trades[i].Bytes())
	}
	return tmp.Bytes()
}

func (b *BtcBlock) SelfHash() []byte {
	return Sha256(b.header.Bytes())
}
func (b *BtcBlock) SuperHash() []byte {
	return b.header.superHash[:]
}
func (b *BtcBlock) ContentHash() []byte {
	return b.header.merkleHash[:]
}

// end of BtcBlock