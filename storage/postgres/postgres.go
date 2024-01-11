package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // Import pq for its side effects (driver install)
)

type postgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connStr string) (*postgresStorage, error) {
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

func NewDynamicConditionQuery(baseQuery string, conditions []string) string {
	for i := 0; i < len(conditions); i++ {
		if i == 0 {
			baseQuery = baseQuery + fmt.Sprintf(" WHERE %s = $%v", conditions[i], i+1)
		} else {
			baseQuery = baseQuery + fmt.Sprintf(" AND %s = $%v", conditions[i], i+1)
		}
	}

	return baseQuery
}
