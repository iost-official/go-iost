package iosbase

func (d * State) Encode() []byte {
	bin, err := d.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return bin
}

func (d * State) Decode(bin []byte) error {
	_,err := d.Unmarshal(bin)
	return err
}
func (d * State) Hash() []byte {
	return Sha256(d.Encode())
}
