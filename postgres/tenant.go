package postgres

import (
	"github.com/lib/pq"

	"multi-tenant-HR-information-system-backend/routes"
)

func (postgres *postgresStorage) CreateTenant(tenant routes.Tenant) error {
	query := "INSERT INTO tenant (name) VALUES ($1)"
	_, err := postgres.db.Exec(query, tenant.Name)

	if pgErr, ok := err.(*pq.Error); ok {
		// 23505 corresponds to the Unique Violation error
		if pgErr.Code == "23505" {
			return NewUniqueViolationError("tenant", pgErr)
		} else {
			return routes.NewInternalServerError(pgErr)
		}
	} else if err != nil {
		return routes.NewInternalServerError(err)
	}

	return nil
}

func (postgres *postgresStorage) CreateDivision(division routes.Division) error {
	query := "INSERT INTO division (tenant, name) VALUES ($1, $2)"
	_, err := postgres.db.Exec(query, division.Tenant, division.Name)
	if pgErr, ok := err.(*pq.Error); ok {
		// 23505 corresponds to the
		switch pgErr.Code {
		case "23505":
			//Unique Violation error
			return NewUniqueViolationError("division", pgErr)
		case "23503":
			//Foreign Key Violation error
			return NewInvalidForeignKeyError(pgErr)
		default:
			return routes.NewInternalServerError(pgErr)
		}
	} else if err != nil {
		return routes.NewInternalServerError(err)
	}

	return nil
}

func (postgres *postgresStorage) CreateDepartment(department routes.Department) error {
	query := "INSERT INTO department (tenant, division, name) VALUES ($1, $2, $3)"
	_, err := postgres.db.Exec(query, department.Tenant, department.Division, department.Name)
	if pgErr, ok := err.(*pq.Error); ok {
		//TODO fill in error handling
		switch pgErr.Code {
		case "23505":
			//Unique Violation error
			return NewUniqueViolationError("department", pgErr)
		case "23503":
			//Foreign Key Violation error
			return NewInvalidForeignKeyError(pgErr)
		default:
			return routes.NewInternalServerError(pgErr)
		}
	} else if err != nil {
		return routes.NewInternalServerError(err)
	}

	return nil

}
