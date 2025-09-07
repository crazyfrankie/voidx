package token

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/voidx/infra/contract/token"
)

type TokenService struct {
	cmd       redis.Cmdable
	signAlgo  string
	secretKey []byte
}

func NewTokenService(cmd redis.Cmdable, signAlgo string, secret string) token.Token {
	return &TokenService{cmd: cmd, signAlgo: signAlgo, secretKey: []byte(secret)}
}

func (s *TokenService) GenerateToken(uid uuid.UUID, ua string) ([]string, error) {
	res := make([]string, 2)
	access, err := s.newToken(uid, time.Hour)
	if err != nil {
		return res, err
	}
	res[0] = access
	refresh, err := s.newToken(uid, time.Hour*24*30)
	if err != nil {
		return res, err
	}
	res[1] = refresh

	// set refresh in redis
	key := tokenKey(uid, ua)

	err = s.cmd.Set(context.Background(), key, refresh, time.Hour*24*30).Err()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *TokenService) newToken(uid uuid.UUID, duration time.Duration) (string, error) {
	now := time.Now()
	claims := &token.Claims{
		UID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	}
	tk := jwt.NewWithClaims(jwt.GetSigningMethod(s.signAlgo), claims)
	str, err := tk.SignedString(s.secretKey)

	return str, err
}

func (s *TokenService) ParseToken(tk string) (*token.Claims, error) {
	t, err := jwt.ParseWithClaims(tk, &token.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := t.Claims.(*token.Claims)
	if ok {
		return claims, nil
	}

	return nil, errors.New("jwt is invalid")
}

func (s *TokenService) TryRefresh(refresh string, ua string) ([]string, uuid.UUID, error) {
	refreshClaims, err := s.ParseToken(refresh)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("invalid refresh jwt")
	}

	res, err := s.cmd.Get(context.Background(), tokenKey(refreshClaims.UID, ua)).Result()
	if err != nil || res != refresh {
		return nil, uuid.Nil, errors.New("jwt invalid or revoked")
	}

	access, err := s.newToken(refreshClaims.UID, time.Hour)
	if err != nil {
		return nil, uuid.Nil, err
	}

	now := time.Now()
	issat, _ := refreshClaims.GetIssuedAt()
	expire, _ := refreshClaims.GetExpirationTime()
	if expire.Sub(now) < expire.Sub(issat.Time)/3 {
		// try refresh
		refresh, err = s.newToken(refreshClaims.UID, time.Hour*24*30)
		err = s.cmd.Set(context.Background(), tokenKey(refreshClaims.UID, ua), refresh, time.Hour*24*30).Err()
		if err != nil {
			return nil, uuid.Nil, err
		}
	}

	return []string{access, refresh}, refreshClaims.UID, nil
}

func (s *TokenService) CleanToken(ctx context.Context, uid uuid.UUID, ua string) error {
	return s.cmd.Del(ctx, tokenKey(uid, ua)).Err()
}

func tokenKey(uid uuid.UUID, ua string) string {
	hash := hashUA(ua)
	return fmt.Sprintf("refresh_token:%s:%s", uid, hash)
}

func hashUA(ua string) string {
	sum := sha1.Sum([]byte(ua))
	return hex.EncodeToString(sum[:])
}
