package trie

import (
	"github.com/iost-official/prototype/common"
	"github.com/go-ethereum/crypto"
)

var (
	// 规定树为空时根节点的哈希值
	emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

	// 规定空state树的哈希值
	emptyState = crypto.Keccak256Hash(nil)
)

// 遍历到trie叶节点时的回调函数
type LeafCallback func(leaf []byte, parent common.Hash) error

// 空值用空树表示，不存储在数据库中
// New 可以创建一个Trie，会存储在数据库的顶端
// 线程不安全
type Trie struct {
	db           *Database
	root         node
	originalRoot common.Hash

	// 每commit一次，版本号+1
	cachegen, cachelimit uint16
}

// 设置cache保存的哈希版本数
func (t *Trie) SetCacheLimit(l uint16) {
	t.cachelimit = l
}

func (t *Trie) newFlag() nodeFlag {
	return nodeFlag{ dirty: true, gen: t.cachegen }
}

func New(root common.Hash, db *Database) (*Trie, error) {
	if db == nil {
		panic("trie.New called without a database")
	}
	trie := &Trie{
		db: db,
		originalRoot: root,
	}
	if (root != common.Hash{}) && root != emptyRoot {
		rootnode, err := trie
	}
}

func (t *Trie) resolveHash(n hashNode, prefix []byte) (node, error) {
	hash := common.BytesToHash(n)

	enc, err := t.db.Node(hash)
	if err != nil || enc == nil {
		return nil, &MissingNodeError{}
	}
}