package p2p

type Request struct {
	Time    int64  // 发送时的时间戳
	From    string // From To是钱包地址的base58编码字符串（就是Member.ID，下同）
	To      string
	ReqType int // 此request的类型码，通过类型可以确定body的格式以方便解码body
	Body    []byte
}

type Response struct {
	From        string
	To          string
	Code        int // http-like的状态码和描述
	Description string
}

// 最基本网络的模块API，之后gossip协议，虚拟的网络延迟都可以在模块内部实现
type Network interface {
	Send(req Request) chan Response
	Listen(port uint16) (chan Request, chan Response, error)
	Close(port uint16) error
}
