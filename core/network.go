package core

//go:generate mockgen -destination mocks/mock_network.go -package core_mock github.com/iost-official/prototype/core Network

type Network interface {
	Send(req Request)
	Listen(port uint16) (<-chan Request, error)
	Close(port uint16) error
}
