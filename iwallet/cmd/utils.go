package cmd

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/iost-official/Go-IOS-Protocol/common"
)

func saveBytes(buf []byte) string {
	return common.Base58Encode(buf)
}

func loadBytes(s string) []byte {
	buf := common.Base58Decode(s)
	return buf
}

func changeSuffix(filename, suffix string) string {
	dist := filename[:strings.LastIndex(filename, ".")]
	dist = dist + suffix
	return dist
}

func saveTo(Dist string, file []byte) error {
	f, err := os.Create(Dist)
	if err != nil {
		return err
	}
	_, err = f.Write(file)
	defer f.Close()
	return err
}

func readFile(src string) ([]byte, error) {
	fi, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	if err != nil {
		return nil, err
	}
	return fd, nil
}
