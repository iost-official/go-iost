package cmd

import (
	"io/ioutil"
	"os"
)

func ReadSourceFile(file string) (code string) {
	buf, err := ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(buf)
}

func ReadFile(src string) ([]byte, error) {
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
