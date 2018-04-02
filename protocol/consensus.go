package protocol

import (
	"sync"
	"time"

	"github.com/iost-official/PrototypeWorks/iosbase"
)

type Character int

const (
	Primary Character = iota
	Backup
	Idle
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
	DataHolder
	NetworkFilter
	RuntimeData
}

func (c *Consensus) Init(bc iosbase.BlockChain, sp iosbase.StatePool, network iosbase.Network) error {
	c.RuntimeData = RuntimeData{}

	err := c.NetworkFilter.Init(&c.RuntimeData, network, Port)
	if err != nil {
		return err
	}

	c.Recorder, err = RecorderFactory("base1.0")
	err = c.Recorder.Init(&c.RuntimeData, &c.NetworkFilter)
	if err != nil {
		return err
	}

	c.Replica, err = ReplicaFactory("base1.0")
	if err != nil {
		return err
	}
	err = c.Replica.Init(&c.RuntimeData, &c.NetworkFilter)
	if err != nil {
		return err
	}

	c.DataHolder, err = DataHolderFactory("base1.0")
	if err != nil {
		return err
	}
	err = c.DataHolder.Init(&c.RuntimeData, &c.NetworkFilter)

	return err
}

func (c *Consensus) Run() {

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()
		c.Router(c.Replica, c.Recorder, c.DataHolder)
	}()
	go func() {
		wg.Add(1)
		defer wg.Done()
		c.ReplicaLoop()
	}()

	wg.Wait()
}

func (c *Consensus) Stop() {
	c.ExitSignal <- true
	c.base.Close(Port)
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
