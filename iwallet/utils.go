package iwallet

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx/pb"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/rpc"

	"google.golang.org/grpc"
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

func getSignAlgo(algo string) crypto.Algorithm {
	switch algo {
	case "secp256k1":
		return crypto.Secp256k1
	case "ed25519":
		return crypto.Ed25519
	default:
		return crypto.Ed25519
	}
}

func getTxReceiptByTxHash(server string, txHashStr string) (*txpb.TxReceipt, error) {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := rpc.NewApisClient(conn)
	resp, err := client.GetTxReceiptByTxHash(context.Background(), &rpc.HashReq{Hash: txHashStr})
	if err != nil {
		return nil, err
	}
	//ilog.Debugf("getTxReceiptByTxHash(%v): %v", txHashStr, resp.TxReceiptRaw)
	return resp.TxReceipt, nil
}
