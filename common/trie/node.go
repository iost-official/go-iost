package trie

type node interface {
	fstring(string) string
	cache() (hashNode, bool)

}

type (
	fullNode struct {
		Children [17]node
		flags nodeFlag
	}
	hashNode []byte
	valueNode []byte
)

type nodeFlag struct {
	hash  hashNode
	gen   uint16
	dirty bool
}


type decodeError struct{
	what error
	stack []string
}