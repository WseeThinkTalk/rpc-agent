package errorx

type Code int32

const (
	OK           Code = 200
	IllegalParam Code = 400
	UnAuthorize  Code = 401
	NotFound     Code = 404
	Unknown      Code = 500

	UserNotExists    Code = 10001
	PasswordError    Code = 10002
	UserAlreadyExist Code = 10003
	MobileInvalid    Code = 10004
	SmsCodeInvalid   Code = 10005
)

var codeMsgMap = map[Code]string{
	OK:               "success",
	IllegalParam:     "非法参数",
	UnAuthorize:      "未授权或Token已失效",
	NotFound:         "资源未找到",
	Unknown:          "系统内部错误",
	UserNotExists:    "用户不存在",
	PasswordError:    "密码错误",
	UserAlreadyExist: "用户已存在",
	MobileInvalid:    "手机号格式不正确",
	SmsCodeInvalid:   "验证码错误或已失效",
}

func (c Code) Value() int {
	return int(c)
}

func (c Code) Msg() string {
	if msg, ok := codeMsgMap[c]; ok {
		return msg
	}
	return "未知错误"
}
