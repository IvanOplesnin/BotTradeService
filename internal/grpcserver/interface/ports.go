package grpcports

import (
	"context"
	"time"

	"github.com/IvanOplesnin/BotTradeService.git/internal/domain/models"
)

type AuthUsecase interface {
	// Web
	Register(ctx context.Context, email, password string) (models.AuthTokens, error)
	Login(ctx context.Context, email, password string) (models.AuthTokens, error)

	// Web: код для привязки Telegram (JWT required, userID берём из ctx)
	CreateTelegramLinkCode(ctx context.Context, userID string, ttl time.Duration) (code string, expiresInSec int64, err error)

	// Telegram: привязка Telegram к web-аккаунту по коду (bot-signature required)
	LinkTelegram(ctx context.Context, code string, tg models.TelegramProfile) error

	// Telegram: логин после привязки (bot-signature required)
	TelegramAuth(ctx context.Context, tg models.TelegramProfile) (models.AuthTokens, error)
}

type TokenVerifier interface {
	ValidateAccessToken(ctx context.Context, accessToken string) (userID string, err error)
}

type BotVerifier interface {
	ValidateBotSignature(ctx context.Context, meta models.BotMeta, fullMethod string, reqBytes []byte) error
}
