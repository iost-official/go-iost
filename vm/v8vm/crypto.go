package v8

/*
#include "v8/vm.h"
*/
import "C"
import "github.com/iost-official/go-iost/common"

const cryptGasBase = 100

//export goSha3
func goSha3(cSbx C.SandboxPtr, msg *C.char, gasUsed *C.size_t) *C.char {
	msgStr := C.GoString(msg)
	val := common.Base58Encode(common.Sha3([]byte(msgStr)))

	*gasUsed = C.size_t(len(msgStr) + cryptGasBase)

	return C.CString(val)
}
