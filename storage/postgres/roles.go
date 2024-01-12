package postgres

import (
	"multi-tenant-HR-information-system-backend/httperror"
	"multi-tenant-HR-information-system-backend/storage"

	"github.com/lib/pq"
)

func (postgres *postgresStorage) CreatePolicies(policies storage.Policies) error {
	for _, resource := range policies.Resources {
		query := "INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', $1, $2, $3, $4)"
		_, err := postgres.db.Exec(query, policies.Role, policies.TenantId, resource.Path, resource.Method)
		if err != nil {
			if pgErr, ok := err.(*pq.Error); ok {
				if pgErr.Code == "23505" {
					return NewUniqueViolationError("policy", pgErr)
				} else {
					return httperror.NewInternalServerError(pgErr)
				}
			}
			return httperror.NewInternalServerError(err)
		}
	}

	return nil	
}

func (postgres *postgresStorage) CreateRoleAssignment(roleAssignment storage.RoleAssignment) error {
	query := "INSERT INTO casbin_rule (Ptype, V0, V1, V2) VALUES ('g', $1, $2, $3)"
	_, err := postgres.db.Exec(query, roleAssignment.UserId, roleAssignment.Role, roleAssignment.TenantId)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" {
				return NewUniqueViolationError("role assignment", pgErr)
			} else {
				return httperror.NewInternalServerError(pgErr)
			}
		}
		return httperror.NewInternalServerError(err)
	}

	return nil
}