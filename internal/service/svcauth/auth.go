package svcauth

import (
	"context"
	"errors"

	"github.com/IvanOplesnin/BotTradeService.git/internal/app/models"
	modelerrors "github.com/IvanOplesnin/BotTradeService.git/internal/domain/errors"
)

type Hasher interface {
	Hash(password string) (hash string, err error)
	ComapareHash(password, hash string) (match bool, err error)
}

type Tokener interface {
	CreateToken(userID int32) (accessToken string, exp int64, err error)
}

type AuthRepo interface {
	CreateUser(ctx context.Context, user models.User) (int32, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
}

type AuthUsecase struct {
	hasher  Hasher
	tokener Tokener
	repo    AuthRepo
}

type AuthUsecaseDeps struct {
	Hasher  Hasher
	Tokener Tokener
	Repo    AuthRepo
}

func New(deps AuthUsecaseDeps) *AuthUsecase {
	return &AuthUsecase{
		hasher:  deps.Hasher,
		tokener: deps.Tokener,
		repo:    deps.Repo,
	}
}

func (a *AuthUsecase) Register(ctx context.Context, email, password string) (models.AuthTokens, error) {
	hash, err := a.hasher.Hash(password)
	if err != nil {
		return models.AuthTokens{}, err
	}
	user := models.User{
		Email:        email,
		HashPassword: hash,
	}
	userID, err := a.repo.CreateUser(ctx, user)
	if errors.Is(err, modelerrors.ErrNoRows) {
		return models.AuthTokens{}, modelerrors.ErrEmailTaken
	}
	if err != nil {
		return models.AuthTokens{}, err
	}
	accessToken, expInSec, err := a.tokener.CreateToken(userID)
	if err != nil {
		return models.AuthTokens{}, err
	}
	return models.AuthTokens{
		AccessToken:  accessToken,
		ExpiresInSec: expInSec,
	}, nil
}

func (a *AuthUsecase) Login(ctx context.Context, email, password string) (models.AuthTokens, error) {
	u, err := a.repo.GetByEmail(ctx, email)
	if errors.Is(err, modelerrors.ErrNoRows) {
		return models.AuthTokens{}, modelerrors.ErrInvalidCredentials
	}
	if err != nil {
		return models.AuthTokens{}, err
	}
	if ok, err := a.hasher.ComapareHash(password, u.HashPassword); !ok {
		if err != nil {
			return models.AuthTokens{}, err
		}
		return models.AuthTokens{}, modelerrors.ErrInvalidCredentials
	}
	accessToken, expInSec, err := a.tokener.CreateToken(u.ID)

	return models.AuthTokens{
		AccessToken:  accessToken,
		ExpiresInSec: expInSec,
	}, nil
}
