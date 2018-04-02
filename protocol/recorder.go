package protocol

import (
	"fmt"
	"time"

	"github.com/iost-official/PrototypeWorks/iosbase"
)

//go:generate mockgen -destination recorder_mock_test.go -package protocol -source recorder.go

type Recorder interface {
	Init(self iosbase.Member, db Database, router Router) error
	Run()
	Stop()
}

func RecorderFactory(kind string) (Recorder, error) {
	switch kind {
	case "base1.0":
		rec := RecorderImpl{}
		return &rec, nil
	}
	return nil, fmt.Errorf("target recorder not found")
}

const (
	RecorderPeriod = time.Minute
)

type RecorderImpl struct {
	Database
	iosbase.Member

	view View

	chView      chan View
	chTx, chBlk chan iosbase.Request
	chReply     chan iosbase.Response
	chSend      chan iosbase.Request
	chReceive   chan iosbase.Response
	ExitSignal  chan bool

	txPool iosbase.TxPool
}

func (r *RecorderImpl) Init(self iosbase.Member, db Database, router Router) error {
	r.Database = db
	r.Member = self

	var err error

	r.chView, err = db.NewViewSignal()
	if err != nil {
		return err
	}

	r.chTx, err = router.FilteredInChan(Filter{
		AcceptType: []ReqType{ReqPublishTx},
	})
	if err != nil {
		return err
	}

	r.chBlk, err = router.FilteredInChan(Filter{
		AcceptType: []ReqType{ReqNewBlock},
	})

	r.chReply, err = router.ReplyChan()
	if err != nil {
		return err
	}

	r.chSend, r.chReceive, err = router.SendChan()
	if err != nil {
		return err
	}

	return nil
}

func (r *RecorderImpl) Run() {
	go r.recorderLoop()
	go r.txLoop()
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
			err = r.VerifyTxWithCache(tx, r.txPool)
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

						r.chSend <- req
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
			err = r.VerifyBlockWithCache(&blk, r.txPool)
			if err != nil {
				r.chReply <- illegalTx(req)
				continue
			}
			r.PushBlock(&blk)
		}
	}
}
