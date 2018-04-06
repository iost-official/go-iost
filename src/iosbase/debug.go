package debug

import (
	"fmt"
)
//will not terminated once assertion not satisfied
func assert(a bool, message string){
	if a {
		return
	} else {
		fmt.Println(message)
	}
}
