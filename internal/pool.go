package internal

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

func NewPool(i do.Injector) (*pgxpool.Pool, error) {
	cfg := do.MustInvoke[*Config](i)
	return pgxpool.New(context.Background(), cfg.DatabaseURL)
}
