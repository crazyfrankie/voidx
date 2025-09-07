package token

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UID uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

type Token interface {
	GenerateToken(uid uuid.UUID, ua string) ([]string, error)
	ParseToken(token string) (*Claims, error)
	TryRefresh(refresh string, ua string) ([]string, uuid.UUID, error)
	CleanToken(ctx context.Context, uid uuid.UUID, ua string) error
}
