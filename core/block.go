package core

import "github.com/iost-official/prototype/common"

func (d *Block) Encode() []byte {
bin, err := d.Marshal(nil)
if err != nil {
panic(err)
}
return bin
}

func (d *Block) Decode(bin []byte) error {
	_, err := d.Unmarshal(bin)
	return err
}
func (d *Block) Hash() []byte {
	return common.Sha256(d.Encode())
}

func (d *Block) HeadHash() []byte {
	return d.Head.Hash()
}

func (d *BlockHead) Encode() []byte {
	bin, err := d.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return bin
}

func (d *BlockHead) Decode(bin []byte) error {
	_, err := d.Unmarshal(bin)
	return err
}
func (d *BlockHead) Hash() []byte {
	return common.Sha256(d.Encode())
}
