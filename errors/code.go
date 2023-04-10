package errors

var (
	ErrUnknown      = NewError("unknow", "未知错误")
	ErrSucceeded    = NewError("succeeded", "成功")
	ErrBodyParser   = NewError("err_body_parser", "请求体解析错误")
	ErrQueryParser  = NewError("err_query_parser", "查询参数解析错误")
	ErrParameter    = NewError("err_parameter", "参数错误")
	ErrDatabase     = NewError("err_database", "数据库操作错误")
	ErrUnauthorized = NewError("err_unauthorized", "未授权")
	ErrForbidden    = NewError("err_forbidden", "无权限")
)
