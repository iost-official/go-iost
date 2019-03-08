package synchro

import "github.com/iost-official/go-iost/metrics"

var (
	neighborHeightGauge       = metrics.NewGauge("iost_synchro_neighbor_height", []string{})
	blockHashSyncTimeGauge    = metrics.NewGauge("iost_synchro_blockhash_sync_time", []string{})
	blockSyncTimeGauge        = metrics.NewGauge("iost_synchro_block_sync_time", []string{})
	incommingBlockBufferGauge = metrics.NewGauge("iost_synchro_incomming_block_buffer", []string{})
)
