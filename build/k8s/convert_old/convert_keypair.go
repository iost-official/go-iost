package main

import (
	"bufio"
	"fmt"
	"github.com/iost-official/go-iost/common"
	"os"
	"strings"
)

func main() {
	f := os.Args[1]
	fout := os.Args[2]
	file, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	out, err := os.Create(fout)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		l := scanner.Text()
		oldKey := strings.Split(l, ",")[0]
		b := common.Base58Decode(oldKey[4:])
		_, err := out.WriteString(fmt.Sprintf("%s,%s\n", common.Base58Encode(b[:len(b)-4]), strings.Split(l, ",")[1]))
		if err != nil {
			panic(err)
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
