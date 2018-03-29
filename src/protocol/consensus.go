package protocol

import (
	"IOS/src/iosbase"
	"sync"
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
	Version    = 1
	Port       = 12306
	ExpireTime = 1 * time.Minute
	Period     = 5 * time.Minute
)

type Consensus struct {
	Recorder
	Replica
	NetworkFilter
}

func (c *Consensus) Init(bc iosbase.BlockChain, sp iosbase.StatePool, network iosbase.Network) error {
	rd := RuntimeData{}

	err := c.NetworkFilter.init(&rd, network)
	if err != nil {
		return err
	}

	c.Recorder, err = RecorderFactory("base1.0")
	err = c.Recorder.Init(&rd, &c.NetworkFilter, bc, sp)
	if err != nil {
		return err
	}

	c.Replica, err = ReplicaFactory("base1.0")
	if err != nil {
		return err
	}
	c.Replica.Init(&rd, &c.NetworkFilter, c.Recorder)
	return err
}

func (c *Consensus) Run() {
	req, res, err := c.base.Listen(Port)
	if err != nil {
		panic(err)
	}
	defer c.base.Close(Port)

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()
		c.router(req)
	}()
	go func() {
		wg.Add(1)
		defer wg.Done()
		c.ReplicaLoop()
	}()
	go func() {
		wg.Add(1)
		defer wg.Done()
		c.replicaFilter(c.Replica, res)
	}()
	go func() {
		wg.Add(1)
		defer wg.Done()
		c.RecorderLoop()
	}()
	go func() {
		wg.Add(1)
		defer wg.Done()
		c.recorderFilter(c.Recorder, res)
	}()

	wg.Wait()
}

func (c *Consensus) Stop() {
	c.isRunning = false
}
func (c *Consensus) PublishTx(tx iosbase.Tx) error {
	return c.Recorder.PublishTx(tx)
}
func (c *Consensus) CheckTx(tx iosbase.Tx) (iosbase.TxStatus, error) {
	return iosbase.POOL, nil // TODO not complete
}
func (c *Consensus) GetStatus() (iosbase.BlockChain, iosbase.StatePool, error) {
	return c.blockChain, c.statePool, nil
}
func (c *Consensus) GetCachedStatus() (iosbase.BlockChain, iosbase.StatePool, error) {
	return c.blockChain, c.statePool, nil
}
