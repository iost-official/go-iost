package block

import (
	"fmt"
	"strconv"

	"github.com/gogo/protobuf/proto"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
)

type Block struct {
	hash     []byte
	Head     BlockHead
	Txs      []*tx.Tx
	Receipts []*tx.TxReceipt
}

func GenGenesis(initTime int64) *Block {
	var code string
	for k, v := range account.GenesisAccount {
		code += fmt.Sprintf("@PutHM iost %v f%v\n", k, v)
	}

	txn := tx.Tx{
		Time: 0,
		// TODO what is the genesis tx?
	}

	genesis := &Block{
		Head: BlockHead{
			Version: 0,
			Number:  0,
			Time:    initTime,
		},
		Txs:      make([]*tx.Tx, 0),
		Receipts: make([]*tx.TxReceipt, 0),
	}
	genesis.Txs = append(genesis.Txs, &txn)
	return genesis
}

/*
func (d *Block) String() string {
	str := "Block{\n"
	str += "	BlockHead{\n"
	str += "		Number: " + strconv.FormatInt(d.Head.Number, 10) + ",\n"
	str += "		Time: " + strconv.FormatInt(d.Head.Time, 10) + ",\n"
	str += "		Witness: " + d.Head.Witness + ",\n"
	str += "	}\n"

	str += "	Txs {\n"
	//for _, tx := range d.Txs {
	//	//str += tx.String()
	//}
	str += "	}\n"
	str += "	Receipts {\n"
	//for _, receipt := range d.Receipts {
	//	str += receipt.String()
	//}
	str += "	}\n"
	str += "}\n"
	return str
}
*/

func (d *Block) CalculateTxsHash() []byte {
	treeHash := make([]byte, 0)
	for _, tx := range d.Txs {
		treeHash = append(treeHash, tx.Publisher.Sig...)
	}
	return common.Sha256(treeHash)
}

func (d *Block) CalculateMerkleHash() []byte {
	return nil
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
	br := &BlockRaw{
		Head:     &d.Head,
		Txs:      txs,
		Receipts: rpts,
	}

	b, err := proto.Marshal(br)
	if err != nil {
		panic(err)
	}
	d.hash = nil
	return b
}

func (d *Block) Decode(bin []byte) (err error) {
	br := &BlockRaw{}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	err = proto.Unmarshal(bin, br)
	d.Head = *br.Head
	for _, t := range br.Txs {
		var tt tx.Tx
		err = tt.Decode(t)
		if err != nil {
			return err
		}
		d.Txs = append(d.Txs, &tt)
	}
	for _, r := range br.Receipts {
		var rcpt tx.TxReceipt
		err = rcpt.Decode(r)
		if err != nil {
			return err
		}
		d.Receipts = append(d.Receipts, &rcpt)
	}
	return nil
}

func (d *Block) HashID() string {
	id := d.Head.Witness +
		strconv.FormatInt(d.Head.Time, 10) +
		strconv.FormatInt(d.Head.Number, 10) +
		strconv.FormatInt(d.Head.Version, 10)
	return id
}

func (d *Block) HeadHash() []byte {
	if d.hash == nil {
		d.hash = d.Head.Hash()
	}
	return d.hash
}

func (d *Block) GetTx(x int) *tx.Tx {
	if x < len(d.Txs) {
		return d.Txs[x]
	} else {
		return &tx.Tx{}
	}
}

func (d *Block) LenTx() int {
	return len(d.Txs)
}

func (d *BlockHead) Encode() []byte {
	bin, err := proto.Marshal(d)
	if err != nil {
		panic(err)
	}
	return bin
}

func (d *BlockHead) Decode(bin []byte) error {
	err := proto.Unmarshal(bin, d)
	return err
}

func (d *BlockHead) Hash() []byte {
	return common.Sha256(d.Encode())
}
