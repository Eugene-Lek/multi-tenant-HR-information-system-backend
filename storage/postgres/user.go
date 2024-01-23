package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

	conditions := []string{"tenant_id = $1"}
	values := []any{userFilter.TenantId}

	if userFilter.Id != "" {
		conditions = append(conditions, fmt.Sprintf("id = $%v", len(conditions)+1))
		values = append(values, userFilter.Id)
	}

	if userFilter.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email = $%v", len(conditions)+1))
		values = append(values, userFilter.Email)
	}

	query := NewQueryWithFilter("SELECT * FROM user_account", conditions)

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

func (postgres *postgresStorage) GetUserSupervisors(userId string, tenantId string) ([]string, error) {
	// subordinate & supervisor position assignments are only joined together (via subordinate_supervisor_relationship)
	// if they are both current (i.e. end date >= today)
	query := `
		SELECT array_agg(supervisor_assignment.user_account_id) AS supervisor_ids
		FROM subordinate_supervisor_relationship AS ssr	
		INNER JOIN position_assignment AS subordinate_assignment 
			ON subordinate_assignment.position_id = ssr.subordinate_position_id
			AND subordinate_assignment.user_account_id = $1	
			AND subordinate_assignment.tenant_id = $2	
			AND subordinate_assignment.end_date >= CURRENT_DATE
		INNER JOIN position_assignment AS supervisor_assignment 
			ON supervisor_assignment.position_id = ssr.supervisor_position_id
			AND supervisor_assignment.tenant_id = $2	
			AND supervisor_assignment.end_date >= CURRENT_DATE										
		GROUP BY subordinate_assignment.user_account_id
	`

	var supervisors []string
	err := postgres.db.QueryRow(query, userId, tenantId).Scan(pq.Array(&supervisors))
	if err != nil && err != sql.ErrNoRows {
		return nil, httperror.NewInternalServerError(err)
	}

	return supervisors, nil
}

