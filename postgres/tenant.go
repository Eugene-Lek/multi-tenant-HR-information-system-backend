package postgres

import (
	"fmt"

	"github.com/lib/pq"

	"multi-tenant-HR-information-system-backend/errors"
	"multi-tenant-HR-information-system-backend/routes"
)

func (postgres *postgresStorage) CreateTenant(tenant routes.Tenant) error {
	query := "INSERT INTO tenant (name) VALUES ($1)"
	_, err := postgres.db.Exec(query, tenant.Name)

	if pgErr, ok := err.(*pq.Error); ok {
		// 23505 corresponds to the Unique Violation error
		if pgErr.Code == "23505" {
			return &errors.ClientError{
				Code:    "SINGLE-ATTRIBUTE-UNIQUE-VIOLATION",
				Message: fmt.Sprintf(errors.SingleAttributeUniqueViolation, pgErr.Table, tenant.Name, tenant.Name),
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

func (postgres *postgresStorage) CreateDivision(division routes.Division) error {
	query := "INSERT INTO division (tenant, name) VALUES ($1, $2)"
	_, err := postgres.db.Exec(query, division.Tenant, division.Name)
	if pgErr, ok := err.(*pq.Error); ok {
		// 23505 corresponds to the
		switch pgErr.Code {
		case "23505":
			//Unique Violation error
			divisionIdentifiers := fmt.Sprintf(`tenant "%s" & division "%s"`, division.Tenant, division.Name)
			return &errors.ClientError{
				Code:    "MULTI-ATTRIBUTE-UNIQUE-VIOLATION",
				Message: fmt.Sprintf(errors.MultiAttributeUniqueViolation, pgErr.Table, divisionIdentifiers),
			}
		case "23503":
			//Foreign Key Violation error
			return &errors.ClientError{
				Code:    "INSERT-FOREIGN-KEY-VIOLATION",
				Message: fmt.Sprintf(errors.InsertForeignKeyViolation, division.Tenant, "tenant"),
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

func (postgres *postgresStorage) CreateDepartment(department routes.Department) error {
	query := "INSERT INTO department (tenant, division, name) VALUES ($1, $2, $3)"
	_, err := postgres.db.Exec(query, department.Tenant, department.Division, department.Name)
	if pgErr, ok := err.(*pq.Error); ok {
		//TODO fill in error handling
		switch pgErr.Code {
		//Unique violation
		case "23505":
			departmentIdentifiers := fmt.Sprintf(`tenant "%s", division "%s", and department "%s"`, department.Tenant, department.Division, department.Name)
			return &errors.ClientError{
				Code: "MULTI-ATTRIBUTE-UNIQUE-VIOLATION",
				Message: fmt.Sprintf(errors.MultiAttributeUniqueViolation, "department",departmentIdentifiers),
			}
		// Foreign key violation
		case "23503":
			tenantDivisionCombination := fmt.Sprintf("%s-%s", department.Tenant, department.Division)
			return &errors.ClientError{
				Code: "INSERT-FOREIGN-KEY-VIOLATION",
				Message: fmt.Sprintf(errors.InsertForeignKeyViolation, tenantDivisionCombination, "tenant-division combination"),
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
