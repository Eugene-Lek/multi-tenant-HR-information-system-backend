package postgres

import (
	"fmt"
	
	"github.com/lib/pq"

	"multi-tenant-HR-information-system-backend/routes"
)

func (postgres *postgresStorage) CreateUser(user routes.User) error {
	query := `
		INSERT INTO user_account (id, email, tenant, division, department, password, totp_secret_key) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := postgres.db.Exec(query, user.Id, user.Email, user.Tenant, user.Division, user.Department, user.Password, user.TotpSecretKey)
	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505": 
			// Unique Violation
			fmt.Println(pgErr.Detail)
			return NewUniqueViolationError("user", pgErr)
		case "23503":
			// Foreign Key Violation	
			return NewInvalidForeignKeyError(pgErr)
		default: 
			return routes.NewInternalServerError(pgErr)
		}
	} else if err != nil {
		return routes.NewInternalServerError(err)
	}

	return nil

}


func (postgres *postgresStorage) CreateAppointment(appointment routes.Appointment) error {
	var err error

	if appointment.EndDate != "" {
		query := `
		INSERT INTO appointment (title, tenant, division, department, user_account_id, start_date, end_date) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err = postgres.db.Exec(query, appointment.Title, appointment.Tenant, appointment.Division, appointment.Department, appointment.UserId, appointment.StartDate, appointment.EndDate)

	} else {
		query := `
		INSERT INTO appointment (title, tenant, division, department, user_account_id, start_date) 
		VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = postgres.db.Exec(query, appointment.Title, appointment.Tenant, appointment.Division, appointment.Department, appointment.UserId, appointment.StartDate)		

	}

	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505": 
			// Unique Violation
			return NewUniqueViolationError("appointment", pgErr)
		case "23503":
			// Foreign Key Violation	
			return NewInvalidForeignKeyError(pgErr)
		default: 
			return routes.NewInternalServerError(pgErr)
		}
	} else if err != nil {
		return routes.NewInternalServerError(err)
	}

	return nil
}