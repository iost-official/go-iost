package network

type PbftNetwork struct {
}

func (pn *PbftNetwork) Send(request Request, response chan Response) {

}

func (pn *PbftNetwork) SendSync(request Request) Response {
	return nil
}

type PbftRequestType int
const (
	PBFT_PrePrepare PbftRequestType = iota
	PBFT_Prepare
	PBFT_Commit
	PBFT_PullBlock
)

type PbftRequest struct {
	From    string
	To      string
	ReqType PbftRequestType
	Body    []byte
	Time	int64
}



type PbftResponseState int
const (
	Accepted PbftResponseState = iota
	Reject
	Error
)

type PbftResponse struct {
	Time     int64
	Response PbftResponseState
}


