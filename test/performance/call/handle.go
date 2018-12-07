package call

import "log"

// Handler ...
type Handler interface {
	Init(add string, conNum int) error
	Publish() error
	Transfer(i int) string
}

// Handle ...
func Handle(handler Handler, num int, results chan<- string) {
	r := handler.Transfer(num)

	results <- r
}

// Display ...
func Display(results chan string, flag bool) {
	for result := range results {
		if flag {
			log.Println("result:", result)
		}
	}
}
