package iosbase

func (d *TxInput) Encode() []byte {
	bin, err := d.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return bin
}

func (d *TxInput) Decode(bin []byte) error {
	_, err := d.Unmarshal(bin)
	return err
}
func (d *TxInput) Hash() []byte {
	return Sha256(d.Encode())
}

func (d *Tx) Encode() []byte {
	bin, err := d.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return bin
}

func (d *Tx) Decode(bin []byte) error {
	_, err := d.Unmarshal(bin)
	return err
}
func (d *Tx) Hash() []byte {
	return Sha256(d.Encode())
}
