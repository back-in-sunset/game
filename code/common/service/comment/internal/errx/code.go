package errx

import "errors"

const (
	CodeBadRequestDefault = "C1000"
	CodeObjIDRequired     = "C1001"
	CodeObjTypeRequired   = "C1002"
	CodeMemberIDRequired  = "C1003"
	CodeCommentIDRequired = "C1004"
	CodeMessageRequired   = "C1005"
	CodeMessageTooLong    = "C1006"
	CodeSortTypeInvalid   = "C1007"
	CodePermissionDenied  = "C2001"
	CodeCommentNotFound   = "C3001"
	CodeInvalidReply      = "C3002"
	CodeInternalDefault   = "C5000"
)

// ErrorSortTypeInvalid is returned when the sort type is invalid.
var ErrorSortTypeInvalid = errors.New("sort type is invalid")
