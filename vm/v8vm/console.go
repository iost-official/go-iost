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
	invalidLogLevelMap        = map[string]bool{
		"Fatal":   true,
		"Fatalln": true,
		"Fatalf":  true,
	}
)

//export goConsoleLog
func goConsoleLog(cSbx C.SandboxPtr, logLevel, logDetail C.CStr) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	levelStr := logLevel.GoString()
	detailStr := logDetail.GoString()

	if sbx.host.Logger() == nil {
		return C.CString(ErrConsoleNoLogger.Error())
	}

	loggerVal := reflect.ValueOf(sbx.host.Logger())
	loggerFunc := loggerVal.MethodByName(levelStr)

	if _, ok := invalidLogLevelMap[levelStr]; ok {
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
