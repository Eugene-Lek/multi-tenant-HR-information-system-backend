package storage

type Storage interface {
	CreateTenant(tenant Tenant) error
	CreateDivision(division Division) error
	CreateDepartment(department Department) error
	CreateUser(user User) error
	GetUsers(userFilter User) ([]User, error)
	CreateAppointment(appointment Appointment) error
	CreatePolicies(policies Policies) error
	CreateRoleAssignment(roleAssignment RoleAssignment) error	
}

type Tenant struct {
	Id        string `validate:"required,notBlank,uuid" name:"tenant id"`
	Name      string `validate:"required,notBlank" name:"tenant name"`
	CreatedAt string
	UpdatedAt string
}

type Division struct {
	Id        string `validate:"required,notBlank,uuid" name:"division id"`
	TenantId  string `validate:"required,notBlank,uuid" name:"tenant id"`
	Name      string `validate:"required,notBlank" name:"division name"`
	CreatedAt string
	UpdatedAt string
}

type Department struct {
	Id         string `validate:"required,notBlank,uuid" name:"department id"`
	TenantId  string `validate:"required,notBlank,uuid" name:"tenant id"`	
	DivisionId string `validate:"required,notBlank,uuid" name:"division id"`
	Name       string `validate:"required,notBlank" name:"department name"`
	CreatedAt  string
	UpdatedAt  string
}

type User struct {
	Id            string `validate:"required,notBlank,uuid" name:"user id"`
	TenantId      string `validate:"required,notBlank,uuid" name:"tenant id"`	
	Email         string `validate:"required,notBlank,email" name:"user email"`
	Password      string
	TotpSecretKey string
	CreatedAt     string
	UpdatedAt     string
	LastLogin     string
}

type Appointment struct {
	Id           string `validate:"required,notBlank,uuid" name:"appointment id"`
	TenantId      string `validate:"required,notBlank,uuid" name:"tenant id"`		
	Title        string `validate:"required,notBlank" name:"appointment title"`	
	DepartmentId string `validate:"required,notBlank,uuid" name:"department id"`
	UserId       string `validate:"required,notBlank,uuid" name:"user id"`
	StartDate    string `validate:"required,notBlank,isIsoDate" name:"start date"`
	EndDate      string `validate:"omitempty,notBlank,isIsoDate,validAppointmentDuration" name:"end date"`
	CreatedAt    string
	UpdatedAt    string
}

type Resource struct {
	Path string `validate:"required,notBlank" name:"resource path"`
	Method string `validate:"required,notBlank,oneof=POST GET PUT DELETE" name:"resource method"`	
}

type Policies struct {
	Role string `validate:"required,notBlank" name:"role name"`
	TenantId string `validate:"required,notBlank,uuid" name:"tenant id"`
	Resources []Resource `validate:"dive"`
}

type RoleAssignment struct {
	UserId string `validate:"required,notBlank,uuid" name:"user id"`
	Role string `validate:"required,notBlank" name:"role name"`
	TenantId string `validate:"required,notBlank,uuid" name:"tenant id"`
}