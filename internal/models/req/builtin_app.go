package req

import "github.com/google/uuid"

type AddBuiltinAppReq struct {
	BuiltinAppID uuid.UUID `json:"builtin_app_id"`
}
