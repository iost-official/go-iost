package v8

/*
#include "v8/vm.h"
*/
import "C"
import (
	"errors"
	"reflect"
)

// Error message
var (
	ErrConsoleNoLogger        = errors.New("no logger error")
	ErrConsoleInvalidLogLevel = errors.New("log invalid level")
	validLogLevelMap          = map[string]bool{
		"Debug": true,
		"Info":  true,
		"Warn":  true,
		"Error": true,
	}
)

//export goConsoleLog
func goConsoleLog(cSbx C.SandboxPtr, logLevel, logDetail C.CStr) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	levelStr := GoString(logLevel)
	detailStr := GoString(logDetail)

	if sbx.host.Logger() == nil {
		return C.CString(ErrConsoleNoLogger.Error())
	}

	loggerVal := reflect.ValueOf(sbx.host.Logger())
	loggerFunc := loggerVal.MethodByName(levelStr)

	if _, ok := validLogLevelMap[levelStr]; !ok {
		return C.CString(ErrConsoleInvalidLogLevel.Error())
	}

	if !loggerFunc.IsValid() {
		return C.CString(ErrConsoleInvalidLogLevel.Error())
	}

	loggerFunc.Call([]reflect.Value{
		reflect.ValueOf(detailStr),
	})

	return nil
}
