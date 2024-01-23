package postgres

import (
	"database/sql"
	"log"
	"time"

	"multi-tenant-HR-information-system-backend/storage"
)

func (s *IntegrationTestSuite) TestCreateJobRequisition() {
	want := storage.JobRequisition{
		Id:              "cb180c6e-af87-4a97-9dcf-bcbe503414a7",
		TenantId:        s.defaultTenant.Id,
		Title:           "Database Administrator",
		DepartmentId:    s.defaultDepartment.Id,
		JobDescription:  "Manages databases of HRIS software",
		JobRequirements: "100 years of experience using postgres",
		Requestor:       s.defaultUser.Id,
		Supervisor:      s.defaultSupervisor.Id,
		HrApprover:      s.defaultHrApprover.Id,
	}

	err := s.postgres.CreateJobRequisition(want)
	s.Equal(nil, err)

	s.expectSelectQueryToReturnOneRow(
		"job_requisition",
		map[string]any{
			"id":              want.Id,
			"tenant_id":       want.TenantId,
			"title":           want.Title,
			"department_id":   want.DepartmentId,
			"job_description": want.JobDescription,
			"requestor":       want.Requestor,
			"supervisor":      want.Supervisor,
			"hr_approver":     want.HrApprover,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateJobRequisitionShouldHaveForeignKeyConstraints() {
	tests := []struct {
		name  string
		input storage.JobRequisition
	}{
		{
			"Should violate foreign key constraint because tenant id is invalid",
			storage.JobRequisition{
				Id:              "cb180c6e-af87-4a97-9dcf-bcbe503414a7",
				TenantId:        "1846f101-27fb-46c9-9128-38937fd0e2b5",
				Title:           "Database Administrator",
				DepartmentId:    s.defaultDepartment.Id,
				JobDescription:  "Manages databases of HRIS software",
				JobRequirements: "100 years of experience using postgres",
				Requestor:       s.defaultUser.Id,
				Supervisor:      s.defaultSupervisor.Id,
				HrApprover:      s.defaultHrApprover.Id,
			},
		},
		{
			"Should violate foreign key constraint because department id is invalid",
			storage.JobRequisition{
				Id:              "cb180c6e-af87-4a97-9dcf-bcbe503414a7",
				TenantId:        s.defaultTenant.Id,
				Title:           "Database Administrator",
				DepartmentId:    "1846f101-27fb-46c9-9128-38937fd0e2b5",
				JobDescription:  "Manages databases of HRIS software",
				JobRequirements: "100 years of experience using postgres",
				Requestor:       s.defaultUser.Id,
				Supervisor:      s.defaultSupervisor.Id,
				HrApprover:      s.defaultHrApprover.Id,
			},
		},
		{
			"Should violate foreign key constraint because requestor is invalid",
			storage.JobRequisition{
				Id:              "cb180c6e-af87-4a97-9dcf-bcbe503414a7",
				TenantId:        s.defaultTenant.Id,
				Title:           "Database Administrator",
				DepartmentId:    s.defaultDepartment.Id,
				JobDescription:  "Manages databases of HRIS software",
				JobRequirements: "100 years of experience using postgres",
				Requestor:       "1846f101-27fb-46c9-9128-38937fd0e2b5",
				Supervisor:      s.defaultSupervisor.Id,
				HrApprover:      s.defaultHrApprover.Id,
			},
		},
		{
			"Should violate foreign key constraint because supervisor id is invalid",
			storage.JobRequisition{
				Id:              "cb180c6e-af87-4a97-9dcf-bcbe503414a7",
				TenantId:        s.defaultTenant.Id,
				Title:           "Database Administrator",
				DepartmentId:    s.defaultDepartment.Id,
				JobDescription:  "Manages databases of HRIS software",
				JobRequirements: "100 years of experience using postgres",
				Requestor:       s.defaultUser.Id,
				Supervisor:      "1846f101-27fb-46c9-9128-38937fd0e2b5",
				HrApprover:      s.defaultHrApprover.Id,
			},
		},
		{
			"Should violate foreign key constraint because hr approver id is invalid",
			storage.JobRequisition{
				Id:              "cb180c6e-af87-4a97-9dcf-bcbe503414a7",
				TenantId:        s.defaultTenant.Id,
				Title:           "Database Administrator",
				DepartmentId:    s.defaultDepartment.Id,
				JobDescription:  "Manages databases of HRIS software",
				JobRequirements: "100 years of experience using postgres",
				Requestor:       s.defaultUser.Id,
				Supervisor:      s.defaultSupervisor.Id,
				HrApprover:      "1846f101-27fb-46c9-9128-38937fd0e2b5",
			},
		},
	}

	for _, test := range tests {
		err := s.postgres.CreateJobRequisition(test.input)
		s.expectErrorCode(err, "INVALID-FOREIGN-KEY-ERROR")

		s.expectSelectQueryToReturnNoRows(
			"job_requisition",
			map[string]any{
				"id":              test.input.Id,
				"tenant_id":       test.input.TenantId,
				"title":           test.input.Title,
				"department_id":   test.input.DepartmentId,
				"job_description": test.input.JobDescription,
				"requestor":       test.input.Requestor,
				"supervisor":      test.input.Supervisor,
				"hr_approver":     test.input.HrApprover,
			},
		)
	}
}

func (s *IntegrationTestSuite) TestUpdateJobRequisitionShouldHaveForeignKeyConstraint() {
	wantFilter := storage.JobRequisition{
		Id:         s.defaultJobRequisition.Id,
		TenantId:   s.defaultJobRequisition.TenantId,
		HrApprover: s.defaultJobRequisition.HrApprover,
	}

	wantUpdate := storage.JobRequisition{
		HrApproverDecision: "APPROVED",
		Recruiter:          "a4d052a0-7016-4863-9e25-a6f8d80be83b",
	}

	_, err := s.dbRootConn.Exec("UPDATE job_requisition SET supervisor_decision = 'APPROVED' WHERE id = $1", wantFilter.Id)
	if err != nil {
		log.Fatalf("Could not set supervisor approval: %s", err)
	}

	err = s.postgres.UpdateJobRequisition(wantUpdate, wantFilter)
	s.expectErrorCode(err, "INVALID-FOREIGN-KEY-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_requisition",
		map[string]any{
			"id":                   wantFilter.Id,
			"tenant_id":            wantFilter.TenantId,
			"hr_approver":          wantFilter.HrApprover,
			"hr_approver_decision": wantUpdate.HrApproverDecision,
			"recruiter":            wantUpdate.Recruiter,
		},
	)
}

func (s *IntegrationTestSuite) TestUpdateJobRequisitionShouldCheckForSupervisorApproval() {
	tests := []struct {
		name  string
		input storage.JobRequisition
	}{
		{
			"Should fail because Supervisor rejected",
			storage.JobRequisition{
				Id:                 s.defaultJobRequisition.Id,
				TenantId:           s.defaultJobRequisition.TenantId,
				SupervisorDecision: "REJECTED",
				HrApprover:         s.defaultJobRequisition.HrApprover,
				HrApproverDecision: "APPROVED",
				Recruiter:          s.defaultRecruiter.Id,
			},
		},
		{
			"Should fail because Supervisor approval is pending",
			storage.JobRequisition{
				Id:                 s.defaultJobRequisition.Id,
				TenantId:           s.defaultJobRequisition.TenantId,
				SupervisorDecision: "PENDING",
				HrApprover:         s.defaultJobRequisition.HrApprover,
				HrApproverDecision: "APPROVED",
				Recruiter:          s.defaultRecruiter.Id,
			},
		},
	}

	for _, test := range tests {
		s.dbRootConn.Exec("UPDATE job_requisition SET supervisor_decision = $1 WHERE id = $2", test.input.SupervisorDecision, test.input.Id)

		wantFilter := storage.JobRequisition{
			Id:         test.input.Id,
			TenantId:   test.input.TenantId,
			HrApprover: test.input.HrApprover,
		}

		wantUpdate := storage.JobRequisition{
			HrApproverDecision: test.input.HrApproverDecision,
			Recruiter:          test.input.Recruiter,
		}

		err := s.postgres.UpdateJobRequisition(wantUpdate, wantFilter)
		s.expectErrorCode(err, "MISSING-SUPERVISOR-APPROVAL-ERROR")

		s.expectSelectQueryToReturnNoRows(
			"job_requisition",
			map[string]any{
				"id":                   wantFilter.Id,
				"tenant_id":            wantFilter.TenantId,
				"hr_approver":          wantFilter.HrApprover,
				"hr_approver_decision": wantUpdate.HrApproverDecision,
				"recruiter":            wantUpdate.Recruiter,
			},
		)

		s.dbRootConn.Exec("UPDATE job_requisition SET supervisor_decision = 'PENDING' WHERE id = $1", test.input.Id)
	}
}

func (s *IntegrationTestSuite) TestUpdateJobRequisitionShouldCheckForHrApproval() {
	tests := []struct {
		name  string
		input storage.JobRequisition
	}{
		{
			"Should fail because setting the recruiter requires Hr approval",
			storage.JobRequisition{
				Id:                 s.defaultJobRequisition.Id,
				TenantId:           s.defaultJobRequisition.TenantId,
				HrApprover:         s.defaultJobRequisition.HrApprover,
				HrApproverDecision: "REJECTED",
				Recruiter:          s.defaultRecruiter.Id,
			},
		},
		{
			"Should fail because setting the filled_by requires Hr approval",
			storage.JobRequisition{
				Id:                 s.defaultJobRequisition.Id,
				TenantId:           s.defaultJobRequisition.TenantId,
				HrApprover:         s.defaultJobRequisition.HrApprover,
				HrApproverDecision: "REJECTED",
				FilledBy:           s.defaultUser.Id,
			},
		},
		{
			"Should fail because setting the filled_at requires Hr approval",
			storage.JobRequisition{
				Id:                 s.defaultJobRequisition.Id,
				TenantId:           s.defaultJobRequisition.TenantId,
				HrApprover:         s.defaultJobRequisition.HrApprover,
				HrApproverDecision: "REJECTED",
				FilledAt:           time.Now(),
			},
		},
	}

	for _, test := range tests {
		wantFilter := storage.JobRequisition{
			Id:         test.input.Id,
			TenantId:   test.input.TenantId,
			HrApprover: test.input.HrApprover,
		}

		wantUpdate := storage.JobRequisition{
			HrApproverDecision: test.input.HrApproverDecision,
			Recruiter:          test.input.Recruiter,
			FilledBy:           test.input.FilledBy,
			FilledAt:           test.input.FilledAt,
		}

		_, err := s.dbRootConn.Exec("UPDATE job_requisition SET supervisor_decision = 'APPROVED' WHERE id = $1", wantFilter.Id)
		if err != nil {
			log.Fatalf("Could not set supervisor approval: %s", err)
		}

		err = s.postgres.UpdateJobRequisition(wantUpdate, wantFilter)
		s.expectErrorCode(err, "MISSING-HR-APPROVAL-ERROR")

		s.expectSelectQueryToReturnNoRows(
			"job_requisition",
			map[string]any{
				"id":                   wantFilter.Id,
				"tenant_id":            wantFilter.TenantId,
				"hr_approver":          wantFilter.HrApprover,
				"hr_approver_decision": wantUpdate.HrApproverDecision,
				"recruiter":            sql.NullString{String: wantUpdate.Recruiter, Valid: wantUpdate.Recruiter != ""},
				"filled_by":            sql.NullString{String: wantUpdate.FilledBy, Valid: wantUpdate.FilledBy != ""},
				"filled_at":            sql.NullTime{Time: wantUpdate.FilledAt, Valid: !wantUpdate.FilledAt.IsZero()},
			},
		)
	}
}
