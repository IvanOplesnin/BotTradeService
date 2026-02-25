package token

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/IvanOplesnin/BotTradeService.git/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Tokener struct {
	secret    []byte
	ttl       time.Duration
	issuer    string
	clockSkew time.Duration
}

func NewTokener(cfg config.Tokener) (*Tokener, error) {
	if len(cfg.Secret) < 16 {
		return nil, fmt.Errorf("jwt secret too short")
	}
	if cfg.TTL <= 0 {
		cfg.TTL = config.SecondsDuration(time.Hour)
	}
	if cfg.Issuer == "" {
		cfg.Issuer = "auth-service"
	}
	if cfg.ClockSkew < 0 {
		cfg.ClockSkew = 0
	}
	return &Tokener{
		secret:    cfg.Secret,
		ttl:       time.Duration(cfg.TTL),
		issuer:    cfg.Issuer,
		clockSkew: time.Duration(cfg.ClockSkew),
	}, nil
}

// JwtClaims — свои claims + стандартные registered claims
type JwtClaims struct {
	UserID int32 `json:"user_id"`
	jwt.RegisteredClaims
}

// CreateToken — соответствует твоему интерфейсу Tokener.CreateToken(...)
func (t *Tokener) Token(userID int32) (accessToken string, expInSec int64, err error) {
	now := time.Now()

	claims := JwtClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    t.issuer,
			Subject:   strconv.FormatInt(int64(userID), 10), // стандартный sub
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-t.clockSkew)),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(t.secret)
	if err != nil {
		return "", 0, err
	}

	return s, int64(t.ttl.Seconds()), nil
}
