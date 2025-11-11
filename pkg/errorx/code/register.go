package code

import "github.com/crazyfrankie/voidx/pkg/errorx/internal"

type RegisterOptionFn = internal.RegisterOption

// WithAffectStability sets the stability flag,
// true: it will affect the system stability and is reflected in the interface error rate,
// false: it will not affect the stability.
func WithAffectStability(affectStability bool) RegisterOptionFn {
	return internal.WithAffectStability(affectStability)
}

// Register the predefined error code information of the registered user,
// and call the code_gen sub-module corresponding to the PSM service when initializing.
func Register(code int32, msg string, opts ...RegisterOptionFn) {
	internal.Register(code, msg, opts...)
}

// SetDefaultErrorCode Code with PSM information staining Replace the default code.
func SetDefaultErrorCode(code int32) {
	internal.SetDefaultErrorCode(code)
}
