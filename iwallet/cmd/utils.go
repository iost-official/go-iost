package cmd

import (
	"encoding/base64"
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
