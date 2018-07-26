package block

import (
	"fmt"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/vm"
	"strconv"
)

//go:generate gencode go -schema=structs.schema -package=block

type Block struct {
	Head    BlockHead
	Content []tx.Tx //TODO:make it general for other structs
}

func (d *Block) String() string {
	str := "Block{\n"
	str += "	BlockHead{\n"
	str += "		Number: " + strconv.FormatInt(d.Head.Number, 10) + ",\n"
	str += "		Time: " + strconv.FormatInt(d.Head.Time, 10) + ",\n"
	str += "		Witness: " + d.Head.Witness + ",\n"
	str += "	}\n"

	str += "	Content{\n"
	for _, tx := range d.Content {
		str += tx.String()
	}
	str += "	}\n"
	str += "}\n"
	return str
}

func (d *Block) CalculateTreeHash() []byte {
	treeHash := make([]byte, 0)
	for _, tx := range d.Content {
		treeHash = append(treeHash, tx.Publisher.Sig...)
	}
	return common.Sha256(treeHash)
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

func (d *Block) Decode(bin []byte) (err error) {
	var br BlockRaw
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	_, err = br.Unmarshal(bin)
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

//
func (d *Block) HashID() string {
	id := d.Head.Witness +
		strconv.FormatInt(d.Head.Time, 10) +
		strconv.FormatInt(d.Head.Number, 10) +
		strconv.FormatInt(d.Head.Version, 10)
	return id
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

	var allContract []vm.Contract

	for _, tx := range d.Content {
		allContract = append(allContract, tx.Contract)
	}

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
