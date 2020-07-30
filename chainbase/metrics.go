package chainbase

import (
	"time"

	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/metrics"
)

var (
	metricsTxTotal = metrics.NewGauge("iost_tx_total", nil)
	metricsDBSize  = metrics.NewGauge("iost_db_size", []string{"Name"})
)

func (c *ChainBase) doMetricsUpdate() {
	metricsTxTotal.Set(float64(c.bChain.TxTotal()), nil)

	if blockchainDBSize, err := c.bChain.Size(); err != nil {
		ilog.Warnf("Get BlockChainDB size failed: %v", err)
	} else {
		metricsDBSize.Set(
			float64(blockchainDBSize),
			map[string]string{
				"Name": "BlockChainDB",
			},
		)
	}

	if stateDBSize, err := c.stateDB.Size(); err != nil {
		ilog.Warnf("Get StateDB size failed: %v", err)
	} else {
		metricsDBSize.Set(
			float64(stateDBSize),
			map[string]string{
				"Name": "StateDB",
			},
		)
	}
}

func (c *ChainBase) metricsController() {
	for {
		select {
		case <-time.After(2 * time.Second):
			c.doMetricsUpdate()
		case <-c.quitCh:
			c.done.Done()
			return
		}
	}
}
