package chain

import (
	"chain/protocol"
	"common"
)




type forkItem struct {
	num     uint32 // initialized in ctor
	prev    *forkItem
	invalid bool
	id      common.BlockIdType
	data    protocol.SignedBlock
	/**
	 * isvalid:
	 * Used to flag a block as invalid and prevent other blocks from
	 * building on top of it.
	 */
}

/**
*  As long as blocks are pushed in order the fork
*  database will maintain a linked tree of all blocks
*  that branch from the start_block.  The tree will
*  have a maximum depth of 1024 blocks after which
*  the database will start lopping off forks.
*
*  Every time a block is pushed into the fork DB the
*  block with the highest block_num will be returned.
*/
const MAX_BLOCK_REORDERING int =1024

type ForkDatabase struct {
	_index
	_head     *forkItem
	_max_size uint32
}
