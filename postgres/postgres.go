package postgres

import (
	"database/sql"
	_ "github.com/lib/pq" // Import pq for its side effects (driver install)
)


type postgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage (connStr string) (*postgresStorage, error) {
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &postgresStorage{
		db: db,
	}, nil 
}

