package block

import (
	"fmt"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/vm"
	"strconv"
)

//go:generate gencode go -schema=structs.schema -package=block

type Block struct {
	Head     BlockHead
	Txs      []tx.Tx
	Receipts []tx.TxReceipt
}

func (d *Block) String() string {
	str := "Block{\n"
	str += "	BlockHead{\n"
	str += "		Number: " + strconv.FormatInt(d.Head.Number, 10) + ",\n"
	str += "		Time: " + strconv.FormatInt(d.Head.Time, 10) + ",\n"
	str += "		Witness: " + d.Head.Witness + ",\n"
	str += "	}\n"

	str += "	Txs {\n"
	for _, tx := range d.Txs {
		str += tx.String()
	}
	str += "	}\n"
	str += "	Receipts {\n"
	for _, receipt := range d.Receipts {
		str += receipt.String()
	}
	str += "	}\n"
	str += "}\n"
	return str
}

func (d *Block) CalculateTxsHash() []byte {
	treeHash := make([]byte, 0)
	for _, tx := range d.Txs {
		treeHash = append(treeHash, tx.Publisher.Sig...)
	}
	return common.Sha256(treeHash)
}

func (d *Block) CalculateMerkleHash() []byte {

}

func (d *Block) Encode() []byte {
	txs := make([][]byte, 0)
	for _, t := range d.Txs {
		txs = append(txs, t.Encode())
	}
	rpts := make([][]byte, 0)
	for _, r := range d.Receipts {
		rpts = append(rpts, r.Encode())
	}
	br := BlockRaw{d.Head, txs, rpts}
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
	for _, t := range br.Txs {
		var tt tx.Tx
		err = tt.Decode(t)
		if err != nil {
			return err
		}
		d.Txs = append(d.Txs, tt)
	}
	for _, r := range br.Receipts {
		var rcpt tx.TxReceipt
		err = rcpt.Decode(r)
		if err != nil {
			return err
		}
		d.Receipts = append(d.Receipts, rcpt)
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
	if x < len(d.Txs) {
		return d.Txs[x]
	} else {
		return tx.Tx{}
	}
}

func (d *Block) LenTx() int {
	return len(d.Txs)
}

func (d *Block) GetAllContract() []vm.Contract {

	var allContract []vm.Contract

	for _, tx := range d.Txs {
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
