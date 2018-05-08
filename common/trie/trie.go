package trie

import (
	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/common"
	"github.com/go-ethereum/crypto"
)

var (
	// 规定树为空时根节点的哈希值
	emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

	// 规定
	emptyState = crypto.Keccak256Hash(nil)
)

type Trie struct {
	db *db.Database
}