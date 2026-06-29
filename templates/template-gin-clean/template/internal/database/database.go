package database

import (
	"database/sql"
	"fmt"

	"{{MODULE_NAME}}/internal/config"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Connect opens a database connection using the config.
func Connect(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}
