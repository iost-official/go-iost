package message

import "github.com/gogo/protobuf/proto"

func (d *RequestBlock) Encode() ([]byte, error) {
	b, err := proto.Marshal(d)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (d *RequestBlock) Decode(bin []byte) error {
	err := proto.Unmarshal(bin, d)
	if err != nil {
		return err
	}

	return nil
}

func (d *BlockHashQuery) Encode() ([]byte, error) {
	b, err := proto.Marshal(d)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (d *BlockHashQuery) Decode(bin []byte) error {
	err := proto.Unmarshal(bin, d)
	if err != nil {
		return err
	}

	return nil
}

func (d *BlockHashResponse) Encode() ([]byte, error) {
	b, err := proto.Marshal(d)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (d *BlockHashResponse) Decode(bin []byte) error {
	err := proto.Unmarshal(bin, d)
	if err != nil {
		return err
	}

	return nil
}
