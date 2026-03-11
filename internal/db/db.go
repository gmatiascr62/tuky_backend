package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"tukychat/internal/config"
)

var (
	once sync.Once
	pool *pgxpool.Pool
	err  error
)

func GetPool() (*pgxpool.Pool, error) {
	once.Do(func() {
		cfg := config.Load()
		if cfg.DatabaseURL == "" {
			err = fmt.Errorf("DATABASE_URL no configurada")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		poolConfig, parseErr := pgxpool.ParseConfig(cfg.DatabaseURL)
		if parseErr != nil {
			err = fmt.Errorf("parse DATABASE_URL: %w", parseErr)
			return
		}

		poolConfig.MaxConns = 5
		poolConfig.MinConns = 0

		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			err = fmt.Errorf("crear pool pgx: %w", err)
			return
		}

		if pingErr := pool.Ping(ctx); pingErr != nil {
			err = fmt.Errorf("ping db: %w", pingErr)
			return
		}
	})

	return pool, err
}