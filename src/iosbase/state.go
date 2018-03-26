package iosbase

func (s * State) Encode() []byte {
	bin, err := s.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return bin
}

func (s * State) Decode(bin []byte) error {
	_,err := s.Unmarshal(bin)
	return err
}
func (s * State) Hash() []byte {
	return Sha256(s.Encode())
}
