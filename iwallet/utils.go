package iwallet

import (
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/rpc/pb"
	"io/ioutil"
	"os"
)

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

func writeFile(output string, data []byte) error {
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

func loadKeyPair(privKeyFile string, algo crypto.Algorithm) (*account.KeyPair, error) {
	fsk, err := loadKey(privKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read key file failed: %v", err)
	}
	return account.NewKeyPair(common.Base58Decode(string(fsk)), algo)
}

func saveProto(pb proto.Message, fileName string) error {
	r, err := (&jsonpb.Marshaler{EmitDefaults: true, Indent: "    "}).MarshalToString(pb)
	if err != nil {
		return err
	}
	return writeFile(fileName, []byte(r))
}

func loadProto(fileName string, pb proto.Message) error {
	bytes, err := readFile(fileName)
	if err != nil {
		return fmt.Errorf("load signature err %v %v", fileName, err)
	}
	err = jsonpb.UnmarshalString(string(bytes), pb)
	if err != nil {
		return fmt.Errorf("not a valid signature json %v %v", fileName, err)
	}
	return nil
}

func marshalTextString(pb proto.Message) string {
	r, err := (&jsonpb.Marshaler{EmitDefaults: true, Indent: "    "}).MarshalToString(pb)
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
