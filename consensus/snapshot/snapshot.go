package snapshot

import (
	"encoding/json"
	"fmt"

	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/db"
)

// Save the function for saving data from snapshot.
func Save(db db.MVCCDB, blk *block.Block) error {
	bhJSON, err := json.Marshal(blk.Head)
	if err != nil {
		return fmt.Errorf("json fail: %v", err)
	}
	err = db.Put("snapshot", "blockHead", string(bhJSON))
	if err != nil {
		return fmt.Errorf("state db put fail: %v", err)
	}
	return nil
}

// Load the function for loading data from snapshot.
func Load(db db.MVCCDB) (*block.Block, error) {
	bhJSON, err := db.Get("snapshot", "blockHead")
	if err != nil {
		return nil, fmt.Errorf("get current block head from state db failed. err: %v", err)
	}
	bh := &block.BlockHead{}
	err = json.Unmarshal([]byte(bhJSON), bh)
	if err != nil {
		return nil, fmt.Errorf("block head decode failed. err: %v", err)
	}

	blk := &block.Block{Head: bh}
	return blk, blk.CalculateHeadHash()
}

// ToSnapshot the function for loading data from snapshot.
func ToSnapshot(db db.MVCCDB) error {
	return nil
}

// FromSnapshot the function for loading data from snapshot.
func FromSnapshot() error {
	return nil
}
