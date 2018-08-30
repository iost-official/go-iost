package main

import (
	"fmt"
	"os"
	"os/exec"
)

var (
	transCommand = exec.Command(os.Getenv("GOPATH") + `/src/github.com/iost-official/Go-IOS-Protocol/stress/test.sh`)
)

func transfer() {
	err := transCommand.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	transfer()
}
