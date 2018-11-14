package v8

/*
#include "v8/vm.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"reflect"
)

// Error message
var (
	ErrConsoleNoLogger        = errors.New("no logger error")
	ErrConsoleInvalidLogLevel = errors.New("log invalid level")
)

//export goConsoleLog
func goConsoleLog(cSbx C.SandboxPtr, logLevel, logDetail *C.char) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	levelStr := C.GoString(logLevel)
	detailStr := C.GoString(logDetail)
	fmt.Printf("console.log %v %v\n", levelStr, detailStr)

	if sbx.host.Logger() == nil {
		return C.CString(ErrConsoleNoLogger.Error())
	}

	loggerVal := reflect.ValueOf(sbx.host.Logger())
	loggerFunc := loggerVal.MethodByName(levelStr)

	if !loggerFunc.IsValid() {
		return C.CString(ErrConsoleInvalidLogLevel.Error())
	}

	loggerFunc.Call([]reflect.Value{
		reflect.ValueOf(detailStr),
	})

	return nil
}
