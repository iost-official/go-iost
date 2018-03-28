package protocol

import (
	"IOS/src/iosbase"
	"time"
)

type Character int

const (
	Primary Character = iota
	Backup
	Idle
)

type Phase int

const (
	StartPhase Phase = iota
	PrePreparePhase
	PreparePhase
	CommitPhase
	PanicPhase
	EndPhase
)

const (
	ReplicaPort  = 12306
	RecorderPort = 12307
	ExpireTime   = 1 * time.Minute
	Period       = 5 * time.Minute
)

type Consensus struct {
	Recorder
	Replica
	NetworkFilter
}

func (c *Consensus) Init(bc iosbase.BlockChain, sp iosbase.StatePool, network iosbase.Network) error {
	rd := RuntimeData{}
	err := c.Recorder.init(&rd, bc, sp)
	if err != nil {
		return err
	}
	err = c.NetworkFilter.init(&rd, network)
	c.Replica.init(&rd, &c.NetworkFilter, &c.Recorder)
	return err
}

func (c *Consensus) Run() {
	go c.replicaLoop()
	rawReq, rawRes, err := c.base.Listen(ReplicaPort)
	if err != nil {
		panic(err)
	}
	defer c.base.Close(ReplicaPort)
	go c.replicaFilter(rawReq, rawRes)

	req2, res2, err := c.base.Listen(RecorderPort)
	if err != nil {
		panic(err)
	}
	defer c.base.Close(RecorderPort)
	go c.recorderFilter(c.Recorder, req2, res2)

}

func (c *Consensus) Stop() {
	c.isRunning = false
}
func (c *Consensus) PublishTx(tx iosbase.Tx) error {
	err := c.verifyTx(tx)
	if err != nil {
		return err
	}
	tx.Recorder = c.ID
	c.txPool.Add(tx)
	return nil
}
func (c *Consensus) CheckTx(tx iosbase.Tx) (iosbase.TxStatus, error) {
	return iosbase.POOL, nil
}
func (c *Consensus) GetStatus() (iosbase.BlockChain, iosbase.StatePool, error) {
	return c.blockChain, c.statePool, nil
}
func (c *Consensus) GetCachedStatus() (iosbase.BlockChain, iosbase.StatePool, error) {
	return c.blockChain, c.statePool, nil
}
