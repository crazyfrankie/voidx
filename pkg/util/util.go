package util

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/pkg/errno"
)

func GetCurrentUserID(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, errno.ErrUnauthorized.AppendBizMessage("未登录")
	}

	return userID, nil
}
