package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
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

		fmt.Println("DATABASE_URL presente:", cfg.DatabaseURL != "")
		fmt.Println("DATABASE_URL:", cfg.DatabaseURL)

		if cfg.DatabaseURL == "" {
			err = fmt.Errorf("DATABASE_URL no configurada")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		fmt.Println("Parseando DATABASE_URL...")

		poolConfig, parseErr := pgxpool.ParseConfig(cfg.DatabaseURL)
		if parseErr != nil {
			fmt.Println("ERROR parse DATABASE_URL:", parseErr)
			err = fmt.Errorf("parse DATABASE_URL: %w", parseErr)
			return
		}

		poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

		poolConfig.MaxConns = 5
		poolConfig.MinConns = 0

		fmt.Println("Creando pool...")

		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			fmt.Println("ERROR crear pool:", err)
			err = fmt.Errorf("crear pool pgx: %w", err)
			return
		}

		fmt.Println("Haciendo ping a la DB...")

		if pingErr := pool.Ping(ctx); pingErr != nil {
			fmt.Println("ERROR ping DB:", pingErr)
			err = fmt.Errorf("ping db: %w", pingErr)
			return
		}

		fmt.Println("Conexión a DB exitosa")
	})

	return pool, err
}
