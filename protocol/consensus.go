/*
 The protocol of iost consensus
*/
package protocol

import (
	"github.com/iost-official/PrototypeWorks/iosbase"
)

type Character int

const (
	Primary Character = iota
	Backup
	Idle
)

const (
	Version = 1
	Port    = 12306
)

type ConsensusImpl struct {
	iosbase.Member

	recorder, replica component
	db                Database
	router            Router
}

func (c *ConsensusImpl) Init(bc iosbase.BlockChain, sp iosbase.StatePool, network iosbase.Network) error {
	var err error

	c.db, err = DatabaseFactory("base", bc, sp)
	if err != nil {
		return err
	}

	c.router, err = RouterFactory("base")
	if err != nil {
		return err
	}
	err = c.router.Init(network, Port)
	if err != nil {
		return err
	}

	c.recorder, err = RecorderFactory("base")
	if err != nil {
		return err
	}
	err = c.recorder.Init(c.Member, c.db, c.router)
	if err != nil {
		return err
	}

	pool := &iosbase.TxPoolImpl{}
	c.replica, err = ReplicaFactory("pbft", pool)
	if err != nil {
		return err
	}
	err = c.replica.Init(c.Member, c.db, c.router)
	if err != nil {
		return err
	}
	return nil
}

func (c *ConsensusImpl) Run() {
	go c.router.Run()
	go c.replica.Run()
	go c.recorder.Run()
}

func (c *ConsensusImpl) Stop() {
	c.replica.Stop()
	c.recorder.Stop()
	c.router.Stop()
}

func (c *ConsensusImpl) PublishTx(tx iosbase.Tx) error {
	req := iosbase.Request{
		From:    c.ID,
		To:      c.ID,
		ReqType: int(ReqPublishTx),
		Body:    tx.Encode(),
	}
	c.router.Send(req)
	return nil
}

func (c *ConsensusImpl) CheckTx(tx iosbase.Tx) (iosbase.TxStatus, error) {
	return iosbase.POOL, nil // TODO not complete
}

func (c *ConsensusImpl) GetStatus() (iosbase.BlockChain, iosbase.StatePool, error) {
	bc, err := c.db.GetBlockChain()
	if err != nil {
		return nil, nil, err
	}
	sp, err := c.db.GetStatePool()
	if err != nil {
		return nil, nil, err
	}
	return bc, sp, nil
}

func (c *ConsensusImpl) GetCachedStatus() (iosbase.BlockChain, iosbase.StatePool, error) {
	return c.GetStatus()
}
