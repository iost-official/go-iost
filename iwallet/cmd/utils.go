package cmd

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"strings"
)

func SaveBytes(buf []byte) string {
	return base64.StdEncoding.EncodeToString(buf)
}

func LoadBytes(s string) []byte {
	buf, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return buf
}

func ChangeSuffix(filename, suffix string) string {
	dist := filename[:strings.LastIndex(filename, ".")]
	dist = dist + suffix
	return dist
}

func SaveTo(Dist string, file []byte) error {
	f, err := os.Create(Dist)
	if err != nil {
		return err
	}
	_, err = f.Write(file)
	defer f.Close()
	return err
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
