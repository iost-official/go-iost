package mvcc

type TrieNode struct {
	revision string
	ch       []*TrieNode
}

func (t *TrieNode) Revision() string {
	return t.revision
}

func (t *TrieNode) Get(key string, i int) (string, bool) {
	return "", false
}

func (t *TrieNode) Put(key string, value string, i int) {

}

func (t *TrieNode) Del(key string, i int) bool {
	return false
}

func (t *TrieNode) Fork() *TrieNode {
	return nil
}
