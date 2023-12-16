package postgres

import (
	"fmt"

	"github.com/lib/pq"

	"multi-tenant-HR-information-system-backend/routes"
	"multi-tenant-HR-information-system-backend/errors"
)

func (postgres *postgresStorage) CreateUser(user routes.User) error {
	query := `
		INSERT INTO user (id, email, tenant, division, department, password, totp_secret_key) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := postgres.db.Exec(query)
	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505": 
			// Unique Violation
			userIdentifiers:= `tenant "%s" and email "%s"`
			return &errors.ClientError{
				Code: "MULTI-ATTRIBUTE-UNIQUE-VIOLATION",
				Message: fmt.Sprintf(errors.MultiAttributeUniqueViolation, "user", userIdentifiers),
			}
		case "23503":
			// Foreign Key Violation
			tenantDivisionDepartmentCombination := fmt.Sprintf("%s-%s-%s", user.Tenant, user.Division, user.Department)
			return &errors.ClientError{
				Code: "INSERT-FOREIGN-KEY-VIOLATION",
				Message: fmt.Sprintf(errors.InsertForeignKeyViolation, tenantDivisionDepartmentCombination, "tenant-division-department combination"),
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


func (postgres *postgresStorage) CreateAppointment(appointment routes.Appointment) error {
	query := `
		INSERT INTO appointment (tenant, division, department, user_id, start_date, end_date) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := postgres.db.Exec(query)
	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505":
			//Unique Violation
			appointmentIdentifiers := fmt.Sprintf(`tenant "%s", division "%s", department "%s", title "%s" and user-id "%s"`, 
												appointment.Tenant, appointment.Division, appointment.Department, appointment.Title, appointment.UserId)
			return &errors.ClientError{
				Code: "MULTI-ATTRIBUTE-UNIQUE-VIOLATION",
				Message: fmt.Sprintf(errors.MultiAttributeUniqueViolation, "appointment", appointmentIdentifiers),
			}
		case "23503":
			//Foreign Key Violation
			tenantDivisionDepartmentUserCombination := fmt.Sprintf(`%s-%s-%s-%s`, 
												appointment.Tenant, appointment.Division, appointment.Department, appointment.UserId)			
			return &errors.ClientError{
				Code: "INSERT-FOREIGN-KEY-VIOLATION",
				Message: fmt.Sprintf(errors.InsertForeignKeyViolation, tenantDivisionDepartmentUserCombination, "tenant-division-department-user combination"),
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