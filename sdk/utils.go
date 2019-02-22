package sdk

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
	"strings"
)

///////////////////////////// file operation /////////////////////////////

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
	return common.Base58Decode(strings.TrimSpace(string(fd))), nil
}

// LoadKeyPair ...
func LoadKeyPair(privKeyFile string, algo string) (*account.KeyPair, error) {
	fsk, err := loadKey(privKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read key file failed: %v", err)
	}
	return account.NewKeyPair(fsk, GetSignAlgoByName(algo))
}

// SaveProtoStructToJSONFile ...
func SaveProtoStructToJSONFile(pb proto.Message, fileName string) error {
	r, err := (&jsonpb.Marshaler{EmitDefaults: true, Indent: "    "}).MarshalToString(pb)
	if err != nil {
		return err
	}
	return writeFile(fileName, []byte(r))
}

// LoadProtoStructFromJSONFile ...
func LoadProtoStructFromJSONFile(fileName string, pb proto.Message) error {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("load signature err %v %v", fileName, err)
	}
	err = jsonpb.UnmarshalString(string(bytes), pb)
	if err != nil {
		return fmt.Errorf("not a valid signature json %v %v", fileName, err)
	}
	return nil
}

// MarshalTextString ...
func MarshalTextString(pb proto.Message) string {
	r, err := (&jsonpb.Marshaler{EmitDefaults: true, Indent: "    "}).MarshalToString(pb)
	if err != nil {
		return "json.Marshal error: " + err.Error()
	}
	return r
}

////////////////////////////////// signature related /////////////////////////////////

func toRPCSign(sig *crypto.Signature) *rpcpb.Signature {
	return &rpcpb.Signature{
		Algorithm: rpcpb.Signature_Algorithm(sig.Algorithm),
		Signature: sig.Sig,
		PublicKey: sig.Pubkey,
	}
}

// GetSignatureOfTx ...
func GetSignatureOfTx(t *rpcpb.TransactionRequest, kp *account.KeyPair) *rpcpb.Signature {
	hash := common.Sha3(txToBytes(t, false))
	sig := toRPCSign(kp.Sign(hash))
	return sig
}

// GetSignAlgoByName ...
func GetSignAlgoByName(name string) crypto.Algorithm {
	switch name {
	case "secp256k1":
		return crypto.Secp256k1
	case "ed25519":
		return crypto.Ed25519
	default:
		return crypto.Ed25519
	}
}

// GetSignAlgoByEnum ...
func GetSignAlgoByEnum(enum rpcpb.Signature_Algorithm) crypto.Algorithm {
	switch enum {
	case rpcpb.Signature_SECP256K1:
		return crypto.Secp256k1
	case rpcpb.Signature_ED25519:
		return crypto.Ed25519
	default:
		return crypto.Ed25519
	}
}

// CheckPubKey check whether a string is a valid public key. Since it is not easy to check it fully, only check length here
func CheckPubKey(k string) bool {
	bytes := common.Base58Decode(k)
	if len(bytes) != 32 {
		return false
	}
	return true
}

// VerifySigForTx ...
func VerifySigForTx(t *rpcpb.TransactionRequest, sig *rpcpb.Signature) bool {
	hash := common.Sha3(txToBytes(t, false))
	return GetSignAlgoByEnum(sig.Algorithm).Verify(hash, sig.PublicKey, sig.Signature)
}

// NewAction ...
func NewAction(contract string, name string, data string) *rpcpb.Action {
	return &rpcpb.Action{
		Contract:   contract,
		ActionName: name,
		Data:       data,
	}
}

/////////////////////////////////// serialize deserialize ///////////////////////////////////////////

func actionToBytes(a *rpcpb.Action) []byte {
	se := common.NewSimpleEncoder()
	se.WriteString(a.Contract)
	se.WriteString(a.ActionName)
	se.WriteString(a.Data)
	return se.Bytes()
}

func amountToBytes(a *rpcpb.AmountLimit) []byte {
	se := common.NewSimpleEncoder()
	se.WriteString(a.Token)
	se.WriteString(a.Value)
	return se.Bytes()
}

func signatureToBytes(s *rpcpb.Signature) []byte {
	se := common.NewSimpleEncoder()
	se.WriteByte(byte(s.Algorithm))
	se.WriteBytes(s.Signature)
	se.WriteBytes(s.PublicKey)
	return se.Bytes()
}

func txToBytes(t *rpcpb.TransactionRequest, withSign bool) []byte {
	se := common.NewSimpleEncoder()
	se.WriteInt64(t.Time)
	se.WriteInt64(t.Expiration)
	se.WriteInt64(int64(t.GasRatio * 100))
	se.WriteInt64(int64(t.GasLimit * 100))
	se.WriteInt64(t.Delay)
	se.WriteInt32(int32(t.ChainId))
	se.WriteBytes(nil)
	se.WriteStringSlice(t.Signers)

	actionBytes := make([][]byte, 0, len(t.Actions))
	for _, a := range t.Actions {
		actionBytes = append(actionBytes, actionToBytes(a))
	}
	se.WriteBytesSlice(actionBytes)

	amountBytes := make([][]byte, 0, len(t.AmountLimit))
	for _, a := range t.AmountLimit {
		amountBytes = append(amountBytes, amountToBytes(a))
	}
	se.WriteBytesSlice(amountBytes)

	if withSign {
		signBytes := make([][]byte, 0, len(t.Signatures))
		for _, sig := range t.Signatures {
			signBytes = append(signBytes, signatureToBytes(sig))
		}
		se.WriteBytesSlice(signBytes)
	}

	return se.Bytes()
}
