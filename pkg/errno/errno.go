package errno

import "github.com/crazyfrankie/gem/gerrors"

type ErrCode gerrors.BizErrorIface

func NewErrCode(code int32, msg string) ErrCode {
	return gerrors.NewBizError(code, msg)
}

var (
	Success ErrCode = gerrors.NewBizError(20000, "success")

	ErrInternalServer ErrCode = gerrors.NewBizError(50000, "internal server error")
	ErrUnauthorized   ErrCode = gerrors.NewBizError(40001, "unauthorized")
	ErrValidate       ErrCode = gerrors.NewBizError(40002, "params in invalid")
	ErrForbidden      ErrCode = gerrors.NewBizError(40003, "forbidden")
	ErrNotFound       ErrCode = gerrors.NewBizError(40004, "resources not found")
)
