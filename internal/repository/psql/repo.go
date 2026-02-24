package psql

import (
	"github.com/IvanOplesnin/BotTradeService.git/internal/repository/psql/query"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pgUniqueViolation = "23505"

type Repo struct {
	db      *pgxpool.Pool
	queries *query.Queries
}

func NewPsqlRepo(db *pgxpool.Pool) *Repo {
	return &Repo{
		db:      db,
		queries: query.New(db),
	}
}
