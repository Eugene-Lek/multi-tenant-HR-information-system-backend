package storage

import "time"

type Storage interface {
	CreateTenant(tenant Tenant) error
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
	Id                 string
	TenantId           string
	Title              string
	DepartmentId       string
	JobDescription     string
	JobRequirements    string
	Requestor          string
	Supervisor         string
	SupervisorDecision string
	HrApprover         string
	HrApproverDecision string
	Recruiter          string
	FilledBy           string
	FilledAt           time.Time
	CreatedAt          string
	UpdatedAt          string
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
