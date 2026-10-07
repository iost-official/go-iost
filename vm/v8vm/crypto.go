package v8

/*
#include "v8/vm.h"
*/
import "C"
import (
	"encoding/hex"

	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/crypto"
)

const cryptGasBase = 100

//export goSha3
func goSha3(cSbx C.SandboxPtr, msg C.CStr, gasUsed *C.size_t) C.CStr {
	msgStr := GoString(msg)
	val := common.Base58Encode(common.Sha3([]byte(msgStr)))

	*gasUsed = C.size_t(len(msgStr) + cryptGasBase)

	return newCStr(val)
}

//export goSha3Hex
func goSha3Hex(cSbx C.SandboxPtr, msg C.CStr, gasUsed *C.size_t) C.CStr {
	msgBytes, err := hex.DecodeString(GoString(msg))
	*gasUsed = C.size_t(len(msgBytes) + cryptGasBase)
	if err != nil {
		return newCStr("")
	}
	val := hex.EncodeToString(common.Sha3(msgBytes))
	return newCStr(val)
}

//export goRipemd160Hex
func goRipemd160Hex(cSbx C.SandboxPtr, msg C.CStr, gasUsed *C.size_t) C.CStr {
	msgBytes, err := hex.DecodeString(GoString(msg))
	*gasUsed = C.size_t(len(msgBytes) + cryptGasBase)
	if err != nil {
		return newCStr("")
	}
	val := hex.EncodeToString(common.Ripemd160(msgBytes))
	return newCStr(val)
}

//export goVerify
func goVerify(cSbx C.SandboxPtr, algo C.CStr, msg C.CStr, sig C.CStr, pubkey C.CStr, gasUsed *C.size_t) C.int {
	algoStr := GoString(algo)
	msgBytes := common.Base58Decode(GoString(msg))
	sigBytes := common.Base58Decode(GoString(sig))
	pubkeyBytes := common.Base58Decode(GoString(pubkey))
	*gasUsed = C.size_t(len(msgBytes) + cryptGasBase)
	if algoStr != "secp256k1" && algoStr != "ed25519" {
		return 0
	}
	if !crypto.NewAlgorithm(algoStr).Verify(msgBytes, pubkeyBytes, sigBytes) {
		return 0
	}
	return 1
}
