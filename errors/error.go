package errors

type ErrorMessage struct {
	Code    string
	Message any
}

func NewError(code string, message any) *ErrorMessage {
	return &ErrorMessage{Code: code, Message: message}
}

func Message(key string, value string) map[string]string {
	return map[string]string{key: value}
}
