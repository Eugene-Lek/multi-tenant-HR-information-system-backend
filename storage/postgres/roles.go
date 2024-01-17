package postgres

import (
	"fmt"
	"strings"

	"multi-tenant-HR-information-system-backend/httperror"
	"multi-tenant-HR-information-system-backend/storage"

	"github.com/lib/pq"
)

func (postgres *postgresStorage) CreatePolicies(policies storage.Policies) error {
	// All inserts must be done in a single query so that a violation of 1 unique constraint will roll back all inserts

	identifiers := []string{}
	values := []any{}

	for i, resource := range policies.Resources {
		values = append(values, policies.Role, policies.TenantId, resource.Path, resource.Method)
		identifiers = append(identifiers, fmt.Sprintf("('p', $%v, $%v, $%v, $%v)", i*4+1, i*4+2, i*4+3, i*4+4))
	}

	query := "INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES " + strings.Join(identifiers, ", ")
	_, err := postgres.db.Exec(query, values...)
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
