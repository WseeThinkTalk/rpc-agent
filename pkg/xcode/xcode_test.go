package xcode

import (
	"testing"
)

func TestNew(t *testing.T) {
	code := New(10001, "参数错误")
	if code.Code() != 10001 {
		t.Errorf("Code() = %d, want 10001", code.Code())
	}
	if code.Message() != "参数错误" {
		t.Errorf("Message() = %q, want %q", code.Message(), "参数错误")
	}
	if code.Error() != "参数错误" {
		t.Errorf("Error() = %q, want %q", code.Error(), "参数错误")
	}
}

func TestCodeErrorEmptyMsg(t *testing.T) {
	code := New(404, "")
	if code.Error() != "404" {
		t.Errorf("Error() with empty msg = %q, want %q", code.Error(), "404")
	}
}

func TestCodeDetails(t *testing.T) {
	code := New(500, "internal error")
	if code.Details() != nil {
		t.Error("Details() should return nil")
	}
}

func TestString(t *testing.T) {
	code := String("0")
	if code.Code() != OK.Code() {
		t.Errorf("String('0') should be OK, got code %d", code.Code())
	}
}

func TestStringEmpty(t *testing.T) {
	code := String("")
	if code.Code() != OK.Code() {
		t.Error("String('') should return OK")
	}
}

func TestStringInvalid(t *testing.T) {
	code := String("abc")
	if code.Code() != ServerErr.Code() {
		t.Errorf("String('abc') should return ServerErr, got code %d", code.Code())
	}
}

func TestStringNumeric(t *testing.T) {
	code := String("403")
	if code.Code() != 403 {
		t.Errorf("String('403') code = %d, want 403", code.Code())
	}
}
