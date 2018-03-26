package iosbase

func (b *Block) Encode() []byte {
	bin, err := b.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return bin
}

func (b *Block) Decode(bin []byte) error {
	_,err := b.Unmarshal(bin)
	return err
}
func (b *Block) Hash() []byte {
	return Sha256(b.Encode())
}


func (b *BlockHead) Encode() []byte {
	bin, err := b.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return bin
}

func (b *BlockHead) Decode(bin []byte) error {
	_,err := b.Unmarshal(bin)
	return err
}
func (b *BlockHead) Hash() []byte {
	return Sha256(b.Encode())
}
