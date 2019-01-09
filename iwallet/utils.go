package iwallet

import (
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/common"
	"io/ioutil"
	"os"
)

func loadBytes(s string) []byte {
	if s[len(s)-1] == 10 {
		s = s[:len(s)-1]
	}
	buf := common.Base58Decode(s)
	return buf
}

/*
func saveBytes(buf []byte) string {
	return common.Base58Encode(buf)
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
*/

func loadKey(src string) ([]byte, error) {
	fi, err := os.Open(src)
	defer fi.Close()
	if err != nil {
		return nil, err
	}
	fileinfo, err := fi.Stat()
	if err != nil {
		return nil, err
	}
	if fileinfo.Mode() != 0400 {
		return nil, fmt.Errorf("private key file should have read only permission. try:\n chmod 0400 %v", src)
	}
	fd, err := ioutil.ReadAll(fi)
	if err != nil {
		return nil, err
	}
	return fd, nil
}

func readFile(src string) ([]byte, error) {
	fi, err := os.Open(src)
	defer fi.Close()
	if err != nil {
		return nil, err
	}
	fd, err := ioutil.ReadAll(fi)
	if err != nil {
		return nil, err
	}
	return fd, nil
}

func marshalTextString(pb proto.Message) string {
	m := jsonpb.Marshaler{}
	m.EmitDefaults = true
	m.Indent = "    "
	r, err := m.MarshalToString(pb)
	if err != nil {
		return "json.Marshal error: " + err.Error()
	}
	return r
}
