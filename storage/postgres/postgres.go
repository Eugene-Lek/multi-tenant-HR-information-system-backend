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

func NewQueryWithFilter(baseQuery string, conditions []string) string {
	// TODO: change columnsToFilter to conditions and expect the caller to return the full condition.
	// They can do so by implementing their own counter

	for i := 0; i < len(conditions); i++ {
		if i == 0 {
			baseQuery = baseQuery + fmt.Sprintf(" WHERE %s", conditions[i])
		} else {
			baseQuery = baseQuery + fmt.Sprintf(" AND %s", conditions[i])
		}
	}

	return baseQuery
}

func NewUpdateQuery(table string, columnsToUpdate []string, columnsToFilterBy []string) string {
	query := fmt.Sprintf("UPDATE %s SET", table)

	for i := 0; i < len(columnsToUpdate); i++ {
		if i == 0 {
			query = query + fmt.Sprintf(" %s = $%v", columnsToUpdate[i], i+1)
		} else {
			query = query + fmt.Sprintf(", %s = $%v", columnsToUpdate[i], i+1)
		}
	}

	for i := 0; i < len(columnsToFilterBy); i++ {
		if i == 0 {
			query = query + fmt.Sprintf(" WHERE %s = $%v", columnsToFilterBy[i], i+1+len(columnsToUpdate))
		} else {
			query = query + fmt.Sprintf(" AND %s = $%v", columnsToFilterBy[i], i+1+len(columnsToUpdate))
		}
	}

	return query
}
