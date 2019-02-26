package synchronizer

import "github.com/iost-official/go-iost/metrics"

var (
	verifyBufferLen      = metrics.NewGauge("iost_sync_verify_buffer_len", nil)
	responseBufferLen    = metrics.NewGauge("iost_sync_response_buffer_len", nil)
	waitMissionCount     = metrics.NewGauge("iost_sync_wait_count", nil)
	downloadMissionCount = metrics.NewGauge("iost_sync_download_count", nil)
	doneBlockCount       = metrics.NewCounter("iost_sync_done_block_count", nil)
	timeoutBlockCount    = metrics.NewCounter("iost_sync_timeout_block_count", nil)
	responseBlockCount   = metrics.NewCounter("iost_sync_response_block_count", nil)
	requestBlockCount    = metrics.NewCounter("iost_sync_request_block_count", nil)
)
