package postgres

import (
	"github.com/lib/pq"

	"multi-tenant-HR-information-system-backend/errors"
	"multi-tenant-HR-information-system-backend/routes"
)

func (postgres *postgresStorage) CreateTenant(tenant routes.Tenant) errors.HttpError {
	query := "INSERT INTO tenant (name) VALUES ($1)"
	_, err := postgres.db.Exec(query, tenant.Name)

	if pgErr, ok := err.(*pq.Error); ok {
		// 23505 corresponds to the Unique Violation error
		if pgErr.Code == "23505" {
			return errors.NewUniqueViolationError("tenant", [][2]string{{"name", tenant.Name}})	
		} else {
			return errors.NewInternalServerError(pgErr.Error())
		}
	} else if err != nil {
		return errors.NewInternalServerError(err.Error())
	}

	return nil
}

func (postgres *postgresStorage) CreateDivision(division routes.Division) errors.HttpError {
	query := "INSERT INTO division (tenant, name) VALUES ($1, $2)"
	_, err := postgres.db.Exec(query, division.Tenant, division.Name)
	if pgErr, ok := err.(*pq.Error); ok {
		// 23505 corresponds to the
		switch pgErr.Code {
		case "23505":
			//Unique Violation error
			return errors.NewUniqueViolationError("division", [][2]string{{"tenant", division.Tenant}, {"name", division.Name}})
		case "23503":
			//Foreign Key Violation error
			return errors.NewInvalidForeignKeyError([][2]string{{"tenant", division.Tenant}})
		default:
			return errors.NewInternalServerError(pgErr.Error())
		}
	} else if err != nil {
		return errors.NewInternalServerError(err.Error())
	}

	return nil
}

func (postgres *postgresStorage) CreateDepartment(department routes.Department) errors.HttpError {
	query := "INSERT INTO department (tenant, division, name) VALUES ($1, $2, $3)"
	_, err := postgres.db.Exec(query, department.Tenant, department.Division, department.Name)
	if pgErr, ok := err.(*pq.Error); ok {
		//TODO fill in error handling
		switch pgErr.Code {
		case "23505":
			//Unique Violation error
			return errors.NewUniqueViolationError("department", [][2]string{{"tenant", department.Tenant}, {"division", department.Division}, {"name", department.Name}})
		case "23503":
			//Foreign Key Violation error
			return errors.NewInvalidForeignKeyError([][2]string{{"tenant", department.Tenant}, {"division", department.Division}})
		default:
			return errors.NewInternalServerError(pgErr.Error())
		}
	} else if err != nil {
		return errors.NewInternalServerError(err.Error())
	}

	return nil

}
