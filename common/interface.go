package common

// Serializable data serializable
type Serializable interface {
	Encode() []byte
	Decode([]byte) error
	Hash() []byte
}
