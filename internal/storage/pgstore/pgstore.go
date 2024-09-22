package pgstore

import (
	"context"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/pressly/goose/v3"
)

type PostgresqlStore struct {
	*sqlx.DB
}

func New(ctx context.Context, dataSourceName string) (*PostgresqlStore, error) {
	db, err := sqlx.ConnectContext(ctx, "pgx", dataSourceName)
	if err != nil {
		return nil, err
	}
	store := &PostgresqlStore{db}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = store.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	err = store.Migrate(ctx)
	if err != nil {
		return nil, err
	}
	return store, nil
}

func (store *PostgresqlStore) Migrate(ctx context.Context) error {
	migrationsDir := "./migrations"
	if err := goose.ResetContext(ctx, store.DB.DB, migrationsDir); err != nil {
		return err
	}
	if err := goose.UpContext(ctx, store.DB.DB, migrationsDir); err != nil {
		return err
	}
	return nil
}
