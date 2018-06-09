package block

import (
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/vm"
	"strconv"
)

//go:generate gencode go -schema=structs.schema -package=block

// Block 是一个区块的结构体定义
type Block struct {
	Head    BlockHead
	Content []tx.Tx //TODO:make it general for other structs
}

// Encode 是区块的序列化方法
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

// Decode 是区块的反序列方法
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

// Hash 返回区块的Hash值
func (d *Block) Hash() []byte {
	return common.Sha256(d.Encode())
}

//
func (d *Block) HashID() string {
	id := d.Head.Witness +
		strconv.FormatInt(d.Head.Time, 10) +
		strconv.FormatInt(d.Head.Number, 10) +
		strconv.FormatInt(d.Head.Version, 10)
	return id
}

// HeadHash 返回区块头部的Hash值
func (d *Block) HeadHash() []byte {
	return d.Head.Hash()
}

// GetTx 通过Content中交易索引，获取交易
func (d *Block) GetTx(x int) tx.Tx {
	if x < len(d.Content) {
		return d.Content[x]
	} else {
		return tx.Tx{}
	}
}

// LenTx 获取一个block中交易的数量
func (d *Block) LenTx() int {
	return len(d.Content)
}

// GetAllContract 获取一个block中，所有Contract的集合
func (d *Block) GetAllContract() []vm.Contract {

	var allContract []vm.Contract

	for _, tx := range d.Content {
		allContract = append(allContract, tx.Contract)
	}

	return allContract
}

// Encode blockhead的序列化方法
func (d *BlockHead) Encode() []byte {
	bin, err := d.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return bin
}

// Decode blockhead反序列方法
func (d *BlockHead) Decode(bin []byte) error {
	_, err := d.Unmarshal(bin)
	return err
}

// Hash 返回blockhead的Hash值
func (d *BlockHead) Hash() []byte {
	return common.Sha256(d.Encode())
}
