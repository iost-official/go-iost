package message

func (d *RequestHeight) Encode() []byte {
	b, err := d.Marshal(nil)
	if err != nil {
		panic(err)
	}

	return b
}

func (d *RequestHeight) Decode(bin []byte) error {
	_, err := d.Unmarshal(bin)
	if err != nil {
		return err
	}

	return nil
}

func (d *ResponseHeight) Encode() []byte {
	b, err := d.Marshal(nil)
	if err != nil {
		panic(err)
	}

	return b
}

func (d *ResponseHeight) Decode(bin []byte) error {
	_, err := d.Unmarshal(bin)
	if err != nil {
		return err
	}

	return nil
}

func (d *RequestBlock) Encode() []byte {
	b, err := d.Marshal(nil)
	if err != nil {
		panic(err)
	}

	return b
}

func (d *RequestBlock) Decode(bin []byte) error {
	_, err := d.Unmarshal(bin)
	if err != nil {
		return err
	}

	return nil
}
