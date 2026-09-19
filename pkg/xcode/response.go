package xcode

import (
	"net/http"
	"os"
	"github.com/pkg/errors"

	"rpc-agent/pkg/xcode/types"
)

func ErrHandler(err error) (int, any) {
	code := CodeFromError(err)

	if code.Code() == 500 {
		logStr := "INTERNAL_ERROR caused by: " + err.Error() + "\n"
		if errCause := errors.Cause(err); errCause != nil {
			logStr += "Cause: " + errCause.Error() + "\n"
		}
		
		f, _ := os.OpenFile("err_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			f.WriteString(logStr)
			f.Close()
		}
	}

	return http.StatusOK, types.Status{
		Code:    int32(code.Code()),
		Message: code.Message(),
	}
}
