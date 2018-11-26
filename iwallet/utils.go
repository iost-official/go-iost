package iwallet

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/common"
	"io/ioutil"
	"os"
)

func saveBytes(buf []byte) string {
	return common.Base58Encode(buf)
}

func loadBytes(s string) []byte {
	if s[len(s)-1] == 10 {
		s = s[:len(s)-1]
	}
	buf := common.Base58Decode(s)
	return buf
}

/*
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
