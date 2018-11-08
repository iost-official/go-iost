package v8

/*
#include "v8/vm.h"
*/
import "C"
import (
	"reflect"
	"errors"
)

var (
	ErrConsoleNoLogger = errors.New("no logger error.")
	ErrConsoleInvalidLogLevel = errors.New("log invalid level.")
)

//export goConsoleLog
func goConsoleLog(cSbx C.SandboxPtr, logLevel, logDetail *C.char) int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	levelStr := C.GoString(logLevel)
	detailStr := C.GoString(logDetail)

	if sbx.host.Logger() == nil {
		return C.Cstring(ErrConsoleNoLogger.Error())
	}

	loggerVal := reflect.ValueOf(sbx.host.Logger())
	loggerFunc := loggerVal.MethodByName(levelStr)

	if !loggerFunc.IsValid() {
		return C.Cstring(ErrConsoleInvalidLogLevel.Error())
	}

	loggerFunc.Call([]reflect.Value{
		reflect.ValueOf(detailStr),
	})

	return C.Cstring(MessageSuccess)
}
