package message

import "github.com/gogo/protobuf/proto"

func (d *RequestHeight) Encode() []byte {
	b, err := proto.Marshal(d)
	if err != nil {
		panic(err)
	}

	return b
}

func (d *RequestHeight) Decode(bin []byte) error {
	err := proto.Unmarshal(bin, d)
	if err != nil {
		return err
	}

	return nil
}

func (d *ResponseHeight) Encode() []byte {
	b, err := proto.Marshal(d)
	if err != nil {
		panic(err)
	}

	return b
}

func (d *ResponseHeight) Decode(bin []byte) error {
	err := proto.Unmarshal(bin, d)
	if err != nil {
		return err
	}

	return nil
}

func (d *RequestBlock) Encode() []byte {
	b, err := proto.Marshal(d)
	if err != nil {
		panic(err)
	}

	return b
}

func (d *RequestBlock) Decode(bin []byte) error {
	err := proto.Unmarshal(bin, d)
	if err != nil {
		return err
	}

	return nil
}
