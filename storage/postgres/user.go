package postgres

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"multi-tenant-HR-information-system-backend/httperror"
	"multi-tenant-HR-information-system-backend/storage"
)

func (postgres *postgresStorage) CreateUser(user storage.User) error {
	query := `
		INSERT INTO user_account (id, tenant_id, email, password, totp_secret_key) 
		VALUES ($1, $2, $3, $4, $5)`
	_, err := postgres.db.Exec(query, user.Id, user.TenantId, user.Email, user.Password, user.TotpSecretKey)
	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505":
			// Unique Violation
			return NewUniqueViolationError("user", pgErr)
		case "23503":
			// Foreign Key Violation
			return NewInvalidForeignKeyError(pgErr)
		default:
			return httperror.NewInternalServerError(pgErr)
		}
	} else if err != nil {
		return httperror.NewInternalServerError(err)
	}

	return nil

}

func (postgres postgresStorage) GetUsers(userFilter storage.User) ([]storage.User, error) {
	// All queries must be conditional on the tenantId
	if userFilter.TenantId == "" {
		return nil, httperror.NewInternalServerError(errors.New("TenantId must be provided to postgres model"))
	}

	conditions := []string{"tenant_id"}
	values := []any{userFilter.TenantId}

	if userFilter.Id != "" {
		conditions = append(conditions, "id")
		values = append(values, userFilter.Id)
	}

	if userFilter.Email != "" {
		conditions = append(conditions, "email")
		values = append(values, userFilter.Email)
	}

	query := NewDynamicConditionQuery("SELECT * FROM user_account", conditions)

	rows, err := postgres.db.Query(query, values...)
	if err != nil {
		return nil, httperror.NewInternalServerError(err)
	}
	defer rows.Close()

	var fetchedUsers []storage.User

	for rows.Next() {
		var user storage.User
		var lastLogin sql.NullString // last_login may be null

		if err := rows.Scan(&user.Id, &user.Email, &user.TenantId, &user.Password,
			&user.TotpSecretKey, &user.CreatedAt, &user.UpdatedAt, &lastLogin); err != nil {
			return nil, httperror.NewInternalServerError(err)
		}

		user.LastLogin = lastLogin.String

		fetchedUsers = append(fetchedUsers, user)
	}

	return fetchedUsers, nil
}

func (postgres *postgresStorage) CreateAppointment(appointment storage.Appointment) error {
	var err error

	if appointment.EndDate != "" {
		query := `
		INSERT INTO appointment (id, tenant_id, title, department_id, user_account_id, start_date, end_date) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err = postgres.db.Exec(query, appointment.Id, appointment.TenantId, appointment.Title, appointment.DepartmentId, appointment.UserId, appointment.StartDate, appointment.EndDate)

	} else {
		query := `
		INSERT INTO appointment (id, tenant_id, title, department_id, user_account_id, start_date) 
		VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = postgres.db.Exec(query, appointment.Id, appointment.TenantId, appointment.Title, appointment.DepartmentId, appointment.UserId, appointment.StartDate)

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
			return httperror.NewInternalServerError(pgErr)
		}
	} else if err != nil {
		return httperror.NewInternalServerError(err)
	}

	return nil
}
