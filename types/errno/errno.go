package errno

import (
	"github.com/crazyfrankie/gem/gerrors"
)

var (
	Success = gerrors.NewBizError(20000, "success")

	ErrInternalServer = gerrors.NewBizError(50000, "internal server error")
	ErrUnauthorized   = gerrors.NewBizError(40001, "unauthorized")
	ErrValidate       = gerrors.NewBizError(40002, "params in invalid")
	ErrForbidden      = gerrors.NewBizError(40003, "forbidden")
	ErrNotFound       = gerrors.NewBizError(40004, "resources not found")
)
