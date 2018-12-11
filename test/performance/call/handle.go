package call

import (
	"fmt"
	"log"
)

// Handler ...
type Handler interface {
	Prepare() error
	Run(i int) (interface{}, error)
}

// Handle ...
func Handle(handler Handler, num int, results chan<- string) {
	r, err := handler.Run(num)

	results <- fmt.Sprintf("res:%v, err:%v", r, err)
}

// Display ...
func Display(results chan string, flag bool) {
	for result := range results {
		if flag {
			log.Println("result:", result)
		}
	}
}
