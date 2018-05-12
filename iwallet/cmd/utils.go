package cmd

import "encoding/base64"

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
