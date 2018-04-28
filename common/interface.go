package common

type Serializable interface {
	Encode() []byte
	Decode([]byte) error
	Hash() []byte
}
