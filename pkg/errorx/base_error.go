package errorx

import (
	"fmt"
	constantpublic "rpc-user/pkg/constant/common/public"
)

type CodeError struct {
	Result string `json:"result"`
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
}

type CodeErrorResponse struct {
	Result string      `json:"result"`
	Code   int         `json:"code"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

func NewCodeError(code Code) error {
	return &CodeError{Result: constantpublic.RespResultFailure, Code: code.Value(), Msg: code.Msg()}
}

func NewDefaultError() error {
	return NewCodeError(Unknown)
}

func NewCodeMsg(code int, msg string) error {
	return &CodeError{Result: constantpublic.RespResultFailure, Code: code, Msg: msg}
}

func NewDefaultMsg(msg string) error {
	return NewCodeMsg(int(IllegalParam), msg)
}

func (e *CodeError) Error() string {
	return e.Msg
}

func (e *CodeError) Data() *CodeErrorResponse {
	code := e.Code
	if code == OK.Value() {
		code = Unknown.Value()
	}

	return &CodeErrorResponse{
		Result: constantpublic.RespResultFailure,
		Code:   code,
		Msg:    e.Msg,
		Data:   nil,
	}
}

func NewMysqlError(err interface{}) error {
	return fmt.Errorf("mysql db err = %v", err)
}

func NewInvokeRpcError(err interface{}) error {
	return fmt.Errorf("invoke rpc err = %v", err)
}

func NewInvokeLogicError(err interface{}) error {
	return fmt.Errorf("invoke logic err = %v", err)
}
