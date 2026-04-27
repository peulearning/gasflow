package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

// Config agrupa as configurações de conexão com o MySQL.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// DSN retorna a string de conexão no formato go-sql-driver/mysql.
func (c Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		c.User, c.Password, c.Host, c.Port, c.Name,
	)
}

// Connect abre e valida a conexão com o MySQL, configurando o pool.
func Connect(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}

	// Pool de conexões recomendado para produção.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	log.Info().Str("host", cfg.Host).Str("db", cfg.Name).Msg("mysql: connected")
	return db, nil
}

// MustConnect conecta ou encerra o processo (usar apenas no startup).
func MustConnect(cfg Config) *sql.DB {
	db, err := Connect(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("mysql: failed to connect")
	}
	return db
}