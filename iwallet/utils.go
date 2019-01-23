package iwallet

import (
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/rpc/pb"
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
*/

func saveTo(output string, data []byte) error {
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	defer f.Close()
	return err
}

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

func actionsFromFlags(args []string) ([]*rpcpb.Action, error) {
	argc := len(args)
	if argc%3 != 0 {
		return nil, fmt.Errorf(`number of args should be a multiplier of 3`)
	}
	var actions = make([]*rpcpb.Action, 0)
	for i := 0; i < len(args); i += 3 {
		act := NewAction(args[i], args[i+1], args[i+2]) //check sth here
		actions = append(actions, act)
	}
	return actions, nil
}
