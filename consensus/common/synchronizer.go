package consensus_common

import (
	"github.com/iost-official/prototype/core/message"
	. "github.com/iost-official/prototype/network"
)

type Synchronizer interface {
	StartListen() error
	RunSync() error
}

type SyncImpl struct {
	blockCache BlockCache
	router	Router
	heightChan chan message.Message
	blkSyncChain chan message.Message
}

func NewSynchronizer(bc BlockCache, router Router) *SyncImpl {
	sync := &SyncImpl{
		blockCache: bc,
		router: router,
	}
	var err error
	sync.heightChan, err = sync.router.FilteredChan(Filter{
		WhiteList:  []message.Message{},
		BlackList:  []message.Message{},
		RejectType: []ReqType{},
		AcceptType: []ReqType{
			ReqHeight,
		}})
	if err != nil {
		return nil
	}

	sync.blkSyncChain, err = sync.router.FilteredChan(Filter{
		WhiteList:  []message.Message{},
		BlackList:  []message.Message{},
		RejectType: []ReqType{},
		AcceptType: []ReqType{
			ReqBlockSync,
		}})
	if err != nil {
		return nil
	}
	return sync
}

type HeightRequest struct {
	localHeight int
	needHeight  int
}

type HeightResponse struct {
	height  int
}

type BlockRequest struct {
	number  int
}

func (sync *SyncImpl) StartListen() error {
	return nil
}

func (sync *SyncImpl) RunSync() error {
	return nil
}
