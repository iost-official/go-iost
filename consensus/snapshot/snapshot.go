package snapshot

import (
	"encoding/json"
	"fmt"

	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/db"
)

func SaveBlockHead(db db.MVCCDB, bh *block.BlockHead) error {
	bhJSON, err := json.Marshal(bh)
	if err != nil {
		return fmt.Errorf("json fail: %v", err)
	}
	err = db.Put("snapshot", "blockHead", string(bhJSON))
	if err != nil {
		return fmt.Errorf("state db put fail: %v", err)
	}
	return nil
}