func (postgres *postgresStorage) CreatePosition(position storage.Position) error {

	tx, err := postgres.db.Begin()
	if err != nil {
		return httperror.NewInternalServerError(err)
	}
	defer tx.Rollback() // Will have no effect if tx.Commit() is called

	// Insert the position
	query := "INSERT INTO position (id, tenant_id, title, department_id) VALUES ($1, $2, $3, $4)"
	_, err = tx.Exec(query, position.Id, position.TenantId, position.Title, position.DepartmentId)

	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505":
			// Unique Violation
			return NewUniqueViolationError("position", pgErr)
		case "23503":
			// Foreign Key Violation
			return NewInvalidForeignKeyError(pgErr)
		default:
			return httperror.NewInternalServerError(pgErr)
		}
	} else if err != nil {
		return httperror.NewInternalServerError(err)
	}

	// Insert the subordinate-supervisor relations, if any
	if len(position.SupervisorPositionIds) > 0 {
		// TODO: ensure that the supervisor is either from the same department or division HQ

		identifiers := []string{}
		values := []any{}

		for i, supervisorId := range position.SupervisorPositionIds {
			values = append(values, position.Id, supervisorId)
			identifiers = append(identifiers, fmt.Sprintf("($%v, $%v)", i*2+1, i*2+2))
		}

		query = "INSERT INTO subordinate_supervisor_relationship (subordinate_position_id, supervisor_position_id) VALUES " + strings.Join(identifiers, ", ")
		_, err = tx.Exec(query, values...)

		if pgErr, ok := err.(*pq.Error); ok {
			switch pgErr.Code {
			case "23505":
				// Unique Violation
				return NewUniqueViolationError("subordinate-supervisor relationship", pgErr)
			case "23503":
				// Foreign Key Violation
				return NewInvalidForeignKeyError(pgErr)
			case "23514":
				// Check violation
				return &httperror.Error{
					Status:  400,
					Message: "Subordinate Position and Supervisor Position cannot be the same",
					Code:    "INVALID-SUBORDINATE-SUPERVISOR-PAIR-ERROR",
				}
			default:
				return httperror.NewInternalServerError(pgErr)
			}
		} else if err != nil {
			return httperror.NewInternalServerError(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return httperror.NewInternalServerError(err)
	}

	return nil
}

func (postgres *postgresStorage) CreatePositionAssignment(positionAssignment storage.PositionAssignment) error {
	var err error

	if positionAssignment.EndDate != "" {
		query := `
		INSERT INTO position_assignment (tenant_id, position_id, user_account_id, start_date, end_date) 
		VALUES ($1, $2, $3, $4, $5)
		`
		_, err = postgres.db.Exec(query, positionAssignment.TenantId, positionAssignment.PositionId, positionAssignment.UserId, positionAssignment.StartDate, positionAssignment.EndDate)

	} else {
		query := `
		INSERT INTO position_assignment (tenant_id,  position_id, user_account_id, start_date) 
		VALUES ($1, $2, $3, $4)
		`
		_, err = postgres.db.Exec(query, positionAssignment.TenantId, positionAssignment.PositionId, positionAssignment.UserId, positionAssignment.StartDate)

	}

	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505":
			// Unique Violation
			return NewUniqueViolationError("position assignment", pgErr)
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

func (postgres *postgresStorage) GetUserPositions(userId string, filter storage.UserPosition) ([]storage.UserPosition, error) {
	// Subquery represents the positions that correspond to the user, inclusive of their respective supervisor position Ids
	baseQuery := `
	SELECT position.id, position.tenant_id, position.title, position.department_id, 
	position.supervisor_position_ids, position.start_date, position.end_date
	FROM (
		SELECT position_assignment.user_account_id, position.id, position.tenant_id, position.title, position.department_id, 
		array_agg(ssr.supervisor_position_id) AS supervisor_position_ids, 
		position_assignment.start_date, position_assignment.end_date
		FROM position_assignment
		INNER JOIN position 
			ON position_assignment.position_id = position.id
			AND position_assignment.user_account_id = $1					
		INNER JOIN subordinate_supervisor_relationship AS ssr	
			ON position.id = ssr.subordinate_position_id		
		GROUP BY position_assignment.user_account_id, position.id
	) AS position
	`

	// All queries must be conditional on the tenantId
	if filter.TenantId == "" {
		return nil, httperror.NewInternalServerError(errors.New("TenantId must be provided to postgres model"))
	}

	conditions := []string{"position.tenant_id = $2"}
	filterByValues := []any{userId, filter.TenantId}

	if filter.Title != "" {
		// Starting number is 2 because $1 is occupied by user id in the base query
		conditions = append(conditions, fmt.Sprintf("position.title = $%v", len(filterByValues)+1))
		filterByValues = append(filterByValues, filter.Title)
	}
	if filter.DepartmentId != "" {
		conditions = append(conditions, fmt.Sprintf("position.department_id = $%v", len(filterByValues)+1))
		filterByValues = append(filterByValues, filter.DepartmentId)
	}
	if len(filter.SupervisorPositionIds) != 0 {
		conditions = append(conditions, fmt.Sprintf("position.supervisor_position_ids && $%v", len(filterByValues)+1))

		supervisorIds := "{" + strings.Join(filter.SupervisorPositionIds, ",") + "}" // Create a literal of the postgres array
		filterByValues = append(filterByValues, supervisorIds)
	}
	if filter.StartDate != "" {
		conditions = append(conditions, fmt.Sprintf("position.start_date < $%v", len(filterByValues)+1))
		filterByValues = append(filterByValues, filter.DepartmentId)
	}
	if filter.EndDate != "" {
		conditions = append(conditions, fmt.Sprintf("position.end_date > $%v", len(filterByValues)+1))
		filterByValues = append(filterByValues, filter.DepartmentId)
	}

	query := NewQueryWithFilter(baseQuery, conditions)

	rows, err := postgres.db.Query(query, filterByValues...)
	if err != nil {
		return nil, httperror.NewInternalServerError(err)
	}
	defer rows.Close()

	positions := []storage.UserPosition{}

	for rows.Next() {
		var position storage.UserPosition

		err := rows.Scan(
			&position.Id, &position.TenantId, &position.Title, &position.DepartmentId, pq.Array(&position.SupervisorPositionIds),
			&position.StartDate, &position.EndDate)
		if err != nil {
			return nil, httperror.NewInternalServerError(err)
		}

		positions = append(positions, position)
	}

	return positions, nil
}
