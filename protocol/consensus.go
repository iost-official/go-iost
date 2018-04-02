package protocol

import (
	. "github.com/iost-official/PrototypeWorks/iosbase"
	"sync"
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
	Member

	recorder, replica Component
	db                Database
	router            Router
}

func (c *ConsensusImpl) Init(bc BlockChain, sp StatePool, network Network) error {
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

	c.replica, err = ReplicaFactory("pbft")
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
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		c.router.Run()
		defer wg.Done()
	}()
	go func() {
		c.replica.Run()
		defer wg.Done()
	}()
	go func() {
		c.recorder.Run()
		defer wg.Done()
	}()

	wg.Wait()
}

func (c *ConsensusImpl) Stop() {
	c.replica.Stop()
	c.recorder.Stop()
	c.router.Stop()
}

func (c *ConsensusImpl) PublishTx(tx Tx) error {
	req := Request{
		From:    c.ID,
		To:      c.ID,
		ReqType: int(ReqPublishTx),
		Body:    tx.Encode(),
	}
	c.router.Send(req)
	return nil
}

func (c *ConsensusImpl) CheckTx(tx Tx) (TxStatus, error) {
	return POOL, nil // TODO not complete
}

func (c *ConsensusImpl) GetStatus() (BlockChain, StatePool, error) {
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

func (c *ConsensusImpl) GetCachedStatus() (BlockChain, StatePool, error) {
	return c.GetStatus()
}
