package psql

import (
	"context"
	"errors"

	modelerrors "github.com/IvanOplesnin/BotTradeService.git/internal/domain/errors"
	"github.com/IvanOplesnin/BotTradeService.git/internal/domain/models"
	"github.com/IvanOplesnin/BotTradeService.git/internal/repository/psql/query"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) CreateUser(ctx context.Context, user models.User) (int32, error) {
	createUserParams := query.CreateUserParams{
		Email:        user.Email,
		HashPassword: user.HashPassword,
	}

	userId, err := r.queries.CreateUser(ctx, createUserParams)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return 0, modelerrors.ErrEmailTaken
	}
	if err != nil {
		return 0, err
	}
	return userId, err
}

func (r *Repo) GetByEmail(ctx context.Context, email string) (models.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, modelerrors.ErrNoRows
		}
		return models.User{}, err
	}
	return models.User{
		ID:           user.ID,
		HashPassword: user.HashPassword,
	}, nil
}
