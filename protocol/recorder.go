package protocol

import (
	"fmt"
	"time"

	"github.com/iost-official/PrototypeWorks/iosbase"
)

func RecorderFactory(target string) (Component, error) {
	switch target {
	case "basic":
		rec := RecorderImpl{}
		return &rec, nil
	}
	return nil, fmt.Errorf("target recorder not found")
}

type RecorderImpl struct {
	iosbase.Member

	db   Database
	net  Router
	view View

	chView      chan View
	chTx, chBlk chan iosbase.Request
	chReply     chan iosbase.Response
	ExitSignal  chan bool

	txPool iosbase.TxPool
}

func (r *RecorderImpl) Init(self iosbase.Member, db Database, router Router) error {
	r.db = db
	r.net = router
	r.Member = self

	var err error

	r.chView, err = db.NewViewSignal()
	if err != nil {
		return err
	}

	r.chTx, r.chReply, err = router.FilteredChan(Filter{
		AcceptType: []ReqType{ReqPublishTx},
	})
	if err != nil {
		return err
	}

	r.chBlk, _, err = router.FilteredChan(Filter{
		AcceptType: []ReqType{ReqNewBlock},
	})

	return nil
}

func (r *RecorderImpl) Run() {
	go r.recorderLoop()
	go r.txLoop()
	go r.blockLoop()
}

func (r *RecorderImpl) Stop() {
	r.ExitSignal <- true
}

func (r *RecorderImpl) txLoop() {
	for true {
		select {
		case <-r.ExitSignal:
			return
		case req := <-r.chTx:
			var tx iosbase.Tx
			err := tx.Decode(req.Body)
			if err != nil {
				r.chReply <- syntaxError(req)
				continue
			}
			err = r.db.VerifyTxWithCache(tx, r.txPool)
			if err != nil {
				r.chReply <- illegalTx(req)
				continue
			}
			r.txPool.Add(tx)
		}
	}
}

func (r *RecorderImpl) recorderLoop() {
	for true {
		select {
		case <-r.ExitSignal:
			return
		case r.view = <-r.chView:
			if r.txPool.Size() > 0 {
				for _, m := range r.view.GetBackup() {
					go func() {
						req := iosbase.Request{
							From:    r.ID,
							To:      m.ID,
							ReqType: int(ReqSubmitTxPack),
							Body:    r.txPool.Encode(),
							Time:    time.Now().Unix(),
						}

						r.net.Send(req)
					}()
				}
			}
		}
	}
}

func (r *RecorderImpl) blockLoop() {
	for true {
		select {
		case <-r.ExitSignal:
			return
		case req := <-r.chBlk:
			var blk iosbase.Block
			err := blk.Decode(req.Body)
			if err != nil {
				r.chReply <- syntaxError(req)
				continue
			}
			err = r.db.VerifyBlockWithCache(&blk, r.txPool)
			if err != nil {
				r.chReply <- illegalTx(req)
				continue
			}
			r.db.PushBlock(&blk)
		}
	}
}
