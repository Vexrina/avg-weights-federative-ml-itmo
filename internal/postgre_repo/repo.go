package postgre_repo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgreDB struct {
	db *pgxpool.Pool
}

var (
	once sync.Once
	db   postgreDB
)

func NewPgDb(ctx context.Context, dsn string) *postgreDB {
	once.Do(func() {
		cfg, cfgErr := pgxpool.ParseConfig(dsn)
		if cfgErr != nil {
			panic("cant get pg db, err is: " + cfgErr.Error())
		}
		cfg.MaxConns = 10
		cfg.MinConns = 2
		cfg.MaxConnLifetime = time.Hour
		cfg.MaxConnIdleTime = 10 * time.Minute

		pool, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			panic(fmt.Errorf("failed to create pgxpool: %w", err))
		}
		db = postgreDB{db: pool}
	})

	return &db
}
