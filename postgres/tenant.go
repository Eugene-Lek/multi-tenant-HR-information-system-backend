package postgres

import (
	"fmt"

	"github.com/lib/pq"

	"multi-tenant-HR-information-system-backend/errors"
)

func (postgres *postgresStorage) CreateTenant(name string) error {
	query := "INSERT INTO tenant (name) VALUES ($1)"
	_, err := postgres.db.Exec(query, name)

	if pgErr, ok := err.(*pq.Error); ok {
		// 23505 corresponds to the Unique Violation error
		if pgErr.Code == "23505" {
			return &errors.ClientError{
				Code:    "SINGLE-ATTRIBUTE-UNIQUE-VIOLATION",
				Message: fmt.Sprintf(errors.SingleAttributeUniqueViolation, pgErr.Table, name, name),
			}
		} else {
			return &errors.InternalError{
				ErrorStack: err.Error(),
			}
		}
	} else if err != nil {
		return &errors.InternalError{
			ErrorStack: err.Error(),
		}
	}

	return nil
}

func (postgres *postgresStorage) CreateDivision(tenant string, name string) error {
	query := "INSERT INTO division (tenant, name) VALUES ($1, $2)"
	_, err := postgres.db.Exec(query, tenant, name)
	if pgErr, ok := err.(*pq.Error); ok {
		// 23505 corresponds to the
		switch pgErr.Code {
		case "23505":
			//Unique Violation error
			existingDivision := fmt.Sprintf(`tenant "%s" & division "%s"`, tenant, name)
			return &errors.ClientError{
				Code:    "MULTI-ATTRIBUTE-UNIQUE-VIOLATION",
				Message: fmt.Sprintf(errors.MultiAttributeUniqueViolation, pgErr.Table, existingDivision),
			}
		case "23503":
			//Foreign Key Violation error
			return &errors.ClientError{
				Code:    "INSERT-FOREIGN-KEY-VIOLATION",
				Message: fmt.Sprintf(errors.InsertForeignKeyViolation, tenant, "tenant"),
			}
		default:
			return &errors.InternalError{
				ErrorStack: err.Error(),
			}
		}
	} else if err != nil {
		return &errors.InternalError{
			ErrorStack: err.Error(),
		}
	}

	return nil
}

func (postgres *postgresStorage) CreateDepartment(tenant string, division string, name string) error {
	query := "INSERT INTO department (tenant, division, name) VALUES ($1, $2, $3)"
	_, err := postgres.db.Exec(query, tenant, division, name)
	if pgErr, ok := err.(*pq.Error); ok {
		//TODO fill in error handling
		switch pgErr.Code {
		//Unique violation
		case "23505":
			departmentExists := fmt.Sprintf(`tenant "%s", division "%s", and department "%s"`, tenant, division, name)
			return &errors.ClientError{
				Code: "MULTI-ATTRIBUTE-UNIQUE-VIOLATION",
				Message: fmt.Sprintf(errors.MultiAttributeUniqueViolation, "department", departmentExists),
			}
		// Foreign key violation
		case "23503":
			invalidTenantDivision := fmt.Sprintf("%s-%s", tenant, division)
			return &errors.ClientError{
				Code: "INSERT-FOREIGN-KEY-VIOLATION",
				Message: fmt.Sprintf(errors.InsertForeignKeyViolation, invalidTenantDivision, "tenant-division combination"),
			}
		default:
			return &errors.InternalError{
				ErrorStack: pgErr.Error(),
			}
		}
	} else if err != nil {
		return &errors.InternalError{
			ErrorStack: err.Error(),
		}
	}

	return nil

}
