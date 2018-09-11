package v8

/*
#include "v8/vm.h"
*/
import "C"
import (
	"reflect"
)

// const Console log status
const (
	ConsoleLogSuccess = iota
	ConsoleLogUnexpectedError
	ConsoleLogNoLoggerError
	ConsoleLogInvalidLogLevelError
)

//export goConsoleLog
func goConsoleLog(cSbx C.SandboxPtr, logLevel, logDetail *C.char) int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return ContractCallUnexpectedError
	}

	levelStr := C.GoString(logLevel)
	detailStr := C.GoString(logDetail)

	if sbx.host.Logger() == nil {
		return ConsoleLogNoLoggerError
	}

	loggerVal := reflect.ValueOf(sbx.host.Logger())
	loggerFunc := loggerVal.MethodByName(levelStr)

	if !loggerFunc.IsValid() {
		return ConsoleLogInvalidLogLevelError
	}

	loggerFunc.Call([]reflect.Value{
		reflect.ValueOf(detailStr),
	})

	return ConsoleLogSuccess
}
