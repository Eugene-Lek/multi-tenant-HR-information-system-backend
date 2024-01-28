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

func (postgres *postgresStorage) CreateJobApplication(jobApplication storage.JobApplication) error {
	query := `INSERT INTO job_application 
				(id, tenant_id, job_requisition_id, first_name, last_name, country_code, phone_number, email, resume_s3_url)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := postgres.db.Exec(query, jobApplication.Id, jobApplication.TenantId, jobApplication.JobRequisitionId,
		jobApplication.FirstName, jobApplication.LastName, jobApplication.CountryCode,
		jobApplication.PhoneNumber, jobApplication.Email, jobApplication.ResumeS3Url)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			switch pgErr.Code {
			case "23505":
				// Unique Violation
				return NewUniqueViolationError("job application", pgErr)
			case "23503":
				// Foreign Key Violation
				return NewInvalidForeignKeyError(pgErr)
			default:
				return httperror.NewInternalServerError(pgErr)
			}
		} else {
			return httperror.NewInternalServerError(err)
		}
	}

	return nil
}

func (postgres *postgresStorage) GetJobApplications(filter storage.JobApplication) ([]storage.JobApplication, error) {
	// All queries must be conditional on the tenantId
	if filter.TenantId == "" {
		return nil, httperror.NewInternalServerError(errors.New("TenantId must be provided to postgres model"))
	}
	conditions := []string{"tenant_id = $1"}
	filterByValues := []any{filter.TenantId}

	if filter.Id != "" {
		conditions = append(conditions, fmt.Sprintf("id = $%v", len(conditions)+1))
		filterByValues = append(filterByValues, filter.Id)
	}
	if filter.JobRequisitionId != "" {
		conditions = append(conditions, fmt.Sprintf("job_requisition_id = $%v", len(conditions)+1))
		filterByValues = append(filterByValues, filter.JobRequisitionId)
	}
	if filter.FirstName != "" {
		conditions = append(conditions, fmt.Sprintf("first_name = $%v", len(conditions)+1))
		filterByValues = append(filterByValues, filter.FirstName)
	}
	if filter.LastName != "" {
		conditions = append(conditions, fmt.Sprintf("last_name = $%v", len(conditions)+1))
		filterByValues = append(filterByValues, filter.LastName)
	}
	if filter.CountryCode != "" {
		conditions = append(conditions, fmt.Sprintf("country_code = $%v", len(conditions)+1))
		filterByValues = append(filterByValues, filter.CountryCode)
	}
	if filter.PhoneNumber != "" {
		conditions = append(conditions, fmt.Sprintf("phone_number = $%v", len(conditions)+1))
		filterByValues = append(filterByValues, filter.PhoneNumber)
	}
	if filter.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email = $%v", len(conditions)+1))
		filterByValues = append(filterByValues, filter.Email)
	}
	if filter.RecruiterDecision != "" {
		conditions = append(conditions, fmt.Sprintf("recruiter_decision = $%v", len(conditions)+1))
		filterByValues = append(filterByValues, filter.RecruiterDecision)
	}
	if filter.InterviewDate != "" {
		conditions = append(conditions, fmt.Sprintf("interview_date = $%v", len(conditions)+1))
		filterByValues = append(filterByValues, filter.InterviewDate)
	}
	if filter.HiringManagerDecision != "" {
		conditions = append(conditions, fmt.Sprintf("hiring_manager_decision = $%v", len(conditions)+1))
		filterByValues = append(filterByValues, filter.HiringManagerDecision)
	}
	if filter.OfferStartDate != "" {
		conditions = append(conditions, fmt.Sprintf("offer_start_date <= $%v", len(conditions)+1))
		filterByValues = append(filterByValues, filter.OfferStartDate)
	}
	if filter.OfferEndDate != "" {
		conditions = append(conditions, fmt.Sprintf("offer_end_date >= $%v", len(conditions)+1))
		filterByValues = append(filterByValues, filter.OfferEndDate)
	}
	if filter.ApplicantDecision != "" {
		conditions = append(conditions, fmt.Sprintf("applicant_decision = $%v", len(conditions)+1))
		filterByValues = append(filterByValues, filter.ApplicantDecision)
	}

	query := NewQueryWithFilter("SELECT * FROM job_application", conditions)
	rows, err := postgres.db.Query(query, filterByValues)
	if err != nil {
		return nil, httperror.NewInternalServerError(err)
	}

	jobApplications := []storage.JobApplication{}
	for rows.Next() {
		var jobApplication storage.JobApplication
		var interviewDate sql.NullString
		var offerStartDate sql.NullString
		var offerEndDate sql.NullString				

		err := rows.Scan(
			&jobApplication.Id,
			&jobApplication.TenantId,
			&jobApplication.JobRequisitionId,
			&jobApplication.FirstName,
			&jobApplication.LastName,
			&jobApplication.CountryCode,
			&jobApplication.PhoneNumber,
			&jobApplication.Email,
			&jobApplication.ResumeS3Url,
			&jobApplication.RecruiterDecision,
			&interviewDate,
			&jobApplication.HiringManagerDecision,
			&offerStartDate,
			&offerEndDate,
			&jobApplication.ApplicantDecision,
			&jobApplication.CreatedAt,
			&jobApplication.UpdatedAt,
		)

		if err != nil {
			return nil, httperror.NewInternalServerError(err)
		}

		jobApplication.InterviewDate = interviewDate.String
		jobApplication.OfferStartDate = offerStartDate.String
		jobApplication.OfferEndDate = offerEndDate.String				
		jobApplications = append(jobApplications, jobApplication)
	}

	return jobApplications, nil
}

func (postgres *postgresStorage) UpdateJobApplication(updatedValues storage.JobApplication, filter storage.JobApplication) error {
	checkConstraints := map[string]*httperror.Error{
		"ck_recruiter_shortlist_before_setting_interview_date": ErrMissingRecruiterShortlist,
		"ck_interview_date_set_before_hiring_manager_offer":    ErrMissingInterviewDate,
		"ck_hiring_manager_offer_before_applicant_acceptance":  ErrMissingHiringManagerOffer,
	}

	columnsToUpdate := []string{"updated_at"}
	newValues := []any{time.Now()}

	if updatedValues.RecruiterDecision != "" {
		columnsToUpdate = append(columnsToUpdate, "recruiter_decision")
		newValues = append(newValues, updatedValues.RecruiterDecision)
	}
	if updatedValues.InterviewDate != "" {
		columnsToUpdate = append(columnsToUpdate, "interview_date")
		newValues = append(newValues, updatedValues.InterviewDate)
	}
	if updatedValues.HiringManagerDecision != "" {
		columnsToUpdate = append(columnsToUpdate, "hiring_manager_decision")
		newValues = append(newValues, updatedValues.HiringManagerDecision)
	}
	if filter.OfferStartDate != "" {
		columnsToUpdate = append(columnsToUpdate, "offer_start_date")
		newValues = append(newValues, filter.OfferStartDate)
	}
	if filter.OfferEndDate != "" {
		columnsToUpdate = append(columnsToUpdate, "offer_end_date")
		newValues = append(newValues, filter.OfferEndDate)
	}
	if updatedValues.ApplicantDecision != "" {
		columnsToUpdate = append(columnsToUpdate, "applicant_decision")
		newValues = append(newValues, updatedValues.ApplicantDecision)
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
	if filter.JobRequisitionId != "" {
		columnsToFilterBy = append(columnsToFilterBy, "job_requisition_id")
		filterByValues = append(filterByValues, filter.JobRequisitionId)
	}

	query := NewUpdateQuery("job_application", columnsToUpdate, columnsToFilterBy)
	values := []any{}
	values = append(values, newValues...)
	values = append(values, filterByValues...)

	result, err := postgres.db.Exec(query, values...)
	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505":
			// Unique Violation
			return NewUniqueViolationError("job application", pgErr)
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
		return New404NotFoundError("job application")
	}

	return nil
}

// Creates the new user account, assigns the position to it, updates the job req to reflect that it is filled,
// and updates the job application to reflect that the applicant has accepted it
func (postgres *postgresStorage) OnboardNewHire(jobApplication storage.JobApplication, newUser storage.User) error {
	tx, err := postgres.db.Begin()
	if err != nil {
		return httperror.NewInternalServerError(err)
	}

	createUser := "INSERT INTO user account (id, tenant_id, email, password, totp_secret_key)"
	_, err = tx.Exec(createUser, newUser.Id, newUser.TenantId, newUser.Email, newUser.Password, newUser.TotpSecretKey)
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

	var positionId string
	getPositionId := "SELECT position_id FROM job_requisition WHERE id = $1 AND tenant_id = $2"
	err = tx.QueryRow(getPositionId, jobApplication.JobRequisitionId, jobApplication.TenantId).Scan(&positionId)
	if err == sql.ErrNoRows {
		return New404NotFoundError("job requisition")
	}
	if err != nil {
		return httperror.NewInternalServerError(err)
	}

	if jobApplication.OfferEndDate == "" {
		assignUserToPosition := `INSERT INTO position_assignment (tenant_id, user_account_id, position_id, start_date)
		VALUES ($1, $2, $3, $4)`
		_, err = tx.Exec(assignUserToPosition, jobApplication.TenantId, newUser.Id, positionId, jobApplication.OfferStartDate)
	} else {
		assignUserToPosition := `INSERT INTO position_assignment (tenant_id, user_account_id, position_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)`
		_, err = tx.Exec(assignUserToPosition, jobApplication.TenantId, newUser.Id, positionId, jobApplication.OfferStartDate, jobApplication.OfferEndDate)		
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

	updateJobReqToFilled := `UPDATE job_requisition SET filled_by = $1, filled_at = now() WHERE id = $2 AND tenant_id = $3`
	result, err := tx.Exec(updateJobReqToFilled, newUser.Id, jobApplication.JobRequisitionId, jobApplication.TenantId)
	if err != nil {
		return httperror.NewInternalServerError(err)		
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return New404NotFoundError("job requisition")
	}

	updateJobApplicationToAccepted := `UPDATE job_application SET applicant_decision = 'ACCEPTED' WHERE id = $1 AND tenant_id = $2`
	result, err = tx.Exec(updateJobApplicationToAccepted, jobApplication.Id, jobApplication.TenantId)
	if err != nil {
		return httperror.NewInternalServerError(err)		
	}		
	if affected, _ := result.RowsAffected(); affected == 0 {
		return New404NotFoundError("job application")
	}	

	err = tx.Commit()
	if err != nil {
		return httperror.NewInternalServerError(err)
	}

	return nil
}
