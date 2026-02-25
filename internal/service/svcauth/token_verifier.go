package svcauth

import (
	"context"

	"github.com/IvanOplesnin/BotTradeService.git/internal/domain/models"
)

func (a *AuthUsecase) ValidateAccessToken(ctx context.Context, accessToken string) (userID string, err error) {
	return "", nil
}

func (a *AuthUsecase) ValidateBotSignature(ctx context.Context, meta models.BotMeta, fullMethod string, reqBytes []byte) error {
	return nil
}
