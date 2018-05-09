package block

import (
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/vm"
	"fmt"
)

//go:generate gencode go -schema=structs.schema -package=block

type Block struct {
	Head    BlockHead
	Content []tx.Tx
}

func (d *Block) Encode() []byte {
	c := make([][]byte, 0)
	for _, t := range d.Content {
		c = append(c, t.Encode())
	}
	br := BlockRaw{d.Head, c}
	b, err := br.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return b
}

func (d *Block) Decode(bin []byte) error {
	var br BlockRaw
	_, err := br.Unmarshal(bin)
	d.Head = br.Head
	for _, t := range br.Content {
		var tt tx.Tx
		err = tt.Decode(t)
		if err != nil {
			return err
		}
		d.Content = append(d.Content, tt)
	}
	return nil
}

func (d *Block) Hash() []byte {
	return common.Sha256(d.Encode())
}

func (d *Block) HeadHash() []byte {
	return d.Head.Hash()
}

func (d *Block) GetTx(x int) tx.Tx {
	if x < len(d.Content) {
		return d.Content[x]
	} else {
		return tx.Tx{}
	}
}

func (d *Block) LenTx() int {
	return len(d.Content)
}
func (d *Block) GetAllContract() []vm.Contract {
	//todo 解析content,获得所有交易

	var allContract []vm.Contract
	//todo 将交易中的contract，添加到contractAll中

	return allContract
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
