package core

type Serializable interface {
	Encode() []byte
	Decode([]byte) error
	Hash() []byte
}
