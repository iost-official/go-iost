package pbft

import (
	"IOS/src/network"
	"fmt"
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
	StartPhase      Phase = iota
	PrePreparePhase
	PreparePhase
	CommitPhase
	PanicPhase
	EndPhase
)

type Validator interface {
	OnNewView() (Phase, error)
	OnPrePrepare(prePrepare PrePrepare) (Phase, error)
	OnPrepare(prepare Prepare) (Phase, error)
	OnCommit(commit Commit) (Phase, error)
	OnTimeOut(current Phase) (Phase, error)
}

func Run(self Validator, request chan network.PbftRequest, response chan network.PbftResponse) error {
	phase := StartPhase
	var req network.PbftRequest
	var err error = nil
	to := time.NewTimer(time.Second)

	for true {
		switch phase {
		case StartPhase:
			phase, err = self.OnNewView()
		case PrePreparePhase:
			if req.ReqType == network.PBFT_PrePrepare {
				pp := PrePrepare{}
				pp.Unmarshal(req.Body)
				phase, err = self.OnPrePrepare(pp)
			} else {
				response <- network.PbftResponse{time.Now().Unix(), network.Reject}
			}
		case PreparePhase:
			if req.ReqType == network.PBFT_Prepare {
				p := Prepare{}
				p.Unmarshal(req.Body)
				phase, err = self.OnPrepare(p)
			} else {
				response <- network.PbftResponse{time.Now().Unix(), network.Reject}
			}
		case CommitPhase:
			if req.ReqType == network.PBFT_Prepare {
				c := Commit{}
				c.Unmarshal(req.Body)
				phase, err = self.OnCommit(c)
			} else {
				response <- network.PbftResponse{time.Now().Unix(), network.Reject}
			}
		case PanicPhase:
			return err
		case EndPhase:
			return nil
		}

		if err != nil {
			fmt.Println(err)
		}

		to.Reset(time.Second)
		select {
		case <-request:
			req = <-request
			response <- network.PbftResponse{time.Now().Unix(), network.Accepted}
		case <-to.C:
			phase, err = self.OnTimeOut(phase)
			if err != nil {
				return err
			}
		}
	}
	return fmt.Errorf("unknown")
}
