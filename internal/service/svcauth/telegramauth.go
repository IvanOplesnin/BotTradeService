package svcauth

import (
	"context"
	"time"

	"github.com/IvanOplesnin/BotTradeService.git/internal/domain/models"
)

func (a *AuthUsecase) CreateTelegramLinkCode(ctx context.Context, userID string, ttl time.Duration) (code string, expiresInSec int64, err error) {
	return "", 0, nil
}

func (a *AuthUsecase) LinkTelegram(ctx context.Context, code string, tg models.TelegramProfile) error {
	return nil
}

func (a *AuthUsecase) TelegramAuth(ctx context.Context, tg models.TelegramProfile) (models.AuthTokens, error) {
	return models.AuthTokens{}, nil
}
