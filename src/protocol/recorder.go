package protocol

import (
	"IOS/src/iosbase"
	"fmt"
	"time"
)

//go:generate mockgen -destination recorder_mock_test.go -package protocol -source recorder.go

type Recorder interface {
	Init(rd *RuntimeData, nw *NetworkFilter) error
	RecordTx(tx iosbase.Tx) error
	RecorderLoop()
	SendTxPack(member iosbase.Member) error
}

func RecorderFactory(kind string) (Recorder, error) {
	switch kind {
	case "base1.0":
		rep := RecorderImpl{}
		return &rep, nil
	}
	return nil, fmt.Errorf("target recorder not found")
}

type RecorderImpl struct {
	*RuntimeData
	network  *NetworkFilter
	replicas []iosbase.Member

	txPool iosbase.TxPool
}

func (r *RecorderImpl) Init(rd *RuntimeData, nw *NetworkFilter) error {
	r.RuntimeData = rd
	r.network = nw
	return nil
}

func (r *RecorderImpl) RecordTx(tx iosbase.Tx) error {
	if err := r.VerifyTx(tx); err != nil {
		return err
	}
	r.txPool.Add(tx)
	return nil
}

func (r *RecorderImpl) SendTxPack(member iosbase.Member) error {
	req := iosbase.Request{
		From:    r.ID,
		To:      member.ID,
		ReqType: int(ReqSubmitTxPack),
		Body:    r.txPool.Encode(),
		Time:    time.Now().Unix(),
	}

	r.network.Send(req)
	return nil
}

func (r *RecorderImpl) RecorderLoop() {
	to := time.NewTimer(Period)
	for true {
		select {
		case <-r.ExitSignal:
			return
		case <-to.C:
			if r.txPool.Size() > 0 {
				for _, m := range r.view.GetBackup() {
					r.SendTxPack(m)
				}
			}
			to.Reset(Period)
		}
	}
}
