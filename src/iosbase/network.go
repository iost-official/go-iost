package iosbase

type Request struct {
	Time    int64
	From    string
	To      string
	ReqType int
	Body    []byte
}

type Response struct {
	From        string
	To          string
	Code        int
	Description string
}

type Network interface {
	Send(req Request) chan Response
	Listen(port uint16) (chan Request, chan Response, error)
	Close(port uint16) error
}
