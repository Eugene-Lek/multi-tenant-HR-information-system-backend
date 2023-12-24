package postgres

import (
	"github.com/lib/pq"

	"multi-tenant-HR-information-system-backend/routes"
	"multi-tenant-HR-information-system-backend/errors"
)

func (postgres *postgresStorage) CreateUser(user routes.User) errors.HttpError {
	query := `
		INSERT INTO user (id, email, tenant, division, department, password, totp_secret_key) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := postgres.db.Exec(query)
	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505": 
			// Unique Violation
			return errors.NewUniqueViolationError("user", [][2]string{{"tenant", user.Tenant}, {"email", user.Email}})
		case "23503":
			// Foreign Key Violation	
			return errors.NewInvalidForeignKeyError([][2]string{{"tenant", user.Tenant}, {"division", user.Division}, {"department", user.Department}})		
		default: 
			return errors.NewInternalServerError(pgErr.Error())
		}
	} else if err != nil {
		return errors.NewInternalServerError(err.Error())
	}

	return nil

}


func (postgres *postgresStorage) CreateAppointment(appointment routes.Appointment) errors.HttpError {
	query := `
		INSERT INTO appointment (tenant, division, department, user_id, start_date, end_date) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := postgres.db.Exec(query)
	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505": 
			// Unique Violation
			return errors.NewUniqueViolationError("appointment", [][2]string{{"tenant", appointment.Tenant}, {"division", appointment.Division}, {"department", appointment.Department}, {"title", appointment.Title}, {"userID", appointment.UserId}})
		case "23503":
			// Foreign Key Violation	
			return errors.NewInvalidForeignKeyError([][2]string{{"tenant", appointment.Tenant}, {"division", appointment.Division}, {"department", appointment.Department}, {"userID", appointment.UserId}})		
		default: 
			return errors.NewInternalServerError(pgErr.Error())
		}
	} else if err != nil {
		return errors.NewInternalServerError(err.Error())
	}

	return nil
}