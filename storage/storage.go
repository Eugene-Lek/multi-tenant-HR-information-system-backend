package storage

import (
	"io"
	"time"
)

type Storage interface {
	CreateTenant(tenant Tenant) error
	GetTenants(filter Tenant) ([]Tenant, error)
	CreateDivision(division Division) error
	CreateDepartment(department Department) error

	CreateUser(user User) error
	GetUsers(userFilter User) ([]User, error)
	GetUserSupervisors(userId string, TenantId string) ([]string, error)

	CreatePosition(position Position) error

	CreatePositionAssignment(positionAssignment PositionAssignment) error

	GetUserPositions(userId string, filter UserPosition) ([]UserPosition, error)

	CreatePolicies(policies Policies) error
	CreateRoleAssignment(roleAssignment RoleAssignment) error

	CreateJobRequisition(jobRequisition JobRequisition) error
	GetJobRequisitions(filter JobRequisition) ([]JobRequisition, error)
	UpdateJobRequisition(newValues JobRequisition, filter JobRequisition) error
	HrApproveJobRequisition(jobRequisitionId string, tenantId string, hrApprover string, recruiter string) error

	CreateJobApplication(jobApplication JobApplication) error
	GetJobApplications(filter JobApplication) ([]JobApplication, error)
	UpdateJobApplication(newValues JobApplication, filter JobApplication) error
	OnboardNewHire(jobApplication JobApplication, newUser User) error
}

type FileStorage interface {
	UploadResume(file io.Reader, jobApplicationId string, firstName string, lastName string, fileExt string) (url string, err error)
}

type Tenant struct {
	Id        string
	Name      string
	CreatedAt string
	UpdatedAt string
}

type Division struct {
	Id        string
	TenantId  string
	Name      string
	CreatedAt string
	UpdatedAt string
}

type Department struct {
	Id         string
	TenantId   string
	DivisionId string
	Name       string
	CreatedAt  string
	UpdatedAt  string
}

type User struct {
	Id            string
	TenantId      string
	Email         string
	Password      string
	TotpSecretKey string
	CreatedAt     string
	UpdatedAt     string
	LastLogin     string
}

type Position struct {
	Id                    string
	TenantId              string
	Title                 string
	DepartmentId          string
	SupervisorPositionIds []string
	CreatedAt             string
	UpdatedAt             string
}

type PositionAssignment struct {
	TenantId   string
	PositionId string
	UserId     string
	StartDate  string
	EndDate    string
	CreatedAt  string
	UpdatedAt  string
}

type UserPosition struct {
	Id                    string
	TenantId              string
	Title                 string
	DepartmentId          string
	SupervisorPositionIds []string
	StartDate             string
	EndDate               string
}

type JobRequisition struct {
	Id                    string
	TenantId              string
	PositionId            string
	Title                 string
	DepartmentId          string
	SupervisorPositionIds []string
	JobDescription        string
	JobRequirements       string
	Requestor             string
	Supervisor            string
	SupervisorDecision    string
	HrApprover            string
	HrApproverDecision    string
	Recruiter             string
	FilledBy              string
	FilledAt              time.Time
	CreatedAt             string
	UpdatedAt             string
}

type JobApplication struct {
	Id                    string
	TenantId              string
	JobRequisitionId      string
	FirstName             string
	LastName              string
	CountryCode           string
	PhoneNumber           string
	Email                 string
	ResumeS3Url           string
	RecruiterDecision     string
	InterviewDate         string
	HiringManagerDecision string
	OfferStartDate string
	OfferEndDate string
	ApplicantDecision     string
	CreatedAt             string
	UpdatedAt             string
}

type Resource struct {
	Path   string
	Method string
}

type Policies struct {
	Subject   string // Can either be userId or Role. If it is a role, users will have to be assigned to the role to gain access to the policy
	TenantId  string
	Resources []Resource
	CreatedAt string
	UpdatedAt string
}

type RoleAssignment struct {
	UserId    string `validate:"required,notBlank,uuid" name:"user id"`
	Role      string `validate:"required,notBlank" name:"role name"`
	TenantId  string `validate:"required,notBlank,uuid" name:"tenant id"`
	CreatedAt string
	UpdatedAt string
}
