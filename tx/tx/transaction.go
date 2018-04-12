package transaction

import (
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/iost-official/prototype/tx/min_framework"
	"strings"
)


// Transaction 由交易 ID，输入和输出构成
type Transaction struct {
	ID []byte
	Vin []TXInput
	Vout []TXOutput
}


// Check whether the transaction is coinbase
// Coinbase: Vin[0].Txid == "" and the index of Vin[0]'s Vout is -1.
func (tx *Transaction) IsCoinbase() bool{
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}


func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}


// NewCoinbaseTX 构建 coinbase 交易，该没有输入，只有一个输出
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{min_framework.Subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}

// NewUTXOTransaction 创建一笔新的交易
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) (*Transaction, string) {
	var inputs []TXInput
	var outputs []TXOutput

	// 找到足够的未花费输出
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		return nil, "ERROR: Not enough funds\n"
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TXOutput{amount, to})

	// 如果 UTXO 总数超过所需，则产生找零
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx, ""
}


// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       ScriptSig: %x", input.ScriptSig))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       ScriptPubKey: %x", output.ScriptPubKey))
	}

	return strings.Join(lines, "\n")
}