package v8

/*
#include "v8/vm.h"
*/
import "C"
import (
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
)

const cryptGasBase = 100

//export goSha3
func goSha3(cSbx C.SandboxPtr, msg C.CStr, gasUsed *C.size_t) C.CStr {
	msgStr := msg.GoString()
	val := common.Base58Encode(common.Sha3([]byte(msgStr)))

	*gasUsed = C.size_t(len(msgStr) + cryptGasBase)

	return newCStr(val)
}

//export goVerify
func goVerify(cSbx C.SandboxPtr, algo C.CStr, msg C.CStr, sig C.CStr, pubkey C.CStr, gasUsed *C.size_t) C.int {
	algoStr := algo.GoString()
	msgBytes := common.Base58Decode(msg.GoString())
	sigBytes := common.Base58Decode(sig.GoString())
	pubkeyBytes := common.Base58Decode(pubkey.GoString())
	if algoStr != "secp256k1" && algoStr != "ed25519" {
		return 0
	}
	*gasUsed = C.size_t(len(msgBytes) + cryptGasBase)
	if !crypto.NewAlgorithm(algoStr).Verify(msgBytes, pubkeyBytes, sigBytes) {
		return 0
	}
	return 1
}
