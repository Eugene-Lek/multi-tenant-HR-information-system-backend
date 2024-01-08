package postgres

import (
	"database/sql"

	"github.com/lib/pq"

	"multi-tenant-HR-information-system-backend/routes"
)

func (postgres *postgresStorage) CreateUser(user routes.User) error {
	query := `
		INSERT INTO user_account (id, email, tenant, password, totp_secret_key) 
		VALUES ($1, $2, $3, $4, $5)`
	_, err := postgres.db.Exec(query, user.Id, user.Email, user.Tenant, user.Password, user.TotpSecretKey)
	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505":
			// Unique Violation
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

func (postgres postgresStorage) GetUsers(userFilter routes.User) ([]routes.User, error) {
	conditions := []string{}
	values := []any{}

	if userFilter.Id != "" {
		conditions = append(conditions, "id")
		values = append(values, userFilter.Id)
	}

	if userFilter.Email != "" {
		conditions = append(conditions, "email")
		values = append(values, userFilter.Email)
	}

	if userFilter.Tenant != "" {
		conditions = append(conditions, "tenant")
		values = append(values, userFilter.Tenant)
	}

	query := NewDynamicConditionQuery("SELECT * FROM user_account", conditions)

	rows, err := postgres.db.Query(query, values...)
	if err != nil {
		return nil, routes.NewInternalServerError(err)
	}
	defer rows.Close()

	var fetchedUsers []routes.User

	for rows.Next() {
		var user routes.User
		var lastLogin sql.NullString // last_login may be null

		if err := rows.Scan(&user.Id, &user.Email, &user.Tenant, &user.Password,
			&user.TotpSecretKey, &user.CreatedAt, &user.UpdatedAt, &lastLogin); err != nil {
			return nil, routes.NewInternalServerError(err)
		}

		user.LastLogin = lastLogin.String

		fetchedUsers = append(fetchedUsers, user)
	}

	return fetchedUsers, nil
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
