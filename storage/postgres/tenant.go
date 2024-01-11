package postgres

import (
	"github.com/lib/pq"

	"multi-tenant-HR-information-system-backend/httperror"	
	"multi-tenant-HR-information-system-backend/storage"		
)

func (postgres *postgresStorage) CreateTenant(tenant storage.Tenant) error {
	query := "INSERT INTO tenant (id, name) VALUES ($1, $2)"
	_, err := postgres.db.Exec(query, tenant.Id, tenant.Name)

	if pgErr, ok := err.(*pq.Error); ok {
		// 23505 corresponds to the Unique Violation error
		if pgErr.Code == "23505" {
			return NewUniqueViolationError("tenant", pgErr)
		} else {
			return httperror.NewInternalServerError(pgErr)
		}
	} else if err != nil {
		return httperror.NewInternalServerError(err)
	}

	return nil
}

func (postgres *postgresStorage) CreateDivision(division storage.Division) error {
	query := "INSERT INTO division (id, tenant_id, name) VALUES ($1, $2, $3)"
	_, err := postgres.db.Exec(query, division.Id, division.TenantId, division.Name)
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
			return httperror.NewInternalServerError(pgErr)
		}
	} else if err != nil {
		return httperror.NewInternalServerError(err)
	}

	return nil
}

func (postgres *postgresStorage) CreateDepartment(department storage.Department) error {
	query := "INSERT INTO department (id, tenant_id, division_id, name) VALUES ($1, $2, $3, $4)"
	_, err := postgres.db.Exec(query, department.Id, department.TenantId, department.DivisionId, department.Name)
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
			return httperror.NewInternalServerError(pgErr)
		}
	} else if err != nil {
		return httperror.NewInternalServerError(err)
	}

	return nil

}
