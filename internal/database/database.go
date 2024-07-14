package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DatabaseSQL struct {
	sqlDB *sql.DB
}

func NewDatabaseSQL(databaseDSN string) *DatabaseSQL {
	fmt.Println(databaseDSN)
	sqlDB, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		panic(err)
	}
	return &DatabaseSQL{
		sqlDB: sqlDB,
	}
}

func (db *DatabaseSQL) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return db.sqlDB.PingContext(ctx)
}

func (db *DatabaseSQL) Close() error {
	return db.sqlDB.Close()
}
