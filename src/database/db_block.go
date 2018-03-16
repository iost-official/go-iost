package database

import (
	"fmt"
	"common"
)

type fork_database

type ChainDatabase struct{
	_fork_db fork_database
	_block_id_to_block block_database

}




func (db *ChainDatabase) is_known_block(id common.BlockIdType) bool {
	return _fork_db.is_known_block(id)|| _block_id_to_block.contains(id)
}
