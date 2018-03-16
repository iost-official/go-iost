package network

type Sender interface {
	Send(request Request, response chan Response)
	SendSync(request Request) Response
}

type Router interface {
	OnReceive(request Request) Response
}

type Request interface {
}

type Response interface {
}
