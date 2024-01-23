package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	"multi-tenant-HR-information-system-backend/httperror"
	"multi-tenant-HR-information-system-backend/storage"
)

func (postgres *postgresStorage) CreateJobRequisition(jobRequisition storage.JobRequisition) error {
	query := `INSERT INTO job_requisition (id, tenant_id, title, department_id, job_description, job_requirements, requestor, supervisor, hr_approver)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := postgres.db.Exec(query,
		jobRequisition.Id, jobRequisition.TenantId, jobRequisition.Title, jobRequisition.DepartmentId, jobRequisition.JobDescription,
		jobRequisition.JobRequirements, jobRequisition.Requestor, jobRequisition.Supervisor, jobRequisition.HrApprover,
	)

	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505":
			// Unique Violation
			return NewUniqueViolationError("job requisition", pgErr)
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

func (postgres *postgresStorage) GetJobRequisitions(filter storage.JobRequisition) ([]storage.JobRequisition, error) {
	// All queries must be conditional on the tenantId
	if filter.TenantId == "" {
		return nil, httperror.NewInternalServerError(errors.New("TenantId must be provided to postgres model"))
	}

	conditions := []string{"tenant_id = $1"}
	values := []any{filter.TenantId}

	if filter.Id != "" {
		conditions = append(conditions, fmt.Sprintf("id = $%v", len(conditions)+1))
		values = append(values, filter.Id)
	}
	if filter.Title != "" {
		conditions = append(conditions, fmt.Sprintf("title = $%v", len(conditions)+1))
		values = append(values, filter.Title)
	}
	if filter.DepartmentId != "" {
		conditions = append(conditions, fmt.Sprintf("department_id = $%v", len(conditions)+1))
		values = append(values, filter.DepartmentId)
	}
	if filter.Requestor != "" {
		conditions = append(conditions, fmt.Sprintf("requestor = $%v", len(conditions)+1))
		values = append(values, filter.Requestor)
	}
	if filter.Supervisor != "" {
		conditions = append(conditions, fmt.Sprintf("supervisor = $%v", len(conditions)+1))
		values = append(values, filter.Supervisor)
	}
	if filter.SupervisorDecision != "" {
		conditions = append(conditions, fmt.Sprintf("supervisor_decision = $%v", len(conditions)+1))
		values = append(values, filter.SupervisorDecision)
	}
	if filter.HrApprover != "" {
		conditions = append(conditions, fmt.Sprintf("hr_approver = $%v", len(conditions)+1))
		values = append(values, filter.HrApprover)
	}
	if filter.HrApproverDecision != "" {
		conditions = append(conditions, fmt.Sprintf("hr_approver_decision = $%v", len(conditions)+1))
		values = append(values, filter.HrApproverDecision)
	}
	if filter.Recruiter != "" {
		conditions = append(conditions, fmt.Sprintf("recruiter = $%v", len(conditions)+1))
		values = append(values, filter.Recruiter)
	}
	if filter.FilledBy != "" {
		conditions = append(conditions, fmt.Sprintf("filled_by = $%v", len(conditions)+1))
		values = append(values, filter.FilledBy)
	}

	query := NewQueryWithFilter("SELECT * FROM job_requisition", conditions)
	rows, err := postgres.db.Query(query, values...)
	if err != nil {
		return nil, httperror.NewInternalServerError(err)
	}
	defer rows.Close()

	jobRequisitions := []storage.JobRequisition{}

	for rows.Next() {
		var jobRequisition storage.JobRequisition

		var recruiter sql.NullString
		var filledBy sql.NullString
		var filledAt sql.NullTime

		err := rows.Scan(
			&jobRequisition.Id,
			&jobRequisition.TenantId,
			&jobRequisition.Title,
			&jobRequisition.DepartmentId,
			&jobRequisition.JobDescription,
			&jobRequisition.JobRequirements,
			&jobRequisition.Requestor,
			&jobRequisition.Supervisor,
			&jobRequisition.SupervisorDecision,
			&jobRequisition.HrApprover,
			&jobRequisition.HrApproverDecision,
			&recruiter,
			&filledBy,
			&filledAt,
			&jobRequisition.CreatedAt,
			&jobRequisition.UpdatedAt,
		)

		if err != nil {
			return nil, httperror.NewInternalServerError(err)
		}

		jobRequisition.Recruiter = recruiter.String
		jobRequisition.FilledBy = filledBy.String
		jobRequisition.FilledAt = filledAt.Time

		jobRequisitions = append(jobRequisitions, jobRequisition)
	}

	return jobRequisitions, nil
}

func (postgres *postgresStorage) UpdateJobRequisition(updatedValues storage.JobRequisition, filter storage.JobRequisition) error {
	checkConstraints := map[string]*httperror.Error{
		"ck_hr_approval_only_with_supervisor_approval":  ErrMissingSupervisorApproval,
		"ck_recruiter_assignment_only_with_hr_approval": ErrMissingHrApproval,
		"ck_req_filled_only_with_hr_approval":           ErrMissingHrApproval,
		"ck_req_filled_at_only_with_hr_approval":        ErrMissingHrApproval,
	}

	columnsToUpdate := []string{"updated_at"}
	newValues := []any{time.Now()}

	if updatedValues.TenantId != "" {
		columnsToUpdate = append(columnsToUpdate, "tenant_id")
		newValues = append(newValues, updatedValues.TenantId)
	}
	if updatedValues.Id != "" {
		columnsToUpdate = append(columnsToUpdate, "id")
		newValues = append(newValues, updatedValues.Id)
	}
	if updatedValues.Title != "" {
		columnsToUpdate = append(columnsToUpdate, "title")
		newValues = append(newValues, updatedValues.Title)
	}
	if updatedValues.DepartmentId != "" {
		columnsToUpdate = append(columnsToUpdate, "department_id")
		newValues = append(newValues, updatedValues.DepartmentId)
	}
	if updatedValues.JobDescription != "" {
		columnsToUpdate = append(columnsToUpdate, "job_description")
		newValues = append(newValues, updatedValues.JobDescription)
	}
	if updatedValues.JobRequirements != "" {
		columnsToUpdate = append(columnsToUpdate, "job_requirements")
		newValues = append(newValues, updatedValues.JobRequirements)
	}
	if updatedValues.Requestor != "" {
		columnsToUpdate = append(columnsToUpdate, "requestor")
		newValues = append(newValues, updatedValues.Requestor)
	}
	if updatedValues.Supervisor != "" {
		columnsToUpdate = append(columnsToUpdate, "supervisor")
		newValues = append(newValues, updatedValues.Supervisor)
	}
	if updatedValues.SupervisorDecision != "" {
		columnsToUpdate = append(columnsToUpdate, "supervisor_decision")
		newValues = append(newValues, updatedValues.SupervisorDecision)
	}
	if updatedValues.HrApprover != "" {
		columnsToUpdate = append(columnsToUpdate, "hr_approver")
		newValues = append(newValues, updatedValues.HrApprover)
	}
	if updatedValues.HrApproverDecision != "" {
		columnsToUpdate = append(columnsToUpdate, "hr_approver_decision")
		newValues = append(newValues, updatedValues.HrApproverDecision)
	}
	if updatedValues.Recruiter != "" {
		columnsToUpdate = append(columnsToUpdate, "recruiter")
		newValues = append(newValues, updatedValues.Recruiter)
	}
	if updatedValues.FilledBy != "" {
		columnsToUpdate = append(columnsToUpdate, "filled_by")
		newValues = append(newValues, updatedValues.FilledBy)
	}
	if !updatedValues.FilledAt.IsZero() {
		columnsToUpdate = append(columnsToUpdate, "filled_at")
		newValues = append(newValues, updatedValues.FilledAt)
	}
	if updatedValues.CreatedAt != "" {
		columnsToUpdate = append(columnsToUpdate, "created_at")
		newValues = append(newValues, updatedValues.CreatedAt)
	}
	if updatedValues.UpdatedAt != "" {
		columnsToUpdate = append(columnsToUpdate, "updated_at")
		newValues = append(newValues, updatedValues.UpdatedAt)
	}

	// All queries must be conditional on the tenantId
	if filter.TenantId == "" {
		return httperror.NewInternalServerError(errors.New("TenantId must be provided to postgres model"))
	}
	columnsToFilterBy := []string{"tenant_id"}
	filterByValues := []any{filter.TenantId}

	if filter.Id != "" {
		columnsToFilterBy = append(columnsToFilterBy, "id")
		filterByValues = append(filterByValues, filter.Id)
	}
	if filter.Title != "" {
		columnsToFilterBy = append(columnsToFilterBy, "title")
		filterByValues = append(filterByValues, filter.Title)
	}
	if filter.DepartmentId != "" {
		columnsToFilterBy = append(columnsToFilterBy, "department_id")
		filterByValues = append(filterByValues, filter.DepartmentId)
	}
	if filter.Requestor != "" {
		columnsToFilterBy = append(columnsToFilterBy, "requestor")
		filterByValues = append(filterByValues, filter.Requestor)
	}
	if filter.Supervisor != "" {
		columnsToFilterBy = append(columnsToFilterBy, "supervisor")
		filterByValues = append(filterByValues, filter.Supervisor)
	}
	if filter.SupervisorDecision != "" {
		columnsToFilterBy = append(columnsToFilterBy, "supervisor_decision")
		filterByValues = append(filterByValues, filter.SupervisorDecision)
	}
	if filter.HrApprover != "" {
		columnsToFilterBy = append(columnsToFilterBy, "hr_approver")
		filterByValues = append(filterByValues, filter.HrApprover)
	}
	if filter.HrApproverDecision != "" {
		columnsToFilterBy = append(columnsToFilterBy, "hr_approver_decision")
		filterByValues = append(filterByValues, filter.HrApproverDecision)
	}
	if filter.Recruiter != "" {
		columnsToFilterBy = append(columnsToFilterBy, "recruiter")
		filterByValues = append(filterByValues, filter.Recruiter)
	}
	if filter.FilledBy != "" {
		columnsToFilterBy = append(columnsToFilterBy, "filled_by")
		filterByValues = append(filterByValues, filter.FilledBy)
	}

	values := []any{}
	values = append(values, newValues...)
	values = append(values, filterByValues...)

	query := NewUpdateQuery("job_requisition", columnsToUpdate, columnsToFilterBy)
	result, err := postgres.db.Exec(query, values...)
	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505":
			// Unique Violation
			return NewUniqueViolationError("job requisition", pgErr)
		case "23503":
			// Foreign Key Violation
			return NewInvalidForeignKeyError(pgErr)
		case "23514":
			// Check violation
			err, ok := checkConstraints[pgErr.Constraint]
			if !ok {
				return httperror.NewInternalServerError(pgErr)
			}
			return err

		default:
			return httperror.NewInternalServerError(pgErr)
		}
	} else if err != nil {
		return httperror.NewInternalServerError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return httperror.NewInternalServerError(err)
	}
	if rowsAffected == 0 {
		return New404NotFoundError("job requisition")
	}

	return nil
}
