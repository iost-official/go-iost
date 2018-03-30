package protocol

//import (
//	"IOS/src/iosbase"
//	"fmt"
//)
//
////
//
//type MockNetworkBase struct {
//	reqChan chan iosbase.Request
//	resChan chan iosbase.Response
//}
//
//func NewNetworkBase() MockNetworkBase {
//	var req = make(chan iosbase.Request)
//	var res = make(chan iosbase.Response)
//	return MockNetworkBase{req, res}
//}
//func (m *MockNetworkBase) Send(req iosbase.Request) chan iosbase.Response {
//	fmt.Println("===== Send req:")
//	fmt.Println(req)
//	return m.resChan
//}
//func (m *MockNetworkBase) Listen(port uint16) (chan iosbase.Request, chan iosbase.Response, error) {
//
//	return m.reqChan, m.resChan, nil
//}
//func (m *MockNetworkBase) Close(port uint16) error {
//	return nil
//}
//func (m *MockNetworkBase) run() {
//	for true {
//		res := <-m.resChan
//		fmt.Println("----- receive res:", res)
//	}
//}
