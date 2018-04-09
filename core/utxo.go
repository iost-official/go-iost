package core

import "github.com/iost-official/prototype/common"

func (d *UTXO) Encode() []byte {
	bin, err := d.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return bin
}

func (d *UTXO) Decode(bin []byte) error {
	_, err := d.Unmarshal(bin)
	return err
}
func (d *UTXO) Hash() []byte {
	return common.Sha256(d.Encode())
}
